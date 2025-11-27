# JSIShell Makefile
# Build, test, and development targets

BINARY_NAME=jsishell
BUILD_DIR=./cmd/jsishell
COVERAGE_FILE=coverage.out
DIST_DIR=dist

# Version info (can be overridden: make VERSION=1.0.0)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go commands
GO=go
GOFMT=gofmt
GOVET=$(GO) vet
GOTEST=$(GO) test
GOBUILD=$(GO) build

# Build flags
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

.PHONY: all build build-release test test-coverage test-race lint fmt vet clean help release

# Default target
all: fmt vet test build

# Build development binary
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BUILD_DIR)

# Build release binary (stripped, smaller)
build-release:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(BUILD_DIR)

# Cross-compile for all platforms
build-all: build-linux build-linux-arm64 build-windows build-darwin build-darwin-arm64

build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 $(BUILD_DIR)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 $(BUILD_DIR)

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe $(BUILD_DIR)

build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 $(BUILD_DIR)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 $(BUILD_DIR)

# Create release artifacts
release: clean build-all
	@mkdir -p $(DIST_DIR)
	@mv $(BINARY_NAME)-linux-amd64 $(DIST_DIR)/
	@mv $(BINARY_NAME)-linux-arm64 $(DIST_DIR)/
	@mv $(BINARY_NAME)-windows-amd64.exe $(DIST_DIR)/
	@mv $(BINARY_NAME)-darwin-amd64 $(DIST_DIR)/
	@mv $(BINARY_NAME)-darwin-arm64 $(DIST_DIR)/
	@echo "Release artifacts created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# Run all tests
test:
	$(GOTEST) ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	$(GOTEST) -race ./...

# Run tests verbose
test-verbose:
	$(GOTEST) -v ./...

# Format code
fmt:
	$(GOFMT) -s -w .

# Check formatting (for CI)
fmt-check:
	@test -z "$$($(GOFMT) -l .)" || (echo "Code not formatted. Run 'make fmt'" && exit 1)

# Run go vet
vet:
	$(GOVET) ./...

# Run staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
lint:
	@which staticcheck > /dev/null || (echo "Installing staticcheck..." && go install honnef.co/go/tools/cmd/staticcheck@latest)
	staticcheck ./...

# Run all checks
check: fmt-check vet lint test

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	rm -f $(COVERAGE_FILE) coverage.html
	rm -rf $(DIST_DIR)

# Run the shell
run: build
	./$(BINARY_NAME)

# Install to GOPATH/bin
install:
	$(GO) install $(BUILD_DIR)

# Tidy go modules
tidy:
	$(GO) mod tidy

# Show help
help:
	@echo "JSIShell Makefile targets:"
	@echo ""
	@echo "  build         - Build development binary"
	@echo "  build-release - Build optimized release binary"
	@echo "  build-all     - Cross-compile for all platforms"
	@echo "  release       - Create release artifacts in dist/"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  fmt           - Format code"
	@echo "  fmt-check     - Check code formatting"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Run staticcheck"
	@echo "  check         - Run all checks (fmt, vet, lint, test)"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the shell"
	@echo "  install       - Install to GOPATH/bin"
	@echo "  tidy          - Tidy go modules"
	@echo "  help          - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION       - Set release version (default: git describe)"
	@echo ""
	@echo "Examples:"
	@echo "  make release VERSION=1.0.0"
	@echo "  make build-release"
