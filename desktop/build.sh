#!/bin/bash

# LIV Desktop Application Build Script
# This script builds the desktop application for multiple platforms

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DESKTOP_DIR="$PROJECT_ROOT/desktop"
BIN_DIR="$PROJECT_ROOT/bin"
DIST_DIR="$DESKTOP_DIR/dist"

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed"
        exit 1
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed"
        exit 1
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Build Go viewer executable
build_viewer() {
    log_info "Building LIV viewer executable..."
    
    cd "$PROJECT_ROOT"
    
    # Create bin directory if it doesn't exist
    mkdir -p "$BIN_DIR"
    
    # Build for current platform
    local platform=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $arch in
        x86_64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        *) log_warning "Unknown architecture: $arch, using amd64"; arch="amd64" ;;
    esac
    
    local executable="liv-viewer"
    if [[ "$platform" == "windows" || "$platform" == "mingw"* || "$platform" == "cygwin"* ]]; then
        executable="liv-viewer.exe"
    fi
    
    log_info "Building for $platform/$arch..."
    
    GOOS="$platform" GOARCH="$arch" go build -o "$BIN_DIR/$executable" cmd/viewer/main.go
    
    if [[ $? -eq 0 ]]; then
        log_success "Viewer executable built: $BIN_DIR/$executable"
    else
        log_error "Failed to build viewer executable"
        exit 1
    fi
}

# Build for multiple platforms
build_viewer_all_platforms() {
    log_info "Building LIV viewer for all platforms..."
    
    cd "$PROJECT_ROOT"
    mkdir -p "$BIN_DIR"
    
    # Platform configurations
    declare -A platforms=(
        ["windows/amd64"]="liv-viewer.exe"
        ["windows/386"]="liv-viewer-32.exe"
        ["darwin/amd64"]="liv-viewer-mac"
        ["darwin/arm64"]="liv-viewer-mac-arm64"
        ["linux/amd64"]="liv-viewer-linux"
        ["linux/arm64"]="liv-viewer-linux-arm64"
    )
    
    for platform_arch in "${!platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform_arch"
        executable="${platforms[$platform_arch]}"
        
        log_info "Building for $os/$arch..."
        
        GOOS="$os" GOARCH="$arch" go build -o "$BIN_DIR/$executable" cmd/viewer/main.go
        
        if [[ $? -eq 0 ]]; then
            log_success "Built: $BIN_DIR/$executable"
        else
            log_error "Failed to build for $os/$arch"
        fi
    done
}

# Install desktop dependencies
install_dependencies() {
    log_info "Installing desktop application dependencies..."
    
    cd "$DESKTOP_DIR"
    
    if [[ -f "package-lock.json" ]]; then
        npm ci
    else
        npm install
    fi
    
    if [[ $? -eq 0 ]]; then
        log_success "Dependencies installed"
    else
        log_error "Failed to install dependencies"
        exit 1
    fi
}

# Build desktop application
build_desktop() {
    local platform="$1"
    
    log_info "Building desktop application..."
    
    cd "$DESKTOP_DIR"
    
    # Clean previous builds
    rm -rf "$DIST_DIR"
    
    case "$platform" in
        "win"|"windows")
            npm run build:win
            ;;
        "mac"|"darwin"|"macos")
            npm run build:mac
            ;;
        "linux")
            npm run build:linux
            ;;
        "all"|"")
            npm run build
            ;;
        *)
            log_error "Unknown platform: $platform"
            log_info "Available platforms: win, mac, linux, all"
            exit 1
            ;;
    esac
    
    if [[ $? -eq 0 ]]; then
        log_success "Desktop application built successfully"
        log_info "Output directory: $DIST_DIR"
    else
        log_error "Failed to build desktop application"
        exit 1
    fi
}

# Package application
package_app() {
    log_info "Packaging desktop application..."
    
    cd "$DESKTOP_DIR"
    npm run pack
    
    if [[ $? -eq 0 ]]; then
        log_success "Application packaged successfully"
    else
        log_error "Failed to package application"
        exit 1
    fi
}

# Clean build artifacts
clean() {
    log_info "Cleaning build artifacts..."
    
    # Clean Go build cache
    cd "$PROJECT_ROOT"
    go clean -cache
    
    # Clean desktop build artifacts
    cd "$DESKTOP_DIR"
    rm -rf dist/
    rm -rf node_modules/.cache/
    
    # Clean bin directory
    rm -rf "$BIN_DIR"
    
    log_success "Build artifacts cleaned"
}

# Show help
show_help() {
    echo "LIV Desktop Application Build Script"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  build [PLATFORM]    Build desktop application for specified platform"
    echo "  viewer              Build only the Go viewer executable"
    echo "  viewer-all          Build viewer for all platforms"
    echo "  deps                Install desktop dependencies"
    echo "  package             Package application without building installer"
    echo "  clean               Clean build artifacts"
    echo "  help                Show this help message"
    echo ""
    echo "Platforms:"
    echo "  win, windows        Build for Windows"
    echo "  mac, darwin, macos  Build for macOS"
    echo "  linux               Build for Linux"
    echo "  all                 Build for all platforms (default)"
    echo ""
    echo "Examples:"
    echo "  $0 build win        Build Windows desktop application"
    echo "  $0 viewer           Build viewer executable for current platform"
    echo "  $0 viewer-all       Build viewer for all platforms"
    echo "  $0 deps             Install dependencies only"
    echo "  $0 clean            Clean all build artifacts"
}

# Main script logic
main() {
    local command="$1"
    local platform="$2"
    
    case "$command" in
        "build")
            check_prerequisites
            build_viewer
            install_dependencies
            build_desktop "$platform"
            ;;
        "viewer")
            check_prerequisites
            build_viewer
            ;;
        "viewer-all")
            check_prerequisites
            build_viewer_all_platforms
            ;;
        "deps")
            install_dependencies
            ;;
        "package")
            check_prerequisites
            package_app
            ;;
        "clean")
            clean
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"