# GitHub Actions Workflows

This directory contains the CI/CD workflows for the AWS Multi-Account Credentials Rotator project.

## Workflows

### build-and-test.yml
**Purpose**: Main CI/CD pipeline for building, testing, and releasing the application.

**Triggers**:
- Push to `main`, `master`, or `develop` branches
- Pull requests to `main`, `master`, or `develop` branches
- Version tags (e.g., `v0.3.0`)

**Jobs**:

1. **test**: Runs on Ubuntu
   - Sets up Go 1.21
   - Downloads and verifies dependencies
   - Runs tests with race detection and coverage
   - Runs `go vet` for static analysis
   - Runs `staticcheck` for advanced linting

2. **build**: Runs after tests pass, builds for multiple platforms
   - Linux AMD64
   - Linux ARM64
   - macOS Intel (AMD64)
   - macOS Apple Silicon (ARM64)
   - Windows AMD64
   - Uploads build artifacts with 30-day retention

3. **release**: Creates GitHub releases (only for version tags)
   - Downloads all build artifacts
   - Creates compressed archives (.tar.gz for Unix, .zip for Windows)
   - Generates changelog from git commits
   - Creates GitHub release with all binaries attached

### lint.yml
**Purpose**: Code quality and formatting checks.

**Triggers**:
- Push to `main`, `master`, or `develop` branches
- Pull requests to `main`, `master`, or `develop` branches

**Jobs**:

1. **golangci-lint**: Comprehensive linting
   - Runs golangci-lint with configuration from `.golangci.yml`
   - Checks for code quality issues, bugs, and style violations

2. **go-fmt**: Code formatting
   - Verifies all code is properly formatted with `gofmt`
   - Fails if any files need formatting

## Creating a Release

To create a new release:

1. Ensure all changes are committed and pushed
2. Create and push a version tag:
   ```bash
   git tag -a v0.3.0 -m "Release v0.3.0 - Description of changes"
   git push origin v0.3.0
   ```
3. GitHub Actions will automatically:
   - Run all tests
   - Build binaries for all platforms
   - Create release archives
   - Generate changelog
   - Publish GitHub release with downloadable binaries

## Monitoring Workflows

- View workflow runs: [Actions tab](https://github.com/neilcroll/aws-rotate/actions)
- Check build status badges can be added to main README.md
- Artifacts are retained for 30 days
- Failed workflows send notifications to repository maintainers

## Local Testing

Before pushing, test locally to avoid CI failures:

```bash
# Run the same checks CI will run
make check

# Build for all platforms
make build-all
```

## Workflow Permissions

The workflows require:
- `contents: write` - For creating releases and uploading artifacts
- Default `GITHUB_TOKEN` - Automatically provided by GitHub Actions

