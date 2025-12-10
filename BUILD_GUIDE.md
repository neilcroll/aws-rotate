# Build Guide

This project contains two applications:
1. **Desktop GUI** (Wails + React) - Main application
2. **CLI** (Go only) - Command-line interface

## Prerequisites

- Go 1.22+
- Node.js 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Desktop GUI (Wails)

### Development
```bash
# Hot reload development
make dev
# or
wails dev
```

### Production Build
```bash
# Build for current platform
make build-desktop
# or
wails build

# Binary location: build/bin/aws-rotate
```

### Install Globally
```bash
make install
# or
sudo cp build/bin/aws-rotate /usr/local/bin/
```

## CLI Only

### Development
```bash
# Build CLI
make build-cli

# Run CLI
make run-cli
```

### Production Build
```bash
# Build for current platform
make build-cli

# Build for all platforms
make build-all

# Binary location: cmd/cli/aws-rotate-cli
```

### Install Globally
```bash
make install-cli
# or
sudo cp cmd/cli/aws-rotate-cli /usr/local/bin/aws-rotate
```

## Common Tasks

```bash
# Run tests
make test

# Run linters
make lint

# Clean build artifacts
make clean

# See all targets
make help
```

## WSL2 Users

To run the desktop GUI on WSL2, you need an X server:
- **Windows 11**: WSLg (built-in, should work out of the box)
- **Windows 10**: Install VcXsrv or X410

## Why Not `go build`?

**Don't use `go build` on the root directory!**

Wails apps require special build tags and the Wails runtime. The root `main.go` is a Wails app and must be built with `wails build`.

Only the CLI version in `cmd/cli/` can be built with standard `go build`.

