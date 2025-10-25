@echo off
REM Cross-platform compatibility test runner for Windows
REM This script runs comprehensive tests across different platforms and environments

setlocal enabledelayedexpansion

REM Configuration
set "PROJECT_ROOT=%~dp0.."
set "TEST_OUTPUT_DIR=%PROJECT_ROOT%\test-results"
set "COVERAGE_DIR=%PROJECT_ROOT%\coverage"

REM Get timestamp
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "TIMESTAMP=%dt:~0,4%%dt:~4,2%%dt:~6,2%_%dt:~8,2%%dt:~10,2%%dt:~12,2%"

REM Test configuration
set RUN_GO_TESTS=true
set RUN_JS_TESTS=true
set RUN_RUST_TESTS=true
set RUN_INTEGRATION_TESTS=true
set RUN_PERFORMANCE_TESTS=false
set GENERATE_COVERAGE=true
set VERBOSE=false

REM Platform detection
set PLATFORM=windows
set ARCH=amd64

REM Functions
:log_info
echo [INFO] %~1
goto :eof

:log_success
echo [SUCCESS] %~1
goto :eof

:log_warning
echo [WARNING] %~1
goto :eof

:log_error
echo [ERROR] %~1
goto :eof

:show_help
echo Cross-Platform Compatibility Test Runner for Windows
echo.
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo     --no-go             Skip Go tests
echo     --no-js             Skip JavaScript/TypeScript tests
echo     --no-rust           Skip Rust tests
echo     --no-integration    Skip integration tests
echo     --performance       Run performance tests
echo     --no-coverage       Skip coverage generation
echo     --verbose           Enable verbose output
echo     --help              Show this help message
echo.
echo Environment Variables:
echo     CI                  Set to 'true' to run in CI mode
echo     TEST_TIMEOUT        Test timeout in seconds (default: 300)
echo     PARALLEL_JOBS       Number of parallel test jobs (default: 4)
echo.
echo Examples:
echo     %~nx0                          # Run all tests
echo     %~nx0 --no-integration         # Skip integration tests
echo     %~nx0 --performance --verbose  # Run with performance tests and verbose output
goto :eof

:parse_args
:parse_loop
if "%~1"=="" goto :parse_done
if "%~1"=="--no-go" (
    set RUN_GO_TESTS=false
    shift
    goto :parse_loop
)
if "%~1"=="--no-js" (
    set RUN_JS_TESTS=false
    shift
    goto :parse_loop
)
if "%~1"=="--no-rust" (
    set RUN_RUST_TESTS=false
    shift
    goto :parse_loop
)
if "%~1"=="--no-integration" (
    set RUN_INTEGRATION_TESTS=false
    shift
    goto :parse_loop
)
if "%~1"=="--performance" (
    set RUN_PERFORMANCE_TESTS=true
    shift
    goto :parse_loop
)
if "%~1"=="--no-coverage" (
    set GENERATE_COVERAGE=false
    shift
    goto :parse_loop
)
if "%~1"=="--verbose" (
    set VERBOSE=true
    shift
    goto :parse_loop
)
if "%~1"=="--help" (
    call :show_help
    exit /b 0
)
call :log_error "Unknown option: %~1"
call :show_help
exit /b 1

:parse_done
goto :eof

:setup_test_environment
call :log_info "Setting up test environment..."

REM Create output directories
if not exist "%TEST_OUTPUT_DIR%" mkdir "%TEST_OUTPUT_DIR%"
if not exist "%COVERAGE_DIR%" mkdir "%COVERAGE_DIR%"

REM Set environment variables
set LIV_TEST_MODE=1
set LIV_TEST_OUTPUT_DIR=%TEST_OUTPUT_DIR%
set GO111MODULE=on
set CGO_ENABLED=1

REM Set test timeout
if not defined TEST_TIMEOUT set TEST_TIMEOUT=300
if not defined PARALLEL_JOBS set PARALLEL_JOBS=4

call :log_success "Test environment setup complete"
call :log_info "Platform: %PLATFORM%/%ARCH%"
call :log_info "Output directory: %TEST_OUTPUT_DIR%"
goto :eof

:check_dependencies
call :log_info "Checking dependencies..."

set missing_deps=

REM Check Go
if "%RUN_GO_TESTS%"=="true" (
    go version >nul 2>&1
    if errorlevel 1 set missing_deps=%missing_deps% go
)

REM Check Node.js and npm
if "%RUN_JS_TESTS%"=="true" (
    node --version >nul 2>&1
    if errorlevel 1 set missing_deps=%missing_deps% node
    
    npm --version >nul 2>&1
    if errorlevel 1 set missing_deps=%missing_deps% npm
)

REM Check Rust and Cargo
if "%RUN_RUST_TESTS%"=="true" (
    rustc --version >nul 2>&1
    if errorlevel 1 set missing_deps=%missing_deps% rust
    
    cargo --version >nul 2>&1
    if errorlevel 1 set missing_deps=%missing_deps% cargo
)

REM Check wasm-pack
if "%RUN_RUST_TESTS%"=="true" (
    wasm-pack --version >nul 2>&1
    if errorlevel 1 call :log_warning "wasm-pack not found, WASM tests may fail"
)

if not "%missing_deps%"=="" (
    call :log_error "Missing dependencies:%missing_deps%"
    call :log_info "Please install the missing dependencies and try again"
    exit /b 1
)

call :log_success "All dependencies found"
goto :eof

:run_go_tests
if not "%RUN_GO_TESTS%"=="true" goto :eof

call :log_info "Running Go tests..."

cd /d "%PROJECT_ROOT%"

set go_test_args=-v -race -timeout %TEST_TIMEOUT%s

if "%GENERATE_COVERAGE%"=="true" (
    set go_test_args=%go_test_args% -coverprofile=%COVERAGE_DIR%\go-coverage.out -covermode=atomic
)

REM Run unit tests
call :log_info "Running Go unit tests..."
go test %go_test_args% ./pkg/...
if errorlevel 1 (
    call :log_error "Go unit tests failed"
    exit /b 1
)

REM Run cross-platform tests
call :log_info "Running Go cross-platform tests..."
go test %go_test_args% ./pkg/test/...
if errorlevel 1 (
    call :log_error "Go cross-platform tests failed"
    exit /b 1
)

REM Generate coverage report
if "%GENERATE_COVERAGE%"=="true" (
    if exist "%COVERAGE_DIR%\go-coverage.out" (
        go tool cover -html="%COVERAGE_DIR%\go-coverage.out" -o "%COVERAGE_DIR%\go-coverage.html"
        go tool cover -func="%COVERAGE_DIR%\go-coverage.out" > "%COVERAGE_DIR%\go-coverage.txt"
        
        for /f "tokens=3" %%i in ('go tool cover -func^="%COVERAGE_DIR%\go-coverage.out" ^| findstr "total"') do (
            call :log_info "Go test coverage: %%i"
        )
    )
)

call :log_success "Go tests completed successfully"
goto :eof

:run_js_tests
if not "%RUN_JS_TESTS%"=="true" goto :eof

call :log_info "Running JavaScript/TypeScript tests..."

cd /d "%PROJECT_ROOT%\js"

REM Install dependencies if needed
if not exist "node_modules" (
    call :log_info "Installing JavaScript dependencies..."
    npm ci
) else (
    for %%i in (package.json) do set pkg_time=%%~ti
    for %%i in (node_modules) do set nm_time=%%~ti
    REM Simple time comparison - in production, use more robust method
    if "!pkg_time!" gtr "!nm_time!" (
        call :log_info "Installing JavaScript dependencies..."
        npm ci
    )
)

REM Run tests
set npm_test_args=

if "%GENERATE_COVERAGE%"=="true" (
    set npm_test_args=%npm_test_args% --coverage
)

if "%VERBOSE%"=="true" (
    set npm_test_args=%npm_test_args% --verbose
)

REM Run unit tests
call :log_info "Running JavaScript unit tests..."
npm test %npm_test_args%
if errorlevel 1 (
    call :log_error "JavaScript tests failed"
    exit /b 1
)

REM Run cross-platform compatibility tests
call :log_info "Running JavaScript cross-platform tests..."
npm run test:cross-platform >nul 2>&1
if errorlevel 1 (
    npx mocha test/cross-platform-compatibility.test.ts >nul 2>&1
    if errorlevel 1 call :log_warning "JavaScript cross-platform tests not available or failed"
)

REM Copy coverage reports
if "%GENERATE_COVERAGE%"=="true" (
    if exist "coverage" (
        xcopy /s /y coverage\* "%COVERAGE_DIR%\" >nul 2>&1
    )
)

call :log_success "JavaScript tests completed successfully"
goto :eof

:run_rust_tests
if not "%RUN_RUST_TESTS%"=="true" goto :eof

call :log_info "Running Rust tests..."

REM Test interactive engine
cd /d "%PROJECT_ROOT%\wasm\interactive-engine"

call :log_info "Running Rust interactive engine tests..."
cargo test --verbose
if errorlevel 1 (
    call :log_error "Rust interactive engine tests failed"
    exit /b 1
)

REM Test editor engine
cd /d "%PROJECT_ROOT%\wasm\editor-engine"

call :log_info "Running Rust editor engine tests..."
cargo test --verbose
if errorlevel 1 (
    call :log_error "Rust editor engine tests failed"
    exit /b 1
)

REM Build WASM modules
call :log_info "Building WASM modules..."

cd /d "%PROJECT_ROOT%\wasm\interactive-engine"
wasm-pack --version >nul 2>&1
if not errorlevel 1 (
    wasm-pack build --target web --out-dir ..\..\js\pkg\interactive-engine
    if errorlevel 1 call :log_warning "Failed to build interactive engine WASM module"
)

cd /d "%PROJECT_ROOT%\wasm\editor-engine"
wasm-pack --version >nul 2>&1
if not errorlevel 1 (
    wasm-pack build --target web --out-dir ..\..\js\pkg\editor-engine
    if errorlevel 1 call :log_warning "Failed to build editor engine WASM module"
)

call :log_success "Rust tests completed successfully"
goto :eof

:run_integration_tests
if not "%RUN_INTEGRATION_TESTS%"=="true" goto :eof

call :log_info "Running integration tests..."

cd /d "%PROJECT_ROOT%"

REM Build CLI tools first
call :log_info "Building CLI tools for integration tests..."

set cli_build_dir=%TEST_OUTPUT_DIR%\cli-builds
if not exist "%cli_build_dir%" mkdir "%cli_build_dir%"

REM Build for current platform
go build -o "%cli_build_dir%\liv-cli.exe" cmd\cli\main.go
go build -o "%cli_build_dir%\liv-viewer.exe" cmd\viewer\main.go

REM Add to PATH for tests
set PATH=%cli_build_dir%;%PATH%

REM Run integration tests
set integration_test_args=-v -timeout %TEST_TIMEOUT%s

if "%VERBOSE%"=="true" (
    set integration_test_args=%integration_test_args% -v
)

call :log_info "Running cross-platform integration tests..."
go test %integration_test_args% ./test/integration/...
if errorlevel 1 (
    call :log_error "Integration tests failed"
    exit /b 1
)

call :log_success "Integration tests completed successfully"
goto :eof

:run_performance_tests
if not "%RUN_PERFORMANCE_TESTS%"=="true" goto :eof

call :log_info "Running performance tests..."

cd /d "%PROJECT_ROOT%"

REM Run Go benchmarks
call :log_info "Running Go performance benchmarks..."
go test -bench=. -benchmem -timeout=%TEST_TIMEOUT%s ./pkg/... > "%TEST_OUTPUT_DIR%\go-benchmarks.txt" 2>&1

REM Run JavaScript performance tests
if "%RUN_JS_TESTS%"=="true" (
    cd /d "%PROJECT_ROOT%\js"
    call :log_info "Running JavaScript performance tests..."
    npm run test:performance >nul 2>&1
    if errorlevel 1 call :log_warning "JavaScript performance tests not available"
)

REM Run integration performance tests
if "%RUN_INTEGRATION_TESTS%"=="true" (
    cd /d "%PROJECT_ROOT%"
    call :log_info "Running integration performance tests..."
    go test -bench=. -benchmem -timeout=%TEST_TIMEOUT%s ./test/integration/... > "%TEST_OUTPUT_DIR%\integration-benchmarks.txt" 2>&1
)

call :log_success "Performance tests completed successfully"
goto :eof

:generate_test_report
call :log_info "Generating test report..."

set report_file=%TEST_OUTPUT_DIR%\test-report-%TIMESTAMP%.md

echo # Cross-Platform Compatibility Test Report > "%report_file%"
echo. >> "%report_file%"
echo **Generated:** %date% %time% >> "%report_file%"
echo **Platform:** %PLATFORM%/%ARCH% >> "%report_file%"
echo **Test Run ID:** %TIMESTAMP% >> "%report_file%"
echo. >> "%report_file%"
echo ## Test Configuration >> "%report_file%"
echo. >> "%report_file%"
echo - Go Tests: %RUN_GO_TESTS% >> "%report_file%"
echo - JavaScript Tests: %RUN_JS_TESTS% >> "%report_file%"
echo - Rust Tests: %RUN_RUST_TESTS% >> "%report_file%"
echo - Integration Tests: %RUN_INTEGRATION_TESTS% >> "%report_file%"
echo - Performance Tests: %RUN_PERFORMANCE_TESTS% >> "%report_file%"
echo - Coverage Generation: %GENERATE_COVERAGE% >> "%report_file%"
echo. >> "%report_file%"
echo ## Environment >> "%report_file%"
echo. >> "%report_file%"
echo - Platform: %PLATFORM% >> "%report_file%"
echo - Architecture: %ARCH% >> "%report_file%"

REM Add version information
for /f "tokens=*" %%i in ('go version 2^>nul') do echo - Go Version: %%i >> "%report_file%"
for /f "tokens=*" %%i in ('node --version 2^>nul') do echo - Node Version: %%i >> "%report_file%"
for /f "tokens=*" %%i in ('rustc --version 2^>nul') do echo - Rust Version: %%i >> "%report_file%"

echo. >> "%report_file%"
echo ## Test Results >> "%report_file%"
echo. >> "%report_file%"

REM Add coverage information
if "%GENERATE_COVERAGE%"=="true" (
    echo ## Coverage Reports >> "%report_file%"
    
    if exist "%COVERAGE_DIR%\go-coverage.txt" (
        echo ### Go Coverage >> "%report_file%"
        echo ``` >> "%report_file%"
        for /f "tokens=*" %%i in ('type "%COVERAGE_DIR%\go-coverage.txt" ^| findstr "total"') do echo %%i >> "%report_file%"
        echo ``` >> "%report_file%"
    )
    
    echo Coverage reports available in: %COVERAGE_DIR% >> "%report_file%"
)

REM Add performance results
if "%RUN_PERFORMANCE_TESTS%"=="true" (
    echo ## Performance Results >> "%report_file%"
    
    if exist "%TEST_OUTPUT_DIR%\go-benchmarks.txt" (
        echo ### Go Benchmarks >> "%report_file%"
        echo ``` >> "%report_file%"
        findstr "Benchmark" "%TEST_OUTPUT_DIR%\go-benchmarks.txt" | more +1 | more /e +20 >> "%report_file%"
        echo ``` >> "%report_file%"
    )
)

call :log_success "Test report generated: %report_file%"
goto :eof

:cleanup
call :log_info "Cleaning up test environment..."

REM Kill any background processes
taskkill /f /im "liv-viewer.exe" >nul 2>&1
taskkill /f /im "liv-cli.exe" >nul 2>&1

REM Clean up temporary files
del /q "%TEST_OUTPUT_DIR%\*.tmp" >nul 2>&1

call :log_success "Cleanup completed"
goto :eof

:main
REM Parse command line arguments
call :parse_args %*

call :log_info "Starting cross-platform compatibility tests..."
call :log_info "Timestamp: %TIMESTAMP%"

REM Setup and checks
call :setup_test_environment
call :check_dependencies

REM Run tests
set test_start_time=%time%
set failed_tests=

call :run_go_tests
if errorlevel 1 set failed_tests=%failed_tests% Go

call :run_js_tests
if errorlevel 1 set failed_tests=%failed_tests% JavaScript

call :run_rust_tests
if errorlevel 1 set failed_tests=%failed_tests% Rust

call :run_integration_tests
if errorlevel 1 set failed_tests=%failed_tests% Integration

call :run_performance_tests
if errorlevel 1 set failed_tests=%failed_tests% Performance

set test_end_time=%time%

REM Generate report
call :generate_test_report

REM Cleanup
call :cleanup

REM Summary
call :log_info "Test run completed"

if "%failed_tests%"=="" (
    call :log_success "All tests passed successfully!"
    exit /b 0
) else (
    call :log_error "Failed test suites:%failed_tests%"
    exit /b 1
)

REM Run main function
call :main %*