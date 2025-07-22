# Klaunch Makefile
# Provides convenient commands for building, testing, and running Klaunch

.PHONY: build test clean install dev run-tests unit-tests integration-tests infrastructure-tests benchmarks static-analysis coverage help

# Build configuration
BINARY_NAME=klaunch
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_LINUX=$(BINARY_NAME)_linux
BUILD_DIR=build
VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go configuration
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

# Default target
all: clean build test

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v

# Build for different platforms
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_LINUX) -v

build-all: build build-linux

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test: run-tests

# Run all tests using the test script
run-tests:
	@echo "Running all tests..."
	./run_tests.sh all

# Run only unit tests
unit-tests:
	@echo "Running unit tests..."
	./run_tests.sh unit

# Run integration tests
integration-tests:
	@echo "Running integration tests..."
	./run_tests.sh integration

# Run infrastructure tests
infrastructure-tests:
	@echo "Running infrastructure tests..."
	./run_tests.sh infrastructure

# Run benchmarks
benchmarks:
	@echo "Running benchmarks..."
	./run_tests.sh benchmarks

# Run static analysis
static-analysis:
	@echo "Running static analysis..."
	./run_tests.sh static

# Generate coverage report
coverage:
	@echo "Generating coverage report..."
	@mkdir -p coverage
	$(GOTEST) -coverprofile=coverage/coverage.out ./test
	$(GOCMD) tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

# Run code quality checks
quality: fmt vet lint static-analysis

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf coverage
	rm -rf test-results
	rm -f klaunch-test klaunch-integration-test klaunch-benchmark
	rm -f *.out *.prof

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(LDFLAGS)

# Development mode - build and run with file watching (requires entr)
dev:
	@echo "Starting development mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name "*.go" | entr -r make build && $(BUILD_DIR)/$(BINARY_NAME); \
	else \
		echo "entr not found. Install it for file watching: brew install entr (macOS) or apt-get install entr (Ubuntu)"; \
		make build; \
	fi

# Start the infrastructure (Docker Compose)
start-infra:
	@echo "Starting infrastructure..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -p klaunch up -d; \
	elif docker compose version >/dev/null 2>&1; then \
		docker compose -p klaunch up -d; \
	else \
		echo "Docker Compose not found. Please install Docker Compose."; \
	fi

# Stop the infrastructure
stop-infra:
	@echo "Stopping infrastructure..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -p klaunch down; \
	elif docker compose version >/dev/null 2>&1; then \
		docker compose -p klaunch down; \
	else \
		echo "Docker Compose not found."; \
	fi

# Show infrastructure status
status-infra:
	@echo "Infrastructure status..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -p klaunch ps; \
	elif docker compose version >/dev/null 2>&1; then \
		docker compose -p klaunch ps; \
	else \
		echo "Docker Compose not found."; \
	fi

# Build and test in one command
build-and-test: clean deps build test

# Release build (optimized)
release: clean
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -ldflags "-s -w" -v
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Create distribution package
dist: release build-linux
	@echo "Creating distribution package..."
	@mkdir -p dist
	@tar -czf dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)
	@tar -czf dist/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_LINUX)
	@echo "Distribution packages created in dist/"

# Run security scan (requires gosec)
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install it with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p docs
	$(GOCMD) doc -all > docs/godoc.txt
	@echo "Documentation generated in docs/"

# Run performance profiling
profile:
	@echo "Running performance profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -memprofile=mem.prof -bench=. ./...
	@echo "Profile files generated: cpu.prof, mem.prof"
	@echo "View with: go tool pprof cpu.prof"

# Database/Volume cleanup
clean-volumes:
	@echo "Cleaning Docker volumes..."
	@if docker volume ls -q -f name=klaunch | grep -q .; then \
		docker volume rm $$(docker volume ls -q -f name=klaunch); \
	else \
		echo "No klaunch volumes found."; \
	fi

# Show project information
info:
	@echo "Project Information:"
	@echo "  Name: $(BINARY_NAME)"
	@echo "  Version: $(VERSION)"
	@echo "  Commit: $(COMMIT)"
	@echo "  Build Date: $(DATE)"
	@echo "  Go Version: $$($(GOCMD) version)"
	@echo "  Build Directory: $(BUILD_DIR)"

# Help target
help:
	@echo "Klaunch Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build           - Build the application"
	@echo "  build-linux     - Build for Linux"
	@echo "  build-all       - Build for all platforms"
	@echo "  release         - Build optimized release version"
	@echo "  install         - Install binary to GOPATH/bin"
	@echo "  dist            - Create distribution packages"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev             - Development mode with file watching"
	@echo "  deps            - Install/update dependencies"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  lint            - Run linter (requires golangci-lint)"
	@echo "  quality         - Run all code quality checks"
	@echo "  security        - Run security scan (requires gosec)"
	@echo ""
	@echo "Test Commands:"
	@echo "  test            - Run all tests"
	@echo "  run-tests       - Run all tests using test script"
	@echo "  unit-tests      - Run unit tests only"
	@echo "  integration-tests - Run integration tests only"
	@echo "  infrastructure-tests - Run infrastructure tests only"
	@echo "  benchmarks      - Run performance benchmarks"
	@echo "  static-analysis - Run static analysis"
	@echo "  coverage        - Generate coverage report"
	@echo "  profile         - Run performance profiling"
	@echo ""
	@echo "Infrastructure Commands:"
	@echo "  start-infra     - Start Docker infrastructure"
	@echo "  stop-infra      - Stop Docker infrastructure"
	@echo "  status-infra    - Show infrastructure status"
	@echo "  clean-volumes   - Clean Docker volumes"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean           - Clean build artifacts"
	@echo "  docs            - Generate documentation"
	@echo "  info            - Show project information"
	@echo "  help            - Show this help message"
	@echo ""