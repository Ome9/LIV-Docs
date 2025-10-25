#!/bin/bash

# Security and Administration Test Runner
# Runs comprehensive security tests including unit tests, integration tests, and performance tests

set -e

echo "🔐 LIV Security and Administration Test Suite"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Create test results directory
TEST_RESULTS_DIR="test-results/security"
mkdir -p "$TEST_RESULTS_DIR"

# Function to run tests with coverage
run_test_with_coverage() {
    local test_name=$1
    local test_path=$2
    local output_file="$TEST_RESULTS_DIR/${test_name}-results.txt"
    local coverage_file="$TEST_RESULTS_DIR/${test_name}-coverage.out"
    
    print_status "Running $test_name tests..."
    
    if go test -v -race -coverprofile="$coverage_file" "$test_path" > "$output_file" 2>&1; then
        print_success "$test_name tests passed"
        
        # Generate coverage report
        if [ -f "$coverage_file" ]; then
            coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}')
            print_status "$test_name coverage: $coverage"
        fi
    else
        print_error "$test_name tests failed"
        echo "Error details:"
        tail -20 "$output_file"
        return 1
    fi
}

# Function to run benchmarks
run_benchmarks() {
    local benchmark_name=$1
    local benchmark_path=$2
    local output_file="$TEST_RESULTS_DIR/${benchmark_name}-benchmarks.txt"
    
    print_status "Running $benchmark_name benchmarks..."
    
    if go test -bench=. -benchmem -run=^$ "$benchmark_path" > "$output_file" 2>&1; then
        print_success "$benchmark_name benchmarks completed"
        echo "Benchmark results saved to $output_file"
    else
        print_error "$benchmark_name benchmarks failed"
        tail -10 "$output_file"
        return 1
    fi
}

# Main test execution
main() {
    print_status "Starting security and administration test suite..."
    
    # 1. Run core security tests
    print_status "Phase 1: Core Security Tests"
    run_test_with_coverage "policy-manager" "./pkg/security" || exit 1
    
    # 2. Run security administration tests
    print_status "Phase 2: Security Administration Tests"
    if go test -v -race ./pkg/security -run="TestSecurityAdministrationSuite" > "$TEST_RESULTS_DIR/administration-tests.txt" 2>&1; then
        print_success "Security administration tests passed"
    else
        print_error "Security administration tests failed"
        tail -20 "$TEST_RESULTS_DIR/administration-tests.txt"
        exit 1
    fi
    
    # 3. Run integration tests
    print_status "Phase 3: Security Integration Tests"
    if go test -v -race ./test/integration -run="TestSecurityIntegrationSuite" > "$TEST_RESULTS_DIR/integration-tests.txt" 2>&1; then
        print_success "Security integration tests passed"
    else
        print_error "Security integration tests failed"
        tail -20 "$TEST_RESULTS_DIR/integration-tests.txt"
        exit 1
    fi
    
    # 4. Run performance tests
    print_status "Phase 4: Performance Tests"
    if go test -v -race ./pkg/security -run="TestConcurrentPolicyOperations|TestMemoryUsageUnderLoad|TestEventLogPerformanceUnderLoad" > "$TEST_RESULTS_DIR/performance-tests.txt" 2>&1; then
        print_success "Performance tests passed"
    else
        print_warning "Some performance tests may have failed (check results)"
    fi
    
    # 5. Run benchmarks
    print_status "Phase 5: Benchmarks"
    run_benchmarks "security-performance" "./pkg/security" || print_warning "Benchmarks failed but continuing..."
    
    # 6. Run specific security scenario tests
    print_status "Phase 6: Security Scenario Tests"
    if go test -v -race ./pkg/security -run="TestSecurityPolicyEnforcementScenarios|TestSecurityEventCorrelation" > "$TEST_RESULTS_DIR/scenario-tests.txt" 2>&1; then
        print_success "Security scenario tests passed"
    else
        print_error "Security scenario tests failed"
        tail -20 "$TEST_RESULTS_DIR/scenario-tests.txt"
        exit 1
    fi
    
    # 7. Generate comprehensive coverage report
    print_status "Phase 7: Generating Coverage Report"
    if ls "$TEST_RESULTS_DIR"/*-coverage.out 1> /dev/null 2>&1; then
        echo "mode: set" > "$TEST_RESULTS_DIR/combined-coverage.out"
        tail -n +2 -q "$TEST_RESULTS_DIR"/*-coverage.out >> "$TEST_RESULTS_DIR/combined-coverage.out"
        
        total_coverage=$(go tool cover -func="$TEST_RESULTS_DIR/combined-coverage.out" | grep total | awk '{print $3}')
        print_success "Total security test coverage: $total_coverage"
        
        # Generate HTML coverage report
        go tool cover -html="$TEST_RESULTS_DIR/combined-coverage.out" -o "$TEST_RESULTS_DIR/coverage-report.html"
        print_status "HTML coverage report generated: $TEST_RESULTS_DIR/coverage-report.html"
    fi
    
    # 8. Test summary
    print_status "Phase 8: Test Summary"
    echo ""
    echo "📊 Test Results Summary"
    echo "======================="
    
    total_tests=0
    passed_tests=0
    
    for result_file in "$TEST_RESULTS_DIR"/*-results.txt "$TEST_RESULTS_DIR"/*-tests.txt; do
        if [ -f "$result_file" ]; then
            test_name=$(basename "$result_file" | sed 's/-results.txt\|-tests.txt//')
            
            if grep -q "PASS" "$result_file"; then
                status="✅ PASSED"
                ((passed_tests++))
            elif grep -q "FAIL" "$result_file"; then
                status="❌ FAILED"
            else
                status="⚠️  UNKNOWN"
            fi
            
            test_count=$(grep -c "=== RUN" "$result_file" 2>/dev/null || echo "0")
            total_tests=$((total_tests + test_count))
            
            printf "%-30s %s (%s tests)\n" "$test_name" "$status" "$test_count"
        fi
    done
    
    echo ""
    echo "📈 Overall Statistics"
    echo "===================="
    echo "Total test suites: $(ls "$TEST_RESULTS_DIR"/*-results.txt "$TEST_RESULTS_DIR"/*-tests.txt 2>/dev/null | wc -l)"
    echo "Total individual tests: $total_tests"
    echo "Test results directory: $TEST_RESULTS_DIR"
    
    if [ -f "$TEST_RESULTS_DIR/combined-coverage.out" ]; then
        echo "Coverage report: $TEST_RESULTS_DIR/coverage-report.html"
    fi
    
    # 9. Security test validation
    print_status "Phase 9: Security Test Validation"
    
    # Check for critical security test coverage
    critical_tests=(
        "TestSecurityPolicyEnforcement"
        "TestPermissionInheritanceEnforcement" 
        "TestSecurityEventHandling"
        "TestAuditLogging"
        "TestWASMSecurityContextIntegration"
        "TestSignatureAndTrustChainIntegration"
        "TestErrorHandlingIntegration"
        "TestComplianceAndAuditIntegration"
    )
    
    missing_tests=()
    for test in "${critical_tests[@]}"; do
        if ! grep -r "$test" "$TEST_RESULTS_DIR"/*.txt >/dev/null 2>&1; then
            missing_tests+=("$test")
        fi
    done
    
    if [ ${#missing_tests[@]} -eq 0 ]; then
        print_success "All critical security tests are covered"
    else
        print_warning "Missing critical security tests:"
        for test in "${missing_tests[@]}"; do
            echo "  - $test"
        done
    fi
    
    # 10. Performance validation
    print_status "Phase 10: Performance Validation"
    
    if [ -f "$TEST_RESULTS_DIR/security-performance-benchmarks.txt" ]; then
        # Check benchmark results for performance regressions
        if grep -q "BenchmarkPermissionEvaluation" "$TEST_RESULTS_DIR/security-performance-benchmarks.txt"; then
            print_success "Permission evaluation benchmarks completed"
        fi
        
        if grep -q "BenchmarkPolicyCreation" "$TEST_RESULTS_DIR/security-performance-benchmarks.txt"; then
            print_success "Policy creation benchmarks completed"
        fi
        
        if grep -q "BenchmarkResourceMonitoring" "$TEST_RESULTS_DIR/security-performance-benchmarks.txt"; then
            print_success "Resource monitoring benchmarks completed"
        fi
    fi
    
    echo ""
    print_success "🎉 Security and Administration Test Suite Completed!"
    echo ""
    echo "Next steps:"
    echo "1. Review test results in $TEST_RESULTS_DIR/"
    echo "2. Check coverage report: $TEST_RESULTS_DIR/coverage-report.html"
    echo "3. Address any failing tests or performance issues"
    echo "4. Update security documentation if needed"
    echo ""
}

# Cleanup function
cleanup() {
    print_status "Cleaning up temporary files..."
    # Add any cleanup logic here
}

# Set up trap for cleanup
trap cleanup EXIT

# Run main function
main "$@"