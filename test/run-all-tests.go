package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// TestSuite represents a test suite configuration
type TestSuite struct {
	Name        string
	Path        string
	Description string
	Tags        []string
	Timeout     time.Duration
}

// TestRunner manages and executes test suites
type TestRunner struct {
	suites   []TestSuite
	verbose  bool
	parallel bool
}

// NewTestRunner creates a new test runner
func NewTestRunner() *TestRunner {
	return &TestRunner{
		suites: []TestSuite{
			{
				Name:        "Unit Tests",
				Path:        "./unit",
				Description: "Core functionality unit tests",
				Tags:        []string{"unit", "fast"},
				Timeout:     5 * time.Minute,
			},
			{
				Name:        "Security Tests",
				Path:        "./unit",
				Description: "Security validation and sanitization tests",
				Tags:        []string{"security", "unit"},
				Timeout:     10 * time.Minute,
			},
			{
				Name:        "Integration Tests",
				Path:        "./integration",
				Description: "Component integration tests",
				Tags:        []string{"integration", "medium"},
				Timeout:     15 * time.Minute,
			},
			{
				Name:        "SDK Integration Tests",
				Path:        "./integration",
				Description: "JavaScript and Python SDK integration tests",
				Tags:        []string{"sdk", "integration", "medium"},
				Timeout:     20 * time.Minute,
			},
			{
				Name:        "Performance Tests",
				Path:        "./performance",
				Description: "Performance and benchmark tests",
				Tags:        []string{"performance", "slow"},
				Timeout:     30 * time.Minute,
			},
			{
				Name:        "End-to-End Tests",
				Path:        "./e2e",
				Description: "Complete workflow tests",
				Tags:        []string{"e2e", "slow"},
				Timeout:     20 * time.Minute,
			},
			{
				Name:        "Cross-Platform Tests",
				Path:        "../pkg/test",
				Description: "Cross-platform compatibility tests",
				Tags:        []string{"cross-platform", "medium"},
				Timeout:     15 * time.Minute,
			},
		},
	}
}

// RunAll executes all test suites
func (tr *TestRunner) RunAll() error {
	fmt.Println("🚀 Starting LIV Document Format Test Suite")
	fmt.Println("==========================================")

	startTime := time.Now()
	totalTests := 0
	passedTests := 0
	failedTests := 0

	for _, suite := range tr.suites {
		fmt.Printf("\\n📋 Running %s\\n", suite.Name)
		fmt.Printf("   Description: %s\\n", suite.Description)
		fmt.Printf("   Path: %s\\n", suite.Path)
		fmt.Printf("   Tags: %s\\n", strings.Join(suite.Tags, ", "))

		result, err := tr.runSuite(suite)
		if err != nil {
			fmt.Printf("   ❌ FAILED: %v\\n", err)
			failedTests++
		} else {
			fmt.Printf("   ✅ PASSED\\n")
			passedTests++
		}

		if result != nil {
			totalTests += result.Total
		}
	}

	duration := time.Since(startTime)

	fmt.Println("\\n==========================================")
	fmt.Println("📊 Test Summary")
	fmt.Printf("   Total Suites: %d\\n", len(tr.suites))
	fmt.Printf("   Passed Suites: %d\\n", passedTests)
	fmt.Printf("   Failed Suites: %d\\n", failedTests)
	fmt.Printf("   Total Tests: %d\\n", totalTests)
	fmt.Printf("   Duration: %v\\n", duration)

	if failedTests > 0 {
		fmt.Println("   Status: ❌ SOME TESTS FAILED")
		return fmt.Errorf("%d test suites failed", failedTests)
	} else {
		fmt.Println("   Status: ✅ ALL TESTS PASSED")
	}

	return nil
}

// TestResult represents the result of running a test suite
type TestResult struct {
	Total    int
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
}

// runSuite executes a single test suite
func (tr *TestRunner) runSuite(suite TestSuite) (*TestResult, error) {
	// Check if test directory exists
	if _, err := os.Stat(suite.Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("test directory does not exist: %s", suite.Path)
	}

	// Prepare go test command
	args := []string{"test"}

	if tr.verbose {
		args = append(args, "-v")
	}

	// Add timeout
	args = append(args, "-timeout", suite.Timeout.String())

	// Add coverage if requested
	if shouldGenerateCoverage() {
		coverageFile := fmt.Sprintf("coverage-%s.out", strings.ToLower(strings.ReplaceAll(suite.Name, " ", "-")))
		args = append(args, "-coverprofile", coverageFile)
	}

	// Add race detection for relevant tests
	if shouldUseRaceDetection(suite) {
		args = append(args, "-race")
	}

	// Add parallel execution if enabled
	if tr.parallel && canRunInParallel(suite) {
		args = append(args, "-parallel", "4")
	}

	// Add test path
	args = append(args, suite.Path)

	// Execute the test
	cmd := exec.Command("go", args...)
	cmd.Dir = getProjectRoot()

	output, err := cmd.CombinedOutput()

	// Parse test results
	result := parseTestOutput(string(output))

	if tr.verbose || err != nil {
		fmt.Printf("   Output:\\n%s\\n", indentOutput(string(output)))
	}

	if err != nil {
		return result, fmt.Errorf("test execution failed: %v", err)
	}

	return result, nil
}

// RunSpecific runs specific test suites by name or tag
func (tr *TestRunner) RunSpecific(filters []string) error {
	var suitesToRun []TestSuite

	for _, suite := range tr.suites {
		shouldRun := false

		// Check if suite name matches any filter
		for _, filter := range filters {
			if strings.Contains(strings.ToLower(suite.Name), strings.ToLower(filter)) {
				shouldRun = true
				break
			}

			// Check if any tag matches the filter
			for _, tag := range suite.Tags {
				if strings.Contains(strings.ToLower(tag), strings.ToLower(filter)) {
					shouldRun = true
					break
				}
			}
		}

		if shouldRun {
			suitesToRun = append(suitesToRun, suite)
		}
	}

	if len(suitesToRun) == 0 {
		return fmt.Errorf("no test suites match the filters: %v", filters)
	}

	// Temporarily replace suites and run
	originalSuites := tr.suites
	tr.suites = suitesToRun
	err := tr.RunAll()
	tr.suites = originalSuites

	return err
}

// SetVerbose enables or disables verbose output
func (tr *TestRunner) SetVerbose(verbose bool) {
	tr.verbose = verbose
}

// SetParallel enables or disables parallel execution
func (tr *TestRunner) SetParallel(parallel bool) {
	tr.parallel = parallel
}

// Helper functions

func shouldGenerateCoverage() bool {
	return os.Getenv("GENERATE_COVERAGE") == "true"
}

func shouldUseRaceDetection(suite TestSuite) bool {
	// Enable race detection for integration and e2e tests
	for _, tag := range suite.Tags {
		if tag == "integration" || tag == "e2e" {
			return true
		}
	}
	return false
}

func canRunInParallel(suite TestSuite) bool {
	// Some tests (like e2e) might not be suitable for parallel execution
	for _, tag := range suite.Tags {
		if tag == "e2e" {
			return false
		}
	}
	return true
}

func getProjectRoot() string {
	// Assume we're running from test directory
	return ".."
}

func parseTestOutput(output string) *TestResult {
	result := &TestResult{}

	lines := strings.Split(output, "\\n")
	for _, line := range lines {
		if strings.Contains(line, "PASS") {
			result.Passed++
		} else if strings.Contains(line, "FAIL") {
			result.Failed++
		} else if strings.Contains(line, "SKIP") {
			result.Skipped++
		}
	}

	result.Total = result.Passed + result.Failed + result.Skipped
	return result
}

func indentOutput(output string) string {
	lines := strings.Split(output, "\\n")
	var indented []string
	for _, line := range lines {
		indented = append(indented, "     "+line)
	}
	return strings.Join(indented, "\\n")
}

// Main function
func main() {
	runner := NewTestRunner()

	// Parse command line arguments
	args := os.Args[1:]

	var filters []string
	verbose := false
	parallel := false

	for i, arg := range args {
		switch arg {
		case "-v", "--verbose":
			verbose = true
		case "-p", "--parallel":
			parallel = true
		case "-h", "--help":
			printHelp()
			return
		default:
			// Treat as filter
			filters = append(filters, arg)
		}
		_ = i
	}

	runner.SetVerbose(verbose)
	runner.SetParallel(parallel)

	var err error
	if len(filters) > 0 {
		err = runner.RunSpecific(filters)
	} else {
		err = runner.RunAll()
	}

	if err != nil {
		fmt.Printf("\\n❌ Test execution failed: %v\\n", err)
		os.Exit(1)
	}

	fmt.Println("\\n🎉 All tests completed successfully!")
}

func printHelp() {
	fmt.Println("LIV Document Format Test Runner")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Usage: go run run-all-tests.go [options] [filters...]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose    Enable verbose output")
	fmt.Println("  -p, --parallel   Enable parallel test execution")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println()
	fmt.Println("Filters:")
	fmt.Println("  You can specify test suite names or tags to run specific tests:")
	fmt.Println("  - unit           Run unit tests")
	fmt.Println("  - security       Run security tests")
	fmt.Println("  - integration    Run integration tests")
	fmt.Println("  - performance    Run performance tests")
	fmt.Println("  - e2e            Run end-to-end tests")
	fmt.Println("  - fast           Run fast tests (unit)")
	fmt.Println("  - slow           Run slow tests (performance, e2e)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run run-all-tests.go                    # Run all tests")
	fmt.Println("  go run run-all-tests.go -v unit            # Run unit tests with verbose output")
	fmt.Println("  go run run-all-tests.go security           # Run security tests only")
	fmt.Println("  go run run-all-tests.go fast -p            # Run fast tests in parallel")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  GENERATE_COVERAGE=true    Generate test coverage reports")
}
