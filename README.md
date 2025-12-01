# AWS Multi-Account Credentials Rotator

A Go-based CLI tool to automate the rotation of IAM user credentials (console passwords and access keys) across multiple AWS accounts. Born from the frustration of manually updating passwords across 10-15 AWS accounts every quarter due to corporate password policies.

## What It Does

This tool automates the tedious process of rotating IAM user credentials across multiple AWS accounts by:

1. Reading your existing AWS profiles from `~/.aws/credentials`
2. Letting you select which profiles (accounts) to update
3. Using your current access keys to authenticate to each account
4. Updating the IAM user's console password
5. Creating new AWS access keys
6. Updating `~/.aws/credentials` with the new keys
7. Activating the new keys and deleting the old ones
8. Moving on to the next account

**Key Assumption**: You use the same IAM username across all AWS accounts (common in multi-account enterprise setups).

## Features

- **Interactive Profile Selection**: Choose specific profiles, all profiles, or all except default
- **Automated Backup**: Creates a timestamped backup of `~/.aws/credentials` before making changes
- **Concurrent or Sequential Updates**: Update multiple accounts simultaneously or one at a time
- **Validation**: Verifies new credentials work before deleting old keys
- **Clear Status Reporting**: Detailed feedback on each profile's update status
- **Rollback Support**: Backup file allows manual restoration if needed
- **Desktop GUI (Wails + React)**: Visual profile table with selection checkboxes, an AWS China toggle per profile, inline password confirmation, and status badges for each rotation attempt

## Prerequisites

- **Go 1.22 or higher** - [Download](https://golang.org/dl/)
- **Node.js 20+** - Needed for the React frontend
- **Wails CLI** - `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **AWS CLI** - [Install Guide](https://aws.amazon.com/cli/)
- **Existing AWS profiles** configured in `~/.aws/credentials`
- **IAM Permissions**: Your IAM user must have:
  - `iam:UpdateLoginProfile` (to change console password)
  - `iam:CreateAccessKey` (to create new access keys)
  - `iam:ListAccessKeys` (to list existing keys)
  - `iam:DeleteAccessKey` (to remove old keys)
  - `sts:GetCallerIdentity` (for session validation)

## Installation

### From Source

```bash
git clone https://github.com/neilcroll/aws-rotate.git
cd aws-rotate
go mod tidy
go build
```

This creates an executable:
- `aws-rotate` on Linux/macOS
- `aws-rotate.exe` on Windows

### Optional: Install Globally

```bash
# Linux/macOS
sudo mv aws-rotate /usr/local/bin/

# Windows (PowerShell as Administrator)
Move-Item aws-rotate.exe C:\Windows\System32\
```

## Usage

### Desktop GUI (Wails + React)

Issues Version 2 introduces a desktop experience that mirrors the CLI workflow but adds a profile table, AWS China toggles, and inline validation.

```bash
# One-time setup
npm install --prefix frontend

# For hot reload during development
/home/ubuntu/go/bin/wails dev   # or wails dev if it's already on your PATH

# To run the packaged app without the dev server
npm run build --prefix frontend
go run .
```

`go run .` (or `wails build`) produces a desktop binary that ships the compiled React assets via `go:embed`.

### Command-Line Workflow

The original terminal workflow still exists under `cmd/cli`.

```bash
# Default concurrent mode
go run ./cmd/cli

# Sequential mode (safer if you regularly hit rate limits)
go run ./cmd/cli -s
```

Flags:

- `-c` / `--concurrent` (default): Update profiles concurrently
- `-s` / `--sequential`: Update profiles one at a time

## How It Works

### Step-by-Step Workflow

1. **Profile Selection**
   - Tool reads all profiles from `~/.aws/credentials`
   - You select which accounts to update via interactive menu
   - Options: specific profiles, all profiles, or all except default

2. **Credential Input**
   - Enter your IAM username (same across all accounts)
   - Enter new console password (must meet AWS password policy requirements)

3. **Backup Creation**
   - Tool creates `~/.aws/credentials.backup.YYYYMMDD-HHMMSS`
   - Original credentials preserved in case rollback is needed

4. **Account Updates** (per profile, concurrent or sequential)
   - Authenticates using existing access key from profile
   - Updates IAM user console password
   - Creates new access key pair
   - Updates `~/.aws/credentials` with new keys
   - Validates new credentials work via `sts:GetCallerIdentity`
   - Activates new keys and deletes old access keys
   - Reports success or failure

5. **Completion**
   - Summary of all profile updates
   - List of any failures for manual remediation

### Example Session

```
$ ./aws-rotate

AWS Multi-Account Credentials Rotator
======================================

Found profiles:
  1. dev-account
  2. staging-account  
  3. prod-account
  4. default

Select profiles to update:
  [ ] All profiles
  [X] All except default
  [ ] default
  [ ] dev-account
  [ ] staging-account
  [ ] prod-account


Selected profiles: dev-account, staging-account, prod-account

Enter IAM username: user.name
Enter new password: ********

Creating backup: ~/.aws/credentials.backup.20251124-143022

Updating profiles concurrently...
  ✓ dev-account: Credentials updated successfully
  ✓ staging-account: Credentials updated successfully  
  ✓ prod-account: Credentials updated successfully

All profiles updated successfully!
```

## Troubleshooting

### "Access Denied" Errors

**Problem**: IAM user lacks required permissions.

**Solution**: Ensure your IAM user has the permissions listed in Prerequisites. You may need to request these from your AWS administrator.

### "InvalidPasswordException"

**Problem**: New password doesn't meet AWS password policy requirements.

**Solution**: Check your account's password policy (AWS Console → IAM → Account Settings). Password must typically:
- Be 12-128 characters long
- Contain uppercase, lowercase, numbers, and symbols
- Not match previous passwords (if password history is enabled)

### Rate Limiting / Throttling

**Problem**: Too many concurrent API calls causing throttling.

**Solution**: Use sequential mode (`-s` flag) to reduce API call rate.

### Failed Mid-Update

**Problem**: Tool failed partway through updating multiple accounts.

**Solution**: 
1. Check which accounts failed from the summary
2. Restore backup if needed: `cp ~/.aws/credentials.backup.YYYYMMDD-HHMMSS ~/.aws/credentials`
3. Re-run tool with only failed profiles selected
4. If credentials are mixed (some old, some new), manually fix via AWS Console

### AWS CLI Not Configured

**Problem**: No `~/.aws/credentials` file exists.

**Solution**: Run `aws configure` first to set up at least one profile, or manually create the credentials file.

## Security Considerations

- **Backup Files**: Contain sensitive access keys. Delete old backups regularly or store securely.
- **Passwords**: Tool prompts for passwords securely (no echo). Ensure your terminal session is secure.
- **API Keys**: Old keys are deleted after new ones are validated, minimizing exposure window.
- **Concurrent Mode**: Uses goroutines; keys briefly exist in memory. Sequential mode is more conservative.

## Limitations

- Only works with IAM user credentials (not SSO or federated access)
- Requires same IAM username across all accounts
- Cannot rotate credentials for users with only one access key (AWS allows max 2)
- Does not update AWS CLI config file (`~/.aws/config`), only credentials file
- MFA not supported for API access (MFA is for console login only)

## Development

### Building Locally

Use the provided Makefile for common development tasks:

```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Run linters and code quality checks
make lint

# Format code
make fmt

# Run all checks (lint + test)
make check

# Clean build artifacts
make clean

# Install to system
make install

# See all available targets
make help
```

### Manual Build

```bash
# For current platform
go build -o aws-rotate

# For specific platform
GOOS=linux GOARCH=amd64 go build -o aws-rotate-linux-amd64
GOOS=darwin GOARCH=arm64 go build -o aws-rotate-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o aws-rotate.exe
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

The project uses several tools to maintain code quality:

- **golangci-lint**: Comprehensive linter aggregator
- **go vet**: Go's built-in static analyzer
- **go fmt**: Code formatter
- **staticcheck**: Advanced Go linter

Install golangci-lint:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### CI/CD

The project uses GitHub Actions for continuous integration and deployment:

#### Build and Test Workflow
- Triggers on push to main branches and pull requests
- Runs tests with race detection and coverage
- Builds binaries for:
  - Linux (AMD64, ARM64)
  - macOS (Intel, Apple Silicon)
  - Windows (AMD64)
- Uploads build artifacts
- Creates GitHub releases on version tags

#### Lint Workflow
- Runs golangci-lint with comprehensive rule set
- Checks code formatting with go fmt
- Ensures code quality standards

### Creating a Release

1. Update version in code if needed
2. Create and push a version tag:
   ```bash
   git tag -a v0.3.0 -m "Release v0.3.0"
   git push origin v0.3.0
   ```
3. GitHub Actions automatically:
   - Builds binaries for all platforms
   - Creates compressed archives
   - Generates changelog
   - Creates GitHub release with binaries attached

### Pre-commit Checklist

Before committing code, ensure:
```bash
# Format code
make fmt

# Run all checks
make check

# Build for all platforms
make build-all
```

All checks should pass before pushing to avoid CI failures.

## Contributing

Contributions welcome! If you have ideas, suggestions, or bug reports:
1. Open an issue describing the problem or enhancement
2. Fork the repository
3. Create a feature branch
4. Submit a pull request

## License

This project is licensed under the MIT License. See LICENSE file for details.

## Author

Created by Neil Croll to solve a recurring problem with corporate AWS password policies.
