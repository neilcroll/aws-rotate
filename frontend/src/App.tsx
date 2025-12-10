import { useEffect, useMemo, useState } from 'react';
import type { FormEvent } from 'react';
import { ListProfiles, RotateProfiles } from '../wailsjs/go/desktop/App';
import { desktop } from '../wailsjs/go/models';
import type { rotator } from '../wailsjs/go/models';
import './App.css';

type ProfileRow = {
  name: string;
  isChina: boolean;
  checked: boolean;
  status?: 'idle' | 'pending' | 'success' | 'error';
  message?: string;
};

type Mode = 'concurrent' | 'sequential';

function App() {
  const [profiles, setProfiles] = useState<ProfileRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [mode, setMode] = useState<Mode>('concurrent');
  const [submitting, setSubmitting] = useState(false);
  const [progressMessage, setProgressMessage] = useState<string | null>(null);
  const [rotationResults, setRotationResults] = useState<rotator.ProfileResult[] | null>(null);
  const [catastrophicError, setCatastrophicError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    ListProfiles()
      .then((data) => {
        if (cancelled) {
          return;
        }
        if (data.length === 0) {
          setError('No AWS profiles were found in ~/.aws/credentials.');
        }
        setProfiles(
          data.map((profile) => ({
            name: profile.name,
            isChina: profile.isChina ?? false,
            checked: false,
            status: 'idle',
            message: '',
          }))
        );
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err?.message ?? 'Failed to load AWS profiles.');
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const selectedCount = useMemo(
    () => profiles.filter((profile) => profile.checked).length,
    [profiles]
  );

  const passwordsMatch = useMemo(() => {
    if (password.length === 0 && confirmPassword.length === 0) {
      return null; // No indicator when both empty
    }
    if (confirmPassword.length === 0) {
      return null; // No indicator until confirm password is entered
    }
    return password === confirmPassword;
  }, [password, confirmPassword]);

  const canSubmit =
    selectedCount > 0 &&
    username.trim().length > 0 &&
    password.length > 0 &&
    password === confirmPassword &&
    !submitting;

  const toggleRow = (name: string, key: 'checked' | 'isChina') => {
    setProfiles((prev) =>
      prev.map((profile) =>
        profile.name === name
          ? {
              ...profile,
              [key]: !profile[key],
              status: key === 'checked' ? 'idle' : profile.status,
            }
          : profile
      )
    );
    setRotationResults(null);
    setCatastrophicError(null);
  };

  const mapResults = (results: rotator.ProfileResult[]) => {
    setProfiles((prev) =>
      prev.map((profile) => {
        const match = results.find((result) => result.name === profile.name);
        if (!match) {
          return profile;
        }
        return {
          ...profile,
          status: match.success ? 'success' : 'error',
          message: match.message,
        };
      })
    );
  };

  const handleRotate = async (event: FormEvent) => {
    event.preventDefault();
    if (!canSubmit) {
      return;
    }
    setSubmitting(true);
    setCatastrophicError(null);
    setRotationResults(null);
    setProgressMessage('Starting credential rotation...');

    const payload = new desktop.RotationRequest({
      profiles: profiles.map(({ name, isChina, checked }) => ({
        name,
        isChina,
        checked,
      })),
      username: username.trim(),
      password,
      mode,
    });

    // Reset all selected profile statuses to pending
    setProfiles((prev) =>
      prev.map((profile) =>
        profile.checked ? { ...profile, status: 'pending' as const, message: 'Processing...' } : profile
      )
    );

    try {
      setProgressMessage(`Rotating credentials for ${selectedCount} profile(s)...`);
      const results = await RotateProfiles(payload);
      
      setProgressMessage(null);
      
      // Process results - update both table and results panel
      if (results && results.length > 0) {
        mapResults(results);
        setRotationResults(results);
      } else {
        // Unexpected: no results returned
        setCatastrophicError('No results returned from rotation operation. This should not happen - please check the application logs.');
      }
    } catch (err: unknown) {
      // Catastrophic failure before processing could start (backup failed, profiles not found, etc.)
      setProgressMessage(null);
      
      let errorMessage = 'A critical error occurred before rotation could begin.';
      const errorDetails: string[] = [];
      
      if (err instanceof Error) {
        errorMessage = err.message || errorMessage;
        
        // Provide helpful context for common pre-processing errors
        if (errorMessage.includes('backup')) {
          errorDetails.push('• Check that ~/.aws/credentials exists and is writable');
          errorDetails.push('• Ensure you have sufficient disk space');
        } else if (errorMessage.includes('profile') && errorMessage.includes('not found')) {
          errorDetails.push('• The credentials file may have been modified externally');
          errorDetails.push('• Try refreshing your browser or restarting the application');
        } else if (errorMessage.includes('load') && errorMessage.includes('profiles')) {
          errorDetails.push('• Check that ~/.aws/credentials exists and is readable');
          errorDetails.push('• Verify the file format is correct');
        }
        
        if (err.stack && import.meta.env.DEV) {
          errorDetails.push(`\nStack trace:\n${err.stack}`);
        }
      } else if (typeof err === 'string') {
        errorMessage = err;
      }
      
      const fullError = errorDetails.length > 0
        ? `${errorMessage}\n\n${errorDetails.join('\n')}`
        : errorMessage;
      
      setCatastrophicError(fullError);
      
      // Reset pending statuses since processing never started
      setProfiles((prev) =>
        prev.map((profile) =>
          profile.status === 'pending' ? { ...profile, status: 'idle' as const, message: '' } : profile
        )
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="app-shell">
      <header>
        <h1>AWS Multi-Account Credentials Rotator</h1>
        <p>Select the accounts you want to update and confirm the new password.</p>
      </header>

      {loading && <div className="panel">Loading profiles…</div>}
      
      {/* Show errors but don't hide content unless it's a load error */}
      {!loading && error && profiles.length === 0 && (
        <div className="alert alert-danger">{error}</div>
      )}

      {!loading && profiles.length > 0 && (
        <>
          <section className="panel">
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th className="checkbox-cell">
                      <input
                        type="checkbox"
                        aria-label="Select all profiles"
                        checked={selectedCount > 0 && selectedCount === profiles.length}
                        onChange={(event) => {
                          const next = event.target.checked;
                          setProfiles((prev) =>
                            prev.map((profile) => ({
                              ...profile,
                              checked: next,
                              status: 'idle',
                            }))
                          );
                        }}
                      />
                    </th>
                    <th>Profile</th>
                    <th className="checkbox-cell">AWS China</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {profiles.map((profile) => (
                    <tr key={profile.name}>
                      <td className="checkbox-cell">
                        <input
                          type="checkbox"
                          aria-label={`Select ${profile.name}`}
                          checked={profile.checked}
                          onChange={() => toggleRow(profile.name, 'checked')}
                        />
                      </td>
                      <td>{profile.name}</td>
                      <td className="checkbox-cell">
                        <input
                          type="checkbox"
                          aria-label={`Mark ${profile.name} as AWS China`}
                          checked={profile.isChina}
                          onChange={() => toggleRow(profile.name, 'isChina')}
                        />
                      </td>
                      <td>
                        {profile.status === 'pending' && (
                          <span className="status pending">
                            <span className="spinner-small"></span>
                            Processing...
                          </span>
                        )}
                        {profile.status === 'success' && (
                          <span className="status success">✓ Success</span>
                        )}
                        {profile.status === 'error' && (
                          <span className="status danger" title={profile.message}>
                            ✗ {profile.message || 'Failed'}
                          </span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <p className="hint">
              Tick the final column for accounts that live inside the AWS China partition so the
              correct region and endpoints are used.
            </p>
          </section>

          {selectedCount > 0 && (
            <form className="panel" onSubmit={handleRotate}>
              <h2>Rotate selected profiles ({selectedCount})</h2>
              
              <div className="form-layout">
                <label className="form-field">
                  <span>IAM username</span>
                  <input
                    type="text"
                    value={username}
                    onChange={(event) => setUsername(event.target.value)}
                    placeholder="e.g. user.name"
                  />
                </label>

                <label className="form-field">
                  <span>Execution mode</span>
                  <select value={mode} onChange={(event) => setMode(event.target.value as Mode)}>
                    <option value="concurrent">Concurrent (faster)</option>
                    <option value="sequential">Sequential (safer)</option>
                  </select>
                </label>

                <label className="form-field password-field">
                  <span>New password</span>
                  <div className="password-input-wrapper">
                    <input
                      type="password"
                      value={password}
                      onChange={(event) => setPassword(event.target.value)}
                    />
                    {passwordsMatch !== null && (
                      <span className={`password-indicator ${passwordsMatch ? 'match' : 'no-match'}`}>
                        {passwordsMatch ? '✓' : '✗'}
                      </span>
                    )}
                  </div>
                </label>

                <label className="form-field password-field">
                  <span>Confirm password</span>
                  <div className="password-input-wrapper">
                    <input
                      type="password"
                      value={confirmPassword}
                      onChange={(event) => setConfirmPassword(event.target.value)}
                    />
                    {passwordsMatch !== null && (
                      <span className={`password-indicator ${passwordsMatch ? 'match' : 'no-match'}`}>
                        {passwordsMatch ? '✓' : '✗'}
                      </span>
                    )}
                  </div>
                </label>
              </div>

              {progressMessage && (
                <div className="progress-message">
                  <span className="spinner"></span>
                  {progressMessage}
                </div>
              )}

              <div className="form-actions">
                <button type="submit" disabled={!canSubmit}>
                  {submitting ? 'Rotating…' : 'Rotate credentials'}
                </button>
                {!canSubmit && !submitting && (
                  <span className="hint">
                    Enter the username, matching passwords, and keep at least one profile ticked.
                  </span>
                )}
              </div>
            </form>
          )}

          {/* Show catastrophic errors or results in a dedicated component below the form */}
          {(catastrophicError || (rotationResults && rotationResults.length > 0)) && (
            <section className="panel results-panel">
              <h2>Rotation Results</h2>
              
              {catastrophicError && (
                <div className="catastrophic-error">
                  <div className="error-header">
                    <span className="error-icon">⚠️</span>
                    <span className="error-title">Critical Error</span>
                  </div>
                  <div className="error-message">{catastrophicError}</div>
                </div>
              )}

              {rotationResults && rotationResults.length > 0 && (
                <>
                  <div className="results-summary">
                <div className="summary-stats">
                  <div className="stat-item success">
                    <span className="stat-icon">✓</span>
                    <div className="stat-content">
                      <div className="stat-number">
                        {rotationResults.filter((r) => r.success).length}
                      </div>
                      <div className="stat-label">Successful</div>
                    </div>
                  </div>
                  <div className="stat-item failed">
                    <span className="stat-icon">✗</span>
                    <div className="stat-content">
                      <div className="stat-number">
                        {rotationResults.filter((r) => !r.success).length}
                      </div>
                      <div className="stat-label">Failed</div>
                    </div>
                  </div>
                  <div className="stat-item total">
                    <span className="stat-icon">≡</span>
                    <div className="stat-content">
                      <div className="stat-number">{rotationResults.length}</div>
                      <div className="stat-label">Total</div>
                    </div>
                  </div>
                </div>
              </div>

              <div className="results-details">
                {rotationResults.map((result) => (
                  <div
                    key={result.name}
                    className={`result-item ${result.success ? 'result-success' : 'result-error'}`}
                  >
                    <div className="result-header">
                      <span className="result-icon">
                        {result.success ? '✓' : '✗'}
                      </span>
                      <span className="result-name">{result.name}</span>
                      <span className={`result-badge ${result.success ? 'badge-success' : 'badge-error'}`}>
                        {result.success ? 'Success' : 'Failed'}
                      </span>
                    </div>
                    <div className="result-message">{result.message}</div>
                  </div>
                ))}
              </div>
                </>
              )}
            </section>
          )}
        </>
      )}
    </div>
  );
}

export default App;
