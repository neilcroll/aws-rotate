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
  const [resultsNote, setResultsNote] = useState<string | null>(null);

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
    setResultsNote(null);
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
    setError(null);
    setResultsNote(null);

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

    try {
      const results = await RotateProfiles(payload);
      mapResults(results);
      const hasFailures = results.some((result) => !result.success);
      setResultsNote(
        hasFailures
          ? 'Rotation finished, but at least one profile needs attention.'
          : 'Rotation finished successfully.'
      );
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'Rotation failed unexpectedly.';
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="app-shell">
      <header>
        <div>
          <h1>AWS Multi-Account Credentials Rotator</h1>
          <p>Select the accounts you want to update and confirm the new password.</p>
        </div>
        <span className="version-pill">Issues v2</span>
      </header>

      {loading && <div className="panel">Loading profiles…</div>}
      {!loading && error && <div className="alert alert-danger">{error}</div>}

      {!loading && !error && (
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
                        {profile.status === 'success' && (
                          <span className="status success">Success</span>
                        )}
                        {profile.status === 'error' && (
                          <span className="status danger">{profile.message}</span>
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
              <div className="form-grid">
                <label>
                  <span>IAM username</span>
                  <input
                    type="text"
                    value={username}
                    onChange={(event) => setUsername(event.target.value)}
                    placeholder="e.g. neil.croll"
                  />
                </label>

                <label>
                  <span>Execution mode</span>
                  <select value={mode} onChange={(event) => setMode(event.target.value as Mode)}>
                    <option value="concurrent">Concurrent (faster)</option>
                    <option value="sequential">Sequential (safer)</option>
                  </select>
                </label>

                <label>
                  <span>New password</span>
                  <input
                    type="password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                  />
                </label>

                <label>
                  <span>Confirm password</span>
                  <input
                    type="password"
                    value={confirmPassword}
                    onChange={(event) => setConfirmPassword(event.target.value)}
                  />
                </label>
              </div>

              <div className="form-actions">
                <button type="submit" disabled={!canSubmit}>
                  {submitting ? 'Rotating…' : 'Rotate credentials'}
                </button>
                {!canSubmit && (
                  <span className="hint">
                    Enter the username, matching passwords, and keep at least one profile ticked.
                  </span>
                )}
              </div>
              {resultsNote && <div className="alert alert-info">{resultsNote}</div>}
            </form>
          )}
        </>
      )}
    </div>
  );
}

export default App;
