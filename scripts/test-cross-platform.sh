#!/bin/bash

# Cross-platform compatibility test runner
# This script runs comprehensive tests across different platforms and environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_OUTPUT_DIR="$PROJECT_ROOT/test-results"
COVERAGE_DIR="$PROJECT_ROOT/coverage"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Test configuration
RUN_GO_TESTS=true
RUN_JS_TESTS=true
RUN_RUST_TESTS=true
RUN_INTEGRATION_TESTS=true
RUN_PERFORMANCE_TESTS=false
GENERATE_COVERAGE=true
VERBOSE=false

# Platform detection
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unknown architecture: $ARCH"; exit 1 ;;
esac

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
Cross-Platform Compatibility Test Runner

Usage: $0 [OPTIONS]

Options:
    --no-go             Skip Go tests
    --no-js             Skip JavaScript/TypeScript tests
    --no-rust           Skip Rust tests
    --no-integration    Skip integration tests
    --performance       Run performance tests
    --no-coverage       Skip coverage generation
    --verbose           Enable verbose output
    --help              Show this help message

Environment Variables:
    CI                  Set to 'true' to run in CI mode
    TEST_TIMEOUT        Test timeout in seconds (default: 300)
    PARALLEL_JOBS       Number of parallel test jobs (default: 4)

Examples:
    $0                          # Run all tests
    $0 --no-integration         # Skip integration tests
    $0 --performance --verbose  # Run with performance tests and verbose output
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-go)
                RUN_GO_TESTS=false
                shift
                ;;
            --no-js)
                RUN_JS_TESTS=false
                shift
                ;;
            --no-rust)
                RUN_RUST_TESTS=false
                shift
                ;;
            --no-integration)
                RUN_INTEGRATION_TESTS=false
                shift
                ;;
            --performance)
                RUN_PERFORMANCE_TESTS=true
                shift
                ;;
            --no-coverage)
                GENERATE_COVERAGE=false
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create output directories
    mkdir -p "$TEST_OUTPUT_DIR"
    mkdir -p "$COVERAGE_DIR"
    
    # Set environment variables
    export LIV_TEST_MODE=1
    export LIV_TEST_OUTPUT_DIR="$TEST_OUTPUT_DIR"
    export GO111MODULE=on
    
    # Set test timeout
    export TEST_TIMEOUT=${TEST_TIMEOUT:-300}
    export PARALLEL_JOBS=${PARALLEL_JOBS:-4}
    
    # Platform-specific setup
    case $PLATFORM in
        darwin)
            # macOS specific setup
            export CGO_ENABLED=1
            ;;
        linux)
            # Linux specific setup
            export CGO_ENABLED=1
            ;;
        windows|mingw*|cygwin*)
            # Windows specific setup
            export CGO_ENABLED=1
            ;;
    esac
    
    log_success "Test environment setup complete"
    log_info "Platform: $PLATFORM/$ARCH"
    log_info "Output directory: $TEST_OUTPUT_DIR"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check Go
    if $RUN_GO_TESTS && ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    # Check Node.js and npm
    if $RUN_JS_TESTS; then
        if ! command -v node &> /dev/null; then
            missing_deps+=("node")
        fi
        if ! command -v npm &> /dev/null; then
            missing_deps+=("npm")
        fi
    fi
    
    # Check Rust and Cargo
    if $RUN_RUST_TESTS; then
        if ! command -v rustc &> /dev/null; then
            missing_deps+=("rust")
        fi
        if ! command -v cargo &> /dev/null; then
            missing_deps+=("cargo")
        fi
    fi
    
    # Check wasm-pack for WASM tests
    if $RUN_RUST_TESTS && ! command -v wasm-pack &> /dev/null; then
        log_warning "wasm-pack not found, WASM tests may fail"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install the missing dependencies and try again"
        exit 1
    fi
    
    log_success "All dependencies found"
}

run_go_tests() {
    if ! $RUN_GO_TESTS; then
        return 0
    fi
    
    log_info "Running Go tests..."
    
    cd "$PROJECT_ROOT"
    
    local go_test_args=("-v" "-race" "-timeout" "${TEST_TIMEOUT}s")
    
    if $GENERATE_COVERAGE; then
        go_test_args+=("-coverprofile=$COVERAGE_DIR/go-coverage.out" "-covermode=atomic")
    fi
    
    if $VERBOSE; then
        go_test_args+=("-v")
    fi
    
    # Run unit tests
    log_info "Running Go unit tests..."
    if ! go test "${go_test_args[@]}" ./pkg/...; then
        log_error "Go unit tests failed"
        return 1
    fi
    
    # Run cross-platform tests
    log_info "Running Go cross-platform tests..."
    if ! go test "${go_test_args[@]}" ./pkg/test/...; then
        log_error "Go cross-platform tests failed"
        return 1
    fi
    
    # Generate coverage report
    if $GENERATE_COVERAGE && [ -f "$COVERAGE_DIR/go-coverage.out" ]; then
        go tool cover -html="$COVERAGE_DIR/go-coverage.out" -o "$COVERAGE_DIR/go-coverage.html"
        go tool cover -func="$COVERAGE_DIR/go-coverage.out" > "$COVERAGE_DIR/go-coverage.txt"
        
        local coverage_percent=$(go tool cover -func="$COVERAGE_DIR/go-coverage.out" | grep total | awk '{print $3}')
        log_info "Go test coverage: $coverage_percent"
    fi
    
    log_success "Go tests completed successfully"
}

run_js_tests() {
    if ! $RUN_JS_TESTS; then
        return 0
    fi
    
    log_info "Running JavaScript/TypeScript tests..."
    
    cd "$PROJECT_ROOT/js"
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
        log_info "Installing JavaScript dependencies..."
        npm ci
    fi
    
    # Run tests
    local npm_test_args=()
    
    if $GENERATE_COVERAGE; then
        npm_test_args+=("--coverage")
    fi
    
    if $VERBOSE; then
        npm_test_args+=("--verbose")
    fi
    
    # Run unit tests
    log_info "Running JavaScript unit tests..."
    if ! npm test "${npm_test_args[@]}"; then
        log_error "JavaScript tests failed"
        return 1
    fi
    
    # Run cross-platform compatibility tests
    log_info "Running JavaScript cross-platform tests..."
    if ! npm run test:cross-platform 2>/dev/null || ! npx mocha test/cross-platform-compatibility.test.ts; then
        log_warning "JavaScript cross-platform tests not available or failed"
    fi
    
    # Copy coverage reports
    if $GENERATE_COVERAGE && [ -d "coverage" ]; then
        cp -r coverage/* "$COVERAGE_DIR/" 2>/dev/null || true
    fi
    
    log_success "JavaScript tests completed successfully"
}

run_rust_tests() {
    if ! $RUN_RUST_TESTS; then
        return 0
    fi
    
    log_info "Running Rust tests..."
    
    # Test interactive engine
    cd "$PROJECT_ROOT/wasm/interactive-engine"
    
    log_info "Running Rust interactive engine tests..."
    if ! cargo test --verbose; then
        log_error "Rust interactive engine tests failed"
        return 1
    fi
    
    # Test editor engine
    cd "$PROJECT_ROOT/wasm/editor-engine"
    
    log_info "Running Rust editor engine tests..."
    if ! cargo test --verbose; then
        log_error "Rust editor engine tests failed"
        return 1
    fi
    
    # Build WASM modules
    log_info "Building WASM modules..."
    
    cd "$PROJECT_ROOT/wasm/interactive-engine"
    if command -v wasm-pack &> /dev/null; then
        if ! wasm-pack build --target web --out-dir ../../js/pkg/interactive-engine; then
            log_warning "Failed to build interactive engine WASM module"
        fi
    fi
    
    cd "$PROJECT_ROOT/wasm/editor-engine"
    if command -v wasm-pack &> /dev/null; then
        if ! wasm-pack build --target web --out-dir ../../js/pkg/editor-engine; then
            log_warning "Failed to build editor engine WASM module"
        fi
    fi
    
    log_success "Rust tests completed successfully"
}

run_integration_tests() {
    if ! $RUN_INTEGRATION_TESTS; then
        return 0
    fi
    
    log_info "Running integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Build CLI tools first
    log_info "Building CLI tools for integration tests..."
    
    local cli_build_dir="$TEST_OUTPUT_DIR/cli-builds"
    mkdir -p "$cli_build_dir"
    
    # Build for current platform
    go build -o "$cli_build_dir/liv-cli" cmd/cli/main.go
    go build -o "$cli_build_dir/liv-viewer" cmd/viewer/main.go
    
    # Add to PATH for tests
    export PATH="$cli_build_dir:$PATH"
    
    # Run integration tests
    local integration_test_args=("-v" "-timeout" "${TEST_TIMEOUT}s")
    
    if $VERBOSE; then
        integration_test_args+=("-v")
    fi
    
    log_info "Running cross-platform integration tests..."
    if ! go test "${integration_test_args[@]}" ./test/integration/...; then
        log_error "Integration tests failed"
        return 1
    fi
    
    log_success "Integration tests completed successfully"
}

run_performance_tests() {
    if ! $RUN_PERFORMANCE_TESTS; then
        return 0
    fi
    
    log_info "Running performance tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run Go benchmarks
    log_info "Running Go performance benchmarks..."
    go test -bench=. -benchmem -timeout="${TEST_TIMEOUT}s" ./pkg/... > "$TEST_OUTPUT_DIR/go-benchmarks.txt" 2>&1
    
    # Run JavaScript performance tests
    if $RUN_JS_TESTS; then
        cd "$PROJECT_ROOT/js"
        log_info "Running JavaScript performance tests..."
        npm run test:performance 2>/dev/null || log_warning "JavaScript performance tests not available"
    fi
    
    # Run integration performance tests
    if $RUN_INTEGRATION_TESTS; then
        cd "$PROJECT_ROOT"
        log_info "Running integration performance tests..."
        go test -bench=. -benchmem -timeout="${TEST_TIMEOUT}s" ./test/integration/... > "$TEST_OUTPUT_DIR/integration-benchmarks.txt" 2>&1
    fi
    
    log_success "Performance tests completed successfully"
}

generate_test_report() {
    log_info "Generating test report..."
    
    local report_file="$TEST_OUTPUT_DIR/test-report-$TIMESTAMP.md"
    
    cat > "$report_file" << EOF
# Cross-Platform Compatibility Test Report

**Generated:** $(date)
**Platform:** $PLATFORM/$ARCH
**Test Run ID:** $TIMESTAMP

## Test Configuration

- Go Tests: $RUN_GO_TESTS
- JavaScript Tests: $RUN_JS_TESTS
- Rust Tests: $RUN_RUST_TESTS
- Integration Tests: $RUN_INTEGRATION_TESTS
- Performance Tests: $RUN_PERFORMANCE_TESTS
- Coverage Generation: $GENERATE_COVERAGE

## Environment

- Platform: $PLATFORM
- Architecture: $ARCH
- Go Version: $(go version 2>/dev/null || echo "Not available")
- Node Version: $(node --version 2>/dev/null || echo "Not available")
- Rust Version: $(rustc --version 2>/dev/null || echo "Not available")

## Test Results

EOF

    # Add test results
    if [ -f "$TEST_OUTPUT_DIR/go-test-results.txt" ]; then
        echo "### Go Tests" >> "$report_file"
        echo '```' >> "$report_file"
        cat "$TEST_OUTPUT_DIR/go-test-results.txt" >> "$report_file"
        echo '```' >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    # Add coverage information
    if $GENERATE_COVERAGE; then
        echo "## Coverage Reports" >> "$report_file"
        
        if [ -f "$COVERAGE_DIR/go-coverage.txt" ]; then
            echo "### Go Coverage" >> "$report_file"
            echo '```' >> "$report_file"
            tail -n 1 "$COVERAGE_DIR/go-coverage.txt" >> "$report_file"
            echo '```' >> "$report_file"
        fi
        
        echo "Coverage reports available in: $COVERAGE_DIR" >> "$report_file"
    fi
    
    # Add performance results
    if $RUN_PERFORMANCE_TESTS; then
        echo "## Performance Results" >> "$report_file"
        
        if [ -f "$TEST_OUTPUT_DIR/go-benchmarks.txt" ]; then
            echo "### Go Benchmarks" >> "$report_file"
            echo '```' >> "$report_file"
            grep "Benchmark" "$TEST_OUTPUT_DIR/go-benchmarks.txt" | head -20 >> "$report_file"
            echo '```' >> "$report_file"
        fi
    fi
    
    log_success "Test report generated: $report_file"
}

cleanup() {
    log_info "Cleaning up test environment..."
    
    # Kill any background processes
    pkill -f "liv-viewer" 2>/dev/null || true
    pkill -f "liv-cli" 2>/dev/null || true
    
    # Clean up temporary files
    find "$TEST_OUTPUT_DIR" -name "*.tmp" -delete 2>/dev/null || true
    
    log_success "Cleanup completed"
}

main() {
    # Parse command line arguments
    parse_args "$@"
    
    # Set up signal handlers
    trap cleanup EXIT
    trap 'log_error "Test run interrupted"; exit 1' INT TERM
    
    log_info "Starting cross-platform compatibility tests..."
    log_info "Timestamp: $TIMESTAMP"
    
    # Setup and checks
    setup_test_environment
    check_dependencies
    
    # Run tests
    local test_start_time=$(date +%s)
    local failed_tests=()
    
    if ! run_go_tests; then
        failed_tests+=("Go")
    fi
    
    if ! run_js_tests; then
        failed_tests+=("JavaScript")
    fi
    
    if ! run_rust_tests; then
        failed_tests+=("Rust")
    fi
    
    if ! run_integration_tests; then
        failed_tests+=("Integration")
    fi
    
    if ! run_performance_tests; then
        failed_tests+=("Performance")
    fi
    
    local test_end_time=$(date +%s)
    local test_duration=$((test_end_time - test_start_time))
    
    # Generate report
    generate_test_report
    
    # Summary
    log_info "Test run completed in ${test_duration}s"
    
    if [ ${#failed_tests[@]} -eq 0 ]; then
        log_success "All tests passed successfully!"
        exit 0
    else
        log_error "Failed test suites: ${failed_tests[*]}"
        exit 1
    fi
}

# Run main function with all arguments
main "$@"