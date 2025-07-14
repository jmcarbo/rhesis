# Makefile for Rhesis presentation generator

# Variables
BINARY_NAME=rhesis
BINARY_PATH=./bin/$(BINARY_NAME)
CMD_PATH=./cmd/rhesis
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go commands
GO=go
GOFMT=gofmt
GOLINT=golangci-lint
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOMOD=$(GO) mod
GOGET=$(GO) get

# Build flags
LDFLAGS=-ldflags "-w -s"
BUILD_FLAGS=-v $(LDFLAGS)

# Default target
.PHONY: all
all: clean deps lint test build

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Update dependencies to latest versions
.PHONY: update-deps
update-deps:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Lint code
.PHONY: lint
lint:
	$(GOLINT) run

# Run tests
.PHONY: test
test:
	$(GOTEST) -v -race ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Build the application
.PHONY: build
build: fmt
	mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_PATH) $(CMD_PATH)

# Install the application
.PHONY: install
install:
	$(GO) install $(CMD_PATH)

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

# Run the application
.PHONY: run
run: build
	$(BINARY_PATH)

# Development workflow
.PHONY: dev
dev: clean fmt lint test build

# Release build
.PHONY: release
release: clean deps
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)

# Check if golangci-lint is installed
.PHONY: check-lint
check-lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)

# Install golangci-lint
.PHONY: install-lint
install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all           - Run clean, deps, lint, test, build"
	@echo "  deps          - Install dependencies"
	@echo "  update-deps   - Update dependencies to latest versions"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the application"
	@echo "  install       - Install the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Development workflow (clean, fmt, lint, test, build)"
	@echo "  release       - Build release binaries for multiple platforms"
	@echo "  install-lint  - Install golangci-lint"
	@echo "  help          - Show this help message"