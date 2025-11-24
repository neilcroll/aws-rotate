.PHONY: all build test lint clean install help

# Variables
BINARY_NAME=aws-rotate
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=dist
GO=go
GOFMT=gofmt
GOLANGCI_LINT=golangci-lint

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Default target
all: clean lint test build

## help: Display this help message
help:
	@echo "AWS Multi-Account Credentials Rotator - Makefile targets:"
	@echo ""
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary for current platform
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@$(GO) build $(LDFLAGS) -o $(BINARY_NAME)
	@echo "Build complete: ./$(BINARY_NAME)"

## build-all: Build binaries for all platforms
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64
	
	@echo "Building Linux ARM64..."
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64
	
	@echo "Building macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64
	
	@echo "Building macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64
	
	@echo "Building Windows AMD64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe
	
	@echo "All builds complete in $(BUILD_DIR)/"

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
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME).exe
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## install: Install binary to system
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin/..."
	@sudo mv $(BINARY_NAME) /usr/local/bin/
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

## run: Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

## check: Run all checks (lint + test)
check: lint test
	@echo "All checks passed"

