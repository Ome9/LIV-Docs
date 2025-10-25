@echo off
REM Security and Administration Test Runner for Windows
REM Runs comprehensive security tests including unit tests, integration tests, and performance tests

setlocal enabledelayedexpansion

echo 🔐 LIV Security and Administration Test Suite
echo ==============================================

REM Check if Go is available
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

echo [INFO] Go version:
go version

REM Create test results directory
set TEST_RESULTS_DIR=test-results\security
if not exist "%TEST_RESULTS_DIR%" mkdir "%TEST_RESULTS_DIR%"

echo [INFO] Starting security and administration test suite...

REM Phase 1: Core Security Tests
echo [INFO] Phase 1: Core Security Tests
echo [INFO] Running policy-manager tests...
go test -v -race -coverprofile="%TEST_RESULTS_DIR%\policy-manager-coverage.out" .\pkg\security > "%TEST_RESULTS_DIR%\policy-manager-results.txt" 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Policy manager tests failed
    type "%TEST_RESULTS_DIR%\policy-manager-results.txt" | findstr /C:"FAIL"
    exit /b 1
)
echo [SUCCESS] Policy manager tests passed

REM Generate coverage for policy manager
if exist "%TEST_RESULTS_DIR%\policy-manager-coverage.out" (
    for /f "tokens=3" %%i in ('go tool cover -func="%TEST_RESULTS_DIR%\policy-manager-coverage.out" ^| findstr "total"') do (
        echo [INFO] Policy manager coverage: %%i
    )
)

REM Phase 2: Security Administration Tests
echo [INFO] Phase 2: Security Administration Tests
echo [INFO] Running security administration tests...
go test -v -race .\pkg\security -run="TestSecurityAdministrationSuite" > "%TEST_RESULTS_DIR%\administration-tests.txt" 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Security administration tests failed
    type "%TEST_RESULTS_DIR%\administration-tests.txt" | findstr /C:"FAIL"
    exit /b 1
)
echo [SUCCESS] Security administration tests passed

REM Phase 3: Security Integration Tests
echo [INFO] Phase 3: Security Integration Tests
echo [INFO] Running security integration tests...
go test -v -race .\test\integration -run="TestSecurityIntegrationSuite" > "%TEST_RESULTS_DIR%\integration-tests.txt" 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Security integration tests failed
    type "%TEST_RESULTS_DIR%\integration-tests.txt" | findstr /C:"FAIL"
    exit /b 1
)
echo [SUCCESS] Security integration tests passed

REM Phase 4: Performance Tests
echo [INFO] Phase 4: Performance Tests
echo [INFO] Running performance tests...
go test -v -race .\pkg\security -run="TestConcurrentPolicyOperations|TestMemoryUsageUnderLoad|TestEventLogPerformanceUnderLoad" > "%TEST_RESULTS_DIR%\performance-tests.txt" 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Some performance tests may have failed (check results)
) else (
    echo [SUCCESS] Performance tests passed
)

REM Phase 5: Benchmarks
echo [INFO] Phase 5: Benchmarks
echo [INFO] Running security performance benchmarks...
go test -bench=. -benchmem -run=^$ .\pkg\security > "%TEST_RESULTS_DIR%\security-performance-benchmarks.txt" 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Benchmarks failed but continuing...
) else (
    echo [SUCCESS] Security performance benchmarks completed
)

REM Phase 6: Security Scenario Tests
echo [INFO] Phase 6: Security Scenario Tests
echo [INFO] Running security scenario tests...
go test -v -race .\pkg\security -run="TestSecurityPolicyEnforcementScenarios|TestSecurityEventCorrelation" > "%TEST_RESULTS_DIR%\scenario-tests.txt" 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Security scenario tests failed
    type "%TEST_RESULTS_DIR%\scenario-tests.txt" | findstr /C:"FAIL"
    exit /b 1
)
echo [SUCCESS] Security scenario tests passed

REM Phase 7: Generate Coverage Report
echo [INFO] Phase 7: Generating Coverage Report
if exist "%TEST_RESULTS_DIR%\*-coverage.out" (
    echo mode: set > "%TEST_RESULTS_DIR%\combined-coverage.out"
    
    REM Combine coverage files (simplified approach for Windows)
    for %%f in ("%TEST_RESULTS_DIR%\*-coverage.out") do (
        if not "%%f"=="%TEST_RESULTS_DIR%\combined-coverage.out" (
            more +1 "%%f" >> "%TEST_RESULTS_DIR%\combined-coverage.out"
        )
    )
    
    REM Generate total coverage
    for /f "tokens=3" %%i in ('go tool cover -func="%TEST_RESULTS_DIR%\combined-coverage.out" ^| findstr "total"') do (
        echo [SUCCESS] Total security test coverage: %%i
    )
    
    REM Generate HTML coverage report
    go tool cover -html="%TEST_RESULTS_DIR%\combined-coverage.out" -o "%TEST_RESULTS_DIR%\coverage-report.html"
    echo [INFO] HTML coverage report generated: %TEST_RESULTS_DIR%\coverage-report.html
)

REM Phase 8: Test Summary
echo [INFO] Phase 8: Test Summary
echo.
echo 📊 Test Results Summary
echo =======================

set total_suites=0
for %%f in ("%TEST_RESULTS_DIR%\*-results.txt" "%TEST_RESULTS_DIR%\*-tests.txt") do (
    if exist "%%f" (
        set /a total_suites+=1
        set "filename=%%~nf"
        set "test_name=!filename:-results=!"
        set "test_name=!test_name:-tests=!"
        
        findstr /C:"PASS" "%%f" >nul 2>nul
        if !errorlevel! equ 0 (
            set "status=✅ PASSED"
        ) else (
            findstr /C:"FAIL" "%%f" >nul 2>nul
            if !errorlevel! equ 0 (
                set "status=❌ FAILED"
            ) else (
                set "status=⚠️  UNKNOWN"
            )
        )
        
        REM Count tests in file
        for /f %%c in ('findstr /C:"=== RUN" "%%f" 2^>nul ^| find /c /v ""') do set test_count=%%c
        if not defined test_count set test_count=0
        
        echo !test_name! - !status! (!test_count! tests)
    )
)

echo.
echo 📈 Overall Statistics
echo ====================
echo Total test suites: !total_suites!
echo Test results directory: %TEST_RESULTS_DIR%

if exist "%TEST_RESULTS_DIR%\coverage-report.html" (
    echo Coverage report: %TEST_RESULTS_DIR%\coverage-report.html
)

REM Phase 9: Security Test Validation
echo [INFO] Phase 9: Security Test Validation

set critical_tests=TestSecurityPolicyEnforcement TestPermissionInheritanceEnforcement TestSecurityEventHandling TestAuditLogging TestWASMSecurityContextIntegration TestSignatureAndTrustChainIntegration TestErrorHandlingIntegration TestComplianceAndAuditIntegration

set missing_count=0
for %%t in (%critical_tests%) do (
    findstr /C:"%%t" "%TEST_RESULTS_DIR%\*.txt" >nul 2>nul
    if !errorlevel! neq 0 (
        if !missing_count! equ 0 (
            echo [WARNING] Missing critical security tests:
        )
        echo   - %%t
        set /a missing_count+=1
    )
)

if !missing_count! equ 0 (
    echo [SUCCESS] All critical security tests are covered
)

REM Phase 10: Performance Validation
echo [INFO] Phase 10: Performance Validation

if exist "%TEST_RESULTS_DIR%\security-performance-benchmarks.txt" (
    findstr /C:"BenchmarkPermissionEvaluation" "%TEST_RESULTS_DIR%\security-performance-benchmarks.txt" >nul 2>nul
    if !errorlevel! equ 0 (
        echo [SUCCESS] Permission evaluation benchmarks completed
    )
    
    findstr /C:"BenchmarkPolicyCreation" "%TEST_RESULTS_DIR%\security-performance-benchmarks.txt" >nul 2>nul
    if !errorlevel! equ 0 (
        echo [SUCCESS] Policy creation benchmarks completed
    )
    
    findstr /C:"BenchmarkResourceMonitoring" "%TEST_RESULTS_DIR%\security-performance-benchmarks.txt" >nul 2>nul
    if !errorlevel! equ 0 (
        echo [SUCCESS] Resource monitoring benchmarks completed
    )
)

echo.
echo [SUCCESS] 🎉 Security and Administration Test Suite Completed!
echo.
echo Next steps:
echo 1. Review test results in %TEST_RESULTS_DIR%\
echo 2. Check coverage report: %TEST_RESULTS_DIR%\coverage-report.html
echo 3. Address any failing tests or performance issues
echo 4. Update security documentation if needed
echo.

endlocal
exit /b 0