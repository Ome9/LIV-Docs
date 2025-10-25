# LIV Format Build System

.PHONY: all build clean test install dev

# Default target
all: build

# Build all components
build: build-go build-wasm build-js

# Build Go components
build-go:
	@echo "Building Go components..."
	go mod tidy
	go build -o bin/liv-cli ./cmd/cli
	go build -o bin/liv-viewer ./cmd/viewer
	go build -o bin/liv-builder ./cmd/builder

# Build WASM modules
build-wasm:
	@echo "Building WASM modules..."
	cd wasm/interactive-engine && wasm-pack build --target web --out-dir ../../js/wasm/interactive
	cd wasm/editor-engine && wasm-pack build --target web --out-dir ../../js/wasm/editor

# Build JavaScript/TypeScript
build-js:
	@echo "Building JavaScript components..."
	npm install
	npm run build

# Development mode
dev:
	@echo "Starting development servers..."
	npm run dev &
	go run ./cmd/viewer &

# Run tests
test: test-go test-wasm test-js

test-go:
	@echo "Running Go tests..."
	go test ./...

test-wasm:
	@echo "Running WASM tests..."
	cd wasm/interactive-engine && cargo test
	cd wasm/editor-engine && cargo test

test-js:
	@echo "Running JavaScript tests..."
	npm test

# Comprehensive test suite
test-all:
	@echo "Running comprehensive test suite..."
	cd test && go run run-all-tests.go

test-unit:
	@echo "Running unit tests..."
	cd test && go run run-all-tests.go unit

test-security:
	@echo "Running security tests..."
	cd test && go run run-all-tests.go security

test-integration:
	@echo "Running integration tests..."
	cd test && go run run-all-tests.go integration

test-performance:
	@echo "Running performance tests..."
	cd test && go run run-all-tests.go performance

test-e2e:
	@echo "Running end-to-end tests..."
	cd test && go run run-all-tests.go e2e

test-cross-platform:
	@echo "Running cross-platform tests..."
	cd test && go run run-all-tests.go cross-platform

test-sdk:
	@echo "Running SDK integration tests..."
	cd test && go run run-all-tests.go sdk

test-fast:
	@echo "Running fast tests..."
	cd test && go run run-all-tests.go fast

test-slow:
	@echo "Running slow tests..."
	cd test && go run run-all-tests.go slow

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-coverage-all:
	@echo "Running comprehensive tests with coverage..."
	GENERATE_COVERAGE=true cd test && go run run-all-tests.go

# Test with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v ./...

test-verbose-all:
	@echo "Running comprehensive tests with verbose output..."
	cd test && go run run-all-tests.go -v

# Test with parallel execution
test-parallel:
	@echo "Running tests in parallel..."
	cd test && go run run-all-tests.go -p

# Test with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -race ./...

# Clean test artifacts
test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage*.out coverage*.html
	rm -rf test/tmp test/temp
	find . -name "*.test" -delete

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	npm install
	rustup target add wasm32-unknown-unknown
	cargo install wasm-pack

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf js/dist/
	rm -rf js/wasm/
	rm -rf wasm/*/pkg/
	go clean

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	cd wasm/interactive-engine && cargo fmt
	cd wasm/editor-engine && cargo fmt
	npm run lint

# Check code quality
lint:
	@echo "Running linters..."
	golangci-lint run
	cd wasm/interactive-engine && cargo clippy
	cd wasm/editor-engine && cargo clippy
	npm run lint

# Generate documentation
docs:
	@echo "Generating documentation..."
	go doc -all ./... > docs/go-api.md
	cd wasm/interactive-engine && cargo doc --no-deps
	cd wasm/editor-engine && cargo doc --no-deps
	npm run docs

# Create release build
release: clean
	@echo "Creating release build..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/linux/liv-cli ./cmd/cli
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/darwin/liv-cli ./cmd/cli
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/windows/liv-cli.exe ./cmd/cli
	make build-wasm
	NODE_ENV=production npm run build

# Docker build
docker:
	@echo "Building Docker image..."
	docker build -t liv-format:latest .

# Help
help:
	@echo "Available targets:"
	@echo "  all               - Build all components (default)"
	@echo "  build             - Build all components"
	@echo "  build-go          - Build Go components only"
	@echo "  build-wasm        - Build WASM modules only"
	@echo "  build-js          - Build JavaScript components only"
	@echo "  dev               - Start development servers"
	@echo "  test              - Run basic tests (Go, WASM, JS)"
	@echo "  test-all          - Run comprehensive test suite"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-security     - Run security tests only"
	@echo "  test-integration  - Run integration tests only"
	@echo "  test-performance  - Run performance tests only"
	@echo "  test-e2e          - Run end-to-end tests only"
	@echo "  test-cross-platform - Run cross-platform tests only"
	@echo "  test-sdk          - Run SDK integration tests only"
	@echo "  test-fast         - Run fast tests only"
	@echo "  test-slow         - Run slow tests only"
	@echo "  test-coverage     - Run tests with coverage"
	@echo "  test-coverage-all - Run comprehensive tests with coverage"
	@echo "  test-verbose      - Run tests with verbose output"
	@echo "  test-verbose-all  - Run comprehensive tests with verbose output"
	@echo "  test-parallel     - Run tests in parallel"
	@echo "  test-race         - Run tests with race detection"
	@echo "  test-clean        - Clean test artifacts"
	@echo "  install           - Install dependencies"
	@echo "  clean             - Clean build artifacts"
	@echo "  fmt               - Format code"
	@echo "  lint              - Run linters"
	@echo "  docs              - Generate documentation"
	@echo "  release           - Create release build"
	@echo "  docker            - Build Docker image"
	@echo "  help              - Show this help"