@echo off
REM LIV Desktop Application Build Script for Windows
REM This script builds the desktop application for Windows

setlocal enabledelayedexpansion

REM Configuration
set "PROJECT_ROOT=%~dp0.."
set "DESKTOP_DIR=%PROJECT_ROOT%\desktop"
set "BIN_DIR=%PROJECT_ROOT%\bin"
set "DIST_DIR=%DESKTOP_DIR%\dist"

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

REM Check prerequisites
:check_prerequisites
call :log_info "Checking prerequisites..."

REM Check Node.js
node --version >nul 2>&1
if errorlevel 1 (
    call :log_error "Node.js is not installed"
    exit /b 1
)

REM Check npm
npm --version >nul 2>&1
if errorlevel 1 (
    call :log_error "npm is not installed"
    exit /b 1
)

REM Check Go
go version >nul 2>&1
if errorlevel 1 (
    call :log_error "Go is not installed"
    exit /b 1
)

call :log_success "Prerequisites check passed"
goto :eof

REM Build Go viewer executable
:build_viewer
call :log_info "Building LIV viewer executable..."

cd /d "%PROJECT_ROOT%"

REM Create bin directory if it doesn't exist
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

REM Build for Windows
call :log_info "Building for Windows/amd64..."

set GOOS=windows
set GOARCH=amd64
go build -o "%BIN_DIR%\liv-viewer.exe" cmd\viewer\main.go

if errorlevel 1 (
    call :log_error "Failed to build viewer executable"
    exit /b 1
) else (
    call :log_success "Viewer executable built: %BIN_DIR%\liv-viewer.exe"
)
goto :eof

REM Build for multiple platforms
:build_viewer_all_platforms
call :log_info "Building LIV viewer for all platforms..."

cd /d "%PROJECT_ROOT%"
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

REM Windows builds
call :log_info "Building for Windows/amd64..."
set GOOS=windows
set GOARCH=amd64
go build -o "%BIN_DIR%\liv-viewer.exe" cmd\viewer\main.go

call :log_info "Building for Windows/386..."
set GOOS=windows
set GOARCH=386
go build -o "%BIN_DIR%\liv-viewer-32.exe" cmd\viewer\main.go

REM macOS builds
call :log_info "Building for macOS/amd64..."
set GOOS=darwin
set GOARCH=amd64
go build -o "%BIN_DIR%\liv-viewer-mac" cmd\viewer\main.go

call :log_info "Building for macOS/arm64..."
set GOOS=darwin
set GOARCH=arm64
go build -o "%BIN_DIR%\liv-viewer-mac-arm64" cmd\viewer\main.go

REM Linux builds
call :log_info "Building for Linux/amd64..."
set GOOS=linux
set GOARCH=amd64
go build -o "%BIN_DIR%\liv-viewer-linux" cmd\viewer\main.go

call :log_info "Building for Linux/arm64..."
set GOOS=linux
set GOARCH=arm64
go build -o "%BIN_DIR%\liv-viewer-linux-arm64" cmd\viewer\main.go

call :log_success "All platform builds completed"
goto :eof

REM Install desktop dependencies
:install_dependencies
call :log_info "Installing desktop application dependencies..."

cd /d "%DESKTOP_DIR%"

if exist "package-lock.json" (
    npm ci
) else (
    npm install
)

if errorlevel 1 (
    call :log_error "Failed to install dependencies"
    exit /b 1
) else (
    call :log_success "Dependencies installed"
)
goto :eof

REM Build desktop application
:build_desktop
set "platform=%~1"

call :log_info "Building desktop application..."

cd /d "%DESKTOP_DIR%"

REM Clean previous builds
if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"

if "%platform%"=="win" (
    npm run build:win
) else if "%platform%"=="windows" (
    npm run build:win
) else if "%platform%"=="mac" (
    npm run build:mac
) else if "%platform%"=="darwin" (
    npm run build:mac
) else if "%platform%"=="macos" (
    npm run build:mac
) else if "%platform%"=="linux" (
    npm run build:linux
) else if "%platform%"=="all" (
    npm run build
) else if "%platform%"=="" (
    npm run build
) else (
    call :log_error "Unknown platform: %platform%"
    call :log_info "Available platforms: win, mac, linux, all"
    exit /b 1
)

if errorlevel 1 (
    call :log_error "Failed to build desktop application"
    exit /b 1
) else (
    call :log_success "Desktop application built successfully"
    call :log_info "Output directory: %DIST_DIR%"
)
goto :eof

REM Package application
:package_app
call :log_info "Packaging desktop application..."

cd /d "%DESKTOP_DIR%"
npm run pack

if errorlevel 1 (
    call :log_error "Failed to package application"
    exit /b 1
) else (
    call :log_success "Application packaged successfully"
)
goto :eof

REM Clean build artifacts
:clean
call :log_info "Cleaning build artifacts..."

REM Clean Go build cache
cd /d "%PROJECT_ROOT%"
go clean -cache

REM Clean desktop build artifacts
cd /d "%DESKTOP_DIR%"
if exist "dist" rmdir /s /q "dist"
if exist "node_modules\.cache" rmdir /s /q "node_modules\.cache"

REM Clean bin directory
if exist "%BIN_DIR%" rmdir /s /q "%BIN_DIR%"

call :log_success "Build artifacts cleaned"
goto :eof

REM Show help
:show_help
echo LIV Desktop Application Build Script for Windows
echo.
echo Usage: %~nx0 [COMMAND] [OPTIONS]
echo.
echo Commands:
echo   build [PLATFORM]    Build desktop application for specified platform
echo   viewer              Build only the Go viewer executable
echo   viewer-all          Build viewer for all platforms
echo   deps                Install desktop dependencies
echo   package             Package application without building installer
echo   clean               Clean build artifacts
echo   help                Show this help message
echo.
echo Platforms:
echo   win, windows        Build for Windows
echo   mac, darwin, macos  Build for macOS
echo   linux               Build for Linux
echo   all                 Build for all platforms (default)
echo.
echo Examples:
echo   %~nx0 build win        Build Windows desktop application
echo   %~nx0 viewer           Build viewer executable for Windows
echo   %~nx0 viewer-all       Build viewer for all platforms
echo   %~nx0 deps             Install dependencies only
echo   %~nx0 clean            Clean all build artifacts
goto :eof

REM Main script logic
:main
set "command=%~1"
set "platform=%~2"

if "%command%"=="build" (
    call :check_prerequisites
    call :build_viewer
    call :install_dependencies
    call :build_desktop "%platform%"
) else if "%command%"=="viewer" (
    call :check_prerequisites
    call :build_viewer
) else if "%command%"=="viewer-all" (
    call :check_prerequisites
    call :build_viewer_all_platforms
) else if "%command%"=="deps" (
    call :install_dependencies
) else if "%command%"=="package" (
    call :check_prerequisites
    call :package_app
) else if "%command%"=="clean" (
    call :clean
) else if "%command%"=="help" (
    call :show_help
) else if "%command%"=="--help" (
    call :show_help
) else if "%command%"=="-h" (
    call :show_help
) else if "%command%"=="" (
    call :show_help
) else (
    call :log_error "Unknown command: %command%"
    call :show_help
    exit /b 1
)

goto :eof

REM Run main function
call :main %*