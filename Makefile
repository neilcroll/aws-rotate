.PHONY: all build build-cli build-desktop test lint clean install help dev

# Variables
BINARY_NAME=aws-rotate
CLI_DIR=cmd/cli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=dist
GO=go
GOFMT=gofmt
GOLANGCI_LINT=golangci-lint
WAILS=wails

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Default target - build desktop GUI
all: clean lint test build-desktop

## help: Display this help message
help:
	@echo "AWS Multi-Account Credentials Rotator - Makefile targets:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the desktop GUI (requires Wails CLI)
build: build-desktop

## build-desktop: Build the desktop GUI with Wails
build-desktop:
	@echo "Building desktop GUI version $(VERSION)..."
	@npm install --prefix frontend
	@npm run build --prefix frontend
	@$(WAILS) build
	@echo "Desktop GUI build complete: ./build/bin/$(BINARY_NAME)"

## build-cli: Build the CLI-only version
build-cli:
	@echo "Building CLI version $(VERSION)..."
	@cd $(CLI_DIR) && $(GO) build $(LDFLAGS) -o $(BINARY_NAME)-cli
	@echo "CLI build complete: $(CLI_DIR)/$(BINARY_NAME)-cli"

## build-all: Build CLI binaries for all platforms
build-all: clean
	@echo "Building CLI for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building Linux AMD64..."
	@cd $(CLI_DIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o ../../$(BUILD_DIR)/$(BINARY_NAME)-cli-linux-amd64
	
	@echo "Building Linux ARM64..."
	@cd $(CLI_DIR) && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o ../../$(BUILD_DIR)/$(BINARY_NAME)-cli-linux-arm64
	
	@echo "Building macOS AMD64..."
	@cd $(CLI_DIR) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o ../../$(BUILD_DIR)/$(BINARY_NAME)-cli-darwin-amd64
	
	@echo "Building macOS ARM64..."
	@cd $(CLI_DIR) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o ../../$(BUILD_DIR)/$(BINARY_NAME)-cli-darwin-arm64
	
	@echo "Building Windows AMD64..."
	@cd $(CLI_DIR) && GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o ../../$(BUILD_DIR)/$(BINARY_NAME)-cli-windows-amd64.exe
	
	@echo "All CLI builds complete in $(BUILD_DIR)/"

## dev: Run desktop app in development mode with hot reload
dev:
	@echo "Starting development server..."
	@$(WAILS) dev

## test: Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint:
	@echo "Running go fmt..."
	@$(GOFMT) -s -w .
	
	@echo "Running go vet..."
	@$(GO) vet ./...
	
	@echo "Running golangci-lint..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run; \
	else \
		echo "golangci-lint not installed. Install it with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Formatting complete"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf build/bin
	@rm -f $(CLI_DIR)/$(BINARY_NAME)-cli
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf frontend/dist
	@echo "Clean complete"

## install: Install desktop GUI to system
install: build-desktop
	@echo "Installing desktop GUI to /usr/local/bin/..."
	@sudo cp build/bin/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

## install-cli: Install CLI to system
install-cli: build-cli
	@echo "Installing CLI to /usr/local/bin/..."
	@sudo cp $(CLI_DIR)/$(BINARY_NAME)-cli /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod verify
	@echo "Dependencies downloaded"

## tidy: Tidy go.mod
tidy:
	@echo "Tidying go.mod..."
	@$(GO) mod tidy
	@echo "Tidy complete"

## run: Run the desktop GUI application
run: build-desktop
	@echo "Running desktop GUI..."
	@./build/bin/$(BINARY_NAME)

## run-cli: Run the CLI application
run-cli: build-cli
	@echo "Running CLI..."
	@./$(CLI_DIR)/$(BINARY_NAME)-cli

## check: Run all checks (lint + test)
check: lint test
	@echo "All checks passed"

