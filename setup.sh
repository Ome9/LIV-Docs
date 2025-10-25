#!/bin/bash

# LIV Document Format - Complete Setup Script
# This script installs and configures the entire LIV Document Format system

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LIV_VERSION="1.0.0"
INSTALL_DIR="/opt/liv"
BIN_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.liv"

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        echo "linux"
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Function to detect architecture
detect_arch() {
    case $(uname -m) in
        x86_64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        armv7l) echo "arm" ;;
        *) echo "unknown" ;;
    esac
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    local missing_deps=()
    
    # Check for required tools
    if ! command_exists "curl" && ! command_exists "wget"; then
        missing_deps+=("curl or wget")
    fi
    
    if ! command_exists "unzip"; then
        missing_deps+=("unzip")
    fi
    
    if ! command_exists "tar"; then
        missing_deps+=("tar")
    fi
    
    # Check for development dependencies (optional)
    if ! command_exists "go"; then
        print_warning "Go not found - required for building from source"
    fi
    
    if ! command_exists "node"; then
        print_warning "Node.js not found - required for JavaScript SDK"
    fi
    
    if ! command_exists "python3"; then
        print_warning "Python 3 not found - required for Python SDK"
    fi
    
    if ! command_exists "rustc"; then
        print_warning "Rust not found - required for WASM modules"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_error "Please install the missing dependencies and run this script again."
        exit 1
    fi
    
    print_success "Prerequisites check completed"
}

# Function to install system dependencies
install_system_deps() {
    print_status "Installing system dependencies..."
    
    local os=$(detect_os)
    
    case $os in
        "linux")
            if command_exists "apt-get"; then
                sudo apt-get update
                sudo apt-get install -y curl wget unzip tar build-essential
            elif command_exists "yum"; then
                sudo yum install -y curl wget unzip tar gcc gcc-c++ make
            elif command_exists "pacman"; then
                sudo pacman -S --noconfirm curl wget unzip tar base-devel
            else
                print_warning "Unknown Linux distribution. Please install curl, wget, unzip, tar manually."
            fi
            ;;
        "macos")
            if command_exists "brew"; then
                brew install curl wget unzip
            else
                print_warning "Homebrew not found. Please install curl, wget, unzip manually."
            fi
            ;;
        "windows")
            print_warning "Windows detected. Please ensure you have curl, wget, unzip available."
            ;;
        *)
            print_warning "Unknown OS. Please install curl, wget, unzip, tar manually."
            ;;
    esac
}

# Function to install development dependencies
install_dev_deps() {
    print_status "Installing development dependencies..."
    
    local os=$(detect_os)
    
    # Install Go
    if ! command_exists "go"; then
        print_status "Installing Go..."
        case $os in
            "linux"|"macos")
                GO_VERSION="1.21.0"
                GO_ARCH=$(detect_arch)
                GO_OS=$os
                if [ "$os" = "macos" ]; then
                    GO_OS="darwin"
                fi
                
                curl -L "https://golang.org/dl/go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz" -o go.tar.gz
                sudo tar -C /usr/local -xzf go.tar.gz
                rm go.tar.gz
                
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
                export PATH=$PATH:/usr/local/go/bin
                ;;
        esac
    fi
    
    # Install Node.js
    if ! command_exists "node"; then
        print_status "Installing Node.js..."
        case $os in
            "linux"|"macos")
                curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
                if [ "$os" = "linux" ]; then
                    sudo apt-get install -y nodejs
                elif [ "$os" = "macos" ] && command_exists "brew"; then
                    brew install node
                fi
                ;;
        esac
    fi
    
    # Install Python 3
    if ! command_exists "python3"; then
        print_status "Installing Python 3..."
        case $os in
            "linux")
                if command_exists "apt-get"; then
                    sudo apt-get install -y python3 python3-pip
                elif command_exists "yum"; then
                    sudo yum install -y python3 python3-pip
                fi
                ;;
            "macos")
                if command_exists "brew"; then
                    brew install python3
                fi
                ;;
        esac
    fi
    
    # Install Rust
    if ! command_exists "rustc"; then
        print_status "Installing Rust..."
        curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
        source ~/.cargo/env
        rustup target add wasm32-unknown-unknown
        cargo install wasm-pack
    fi
}

# Function to download and install LIV binaries
install_binaries() {
    print_status "Installing LIV binaries..."
    
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    # Create directories
    sudo mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    
    # Download release package
    local package_name="liv-document-format-v${LIV_VERSION}-${os}-${arch}"
    local download_url="https://github.com/your-org/liv-document-format/releases/download/v${LIV_VERSION}/${package_name}.zip"
    
    print_status "Downloading $package_name..."
    
    if command_exists "curl"; then
        curl -L "$download_url" -o "/tmp/${package_name}.zip"
    elif command_exists "wget"; then
        wget "$download_url" -O "/tmp/${package_name}.zip"
    else
        print_error "Neither curl nor wget found. Cannot download package."
        exit 1
    fi
    
    # Extract package
    print_status "Extracting package..."
    cd /tmp
    unzip -q "${package_name}.zip"
    
    # Install binaries
    print_status "Installing binaries to $INSTALL_DIR..."
    sudo cp -r "${package_name}"/* "$INSTALL_DIR/"
    
    # Create symlinks
    print_status "Creating symlinks in $BIN_DIR..."
    sudo ln -sf "$INSTALL_DIR/bin/liv-cli" "$BIN_DIR/liv-cli"
    sudo ln -sf "$INSTALL_DIR/bin/liv-viewer" "$BIN_DIR/liv-viewer"
    sudo ln -sf "$INSTALL_DIR/bin/liv-builder" "$BIN_DIR/liv-builder"
    sudo ln -sf "$INSTALL_DIR/bin/permission-server" "$BIN_DIR/permission-server"
    sudo ln -sf "$INSTALL_DIR/bin/security-admin" "$BIN_DIR/security-admin"
    
    # Make binaries executable
    sudo chmod +x "$INSTALL_DIR/bin/"*
    
    # Cleanup
    rm -rf "/tmp/${package_name}" "/tmp/${package_name}.zip"
    
    print_success "Binaries installed successfully"
}

# Function to install JavaScript SDK
install_js_sdk() {
    print_status "Installing JavaScript SDK..."
    
    if command_exists "npm"; then
        # Install globally
        sudo npm install -g "$INSTALL_DIR/js/liv-document-format-${LIV_VERSION}.tgz"
        print_success "JavaScript SDK installed globally"
    else
        print_warning "npm not found. JavaScript SDK not installed."
    fi
}

# Function to install Python SDK
install_python_sdk() {
    print_status "Installing Python SDK..."
    
    if command_exists "pip3"; then
        # Install from wheel
        pip3 install --user "$INSTALL_DIR/python/dist/liv_document_format-${LIV_VERSION}-py3-none-any.whl"
        print_success "Python SDK installed for current user"
    elif command_exists "pip"; then
        pip install --user "$INSTALL_DIR/python/dist/liv_document_format-${LIV_VERSION}-py3-none-any.whl"
        print_success "Python SDK installed for current user"
    else
        print_warning "pip not found. Python SDK not installed."
    fi
}

# Function to install desktop application
install_desktop_app() {
    print_status "Installing desktop application..."
    
    local os=$(detect_os)
    
    case $os in
        "linux")
            # Install AppImage
            sudo cp "$INSTALL_DIR/liv-document-viewer.AppImage" "/usr/local/bin/"
            sudo chmod +x "/usr/local/bin/liv-document-viewer.AppImage"
            
            # Create desktop entry
            mkdir -p "$HOME/.local/share/applications"
            cat > "$HOME/.local/share/applications/liv-document-viewer.desktop" << EOF
[Desktop Entry]
Name=LIV Document Viewer
Comment=View and edit LIV documents
Exec=/usr/local/bin/liv-document-viewer.AppImage %f
Icon=$INSTALL_DIR/icons/liv-icon.png
Type=Application
Categories=Office;Viewer;
MimeType=application/x-liv-document;
EOF
            
            # Update desktop database
            if command_exists "update-desktop-database"; then
                update-desktop-database "$HOME/.local/share/applications"
            fi
            ;;
        "macos")
            # Install DMG (would need to be mounted and copied)
            print_status "Please manually install the desktop application from $INSTALL_DIR/LIV-Document-Viewer.dmg"
            ;;
        "windows")
            # Run installer (would need to be executed)
            print_status "Please manually run the installer at $INSTALL_DIR/LIV-Document-Viewer-Setup.exe"
            ;;
    esac
    
    print_success "Desktop application installation completed"
}

# Function to configure file associations
configure_file_associations() {
    print_status "Configuring file associations..."
    
    local os=$(detect_os)
    
    case $os in
        "linux")
            # Create MIME type
            mkdir -p "$HOME/.local/share/mime/packages"
            cat > "$HOME/.local/share/mime/packages/liv-document.xml" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
    <mime-type type="application/x-liv-document">
        <comment>LIV Document</comment>
        <glob pattern="*.liv"/>
        <icon name="application-x-liv-document"/>
    </mime-type>
</mime-info>
EOF
            
            # Update MIME database
            if command_exists "update-mime-database"; then
                update-mime-database "$HOME/.local/share/mime"
            fi
            ;;
        "macos")
            # macOS file associations are handled by the app bundle
            print_status "File associations will be configured when the desktop app is installed"
            ;;
        "windows")
            # Windows file associations are handled by the installer
            print_status "File associations will be configured by the installer"
            ;;
    esac
}

# Function to create configuration files
create_config() {
    print_status "Creating configuration files..."
    
    # Create main config
    cat > "$CONFIG_DIR/config.yaml" << EOF
# LIV Document Format Configuration

# Default settings
default_author: "$USER"
default_license: "MIT"
compression_enabled: true
security_level: "strict"

# Signing settings
signing:
  algorithm: "RSA-SHA256"

# Viewer settings
viewer:
  default_renderer: "webgl"
  enable_animations: true
  sandbox_mode: true

# Performance settings
performance:
  memory_limit: "512MB"
  cache_size: "100MB"
  parallel_processing: true

# Security settings
security:
  allow_network_access: false
  allow_file_system: false
  memory_limit: "64MB"
  
# Paths
paths:
  templates: "$CONFIG_DIR/templates"
  keys: "$CONFIG_DIR/keys"
  cache: "$CONFIG_DIR/cache"
EOF

    # Create directories
    mkdir -p "$CONFIG_DIR/templates"
    mkdir -p "$CONFIG_DIR/keys"
    mkdir -p "$CONFIG_DIR/cache"
    mkdir -p "$CONFIG_DIR/logs"
    
    # Copy example templates
    if [ -d "$INSTALL_DIR/examples/templates" ]; then
        cp -r "$INSTALL_DIR/examples/templates/"* "$CONFIG_DIR/templates/"
    fi
    
    print_success "Configuration files created"
}

# Function to run tests
run_tests() {
    print_status "Running system tests..."
    
    # Test CLI tools
    if command_exists "liv-cli"; then
        liv-cli --version
        print_success "CLI tools working"
    else
        print_error "CLI tools not found in PATH"
    fi
    
    # Test document creation
    echo '<html><body><h1>Test Document</h1></body></html>' > /tmp/test.html
    
    if liv-cli build --source /tmp --output /tmp/test.liv; then
        print_success "Document creation test passed"
        
        # Test validation
        if liv-cli validate /tmp/test.liv; then
            print_success "Document validation test passed"
        else
            print_warning "Document validation test failed"
        fi
        
        # Test viewing (headless)
        if liv-cli view /tmp/test.liv --headless; then
            print_success "Document viewing test passed"
        else
            print_warning "Document viewing test failed"
        fi
        
        # Cleanup
        rm -f /tmp/test.html /tmp/test.liv
    else
        print_error "Document creation test failed"
    fi
    
    # Test JavaScript SDK
    if command_exists "node" && npm list -g liv-document-format >/dev/null 2>&1; then
        print_success "JavaScript SDK available"
    else
        print_warning "JavaScript SDK not available"
    fi
    
    # Test Python SDK
    if python3 -c "import liv" >/dev/null 2>&1; then
        print_success "Python SDK available"
    else
        print_warning "Python SDK not available"
    fi
}

# Function to display installation summary
show_summary() {
    print_success "LIV Document Format installation completed!"
    echo
    echo "📦 Installation Summary:"
    echo "  • Binaries installed in: $INSTALL_DIR"
    echo "  • CLI tools available: liv-cli, liv-viewer, liv-builder"
    echo "  • Configuration directory: $CONFIG_DIR"
    echo "  • Desktop application: $([ -f "/usr/local/bin/liv-document-viewer.AppImage" ] && echo "Installed" || echo "Manual installation required")"
    echo
    echo "🚀 Quick Start:"
    echo "  # Create a document"
    echo "  echo '<html><body><h1>Hello LIV!</h1></body></html>' > hello.html"
    echo "  liv-cli build --source . --output hello.liv"
    echo
    echo "  # View a document"
    echo "  liv-cli view hello.liv"
    echo
    echo "  # Validate a document"
    echo "  liv-cli validate hello.liv"
    echo
    echo "📚 Documentation:"
    echo "  • User Guide: $INSTALL_DIR/docs/USER_GUIDE.md"
    echo "  • API Reference: $INSTALL_DIR/docs/reference/"
    echo "  • Examples: $INSTALL_DIR/examples/"
    echo
    echo "🔧 Configuration:"
    echo "  • Config file: $CONFIG_DIR/config.yaml"
    echo "  • Templates: $CONFIG_DIR/templates/"
    echo "  • Logs: $CONFIG_DIR/logs/"
    echo
    echo "For more information, visit: https://github.com/your-org/liv-document-format"
}

# Main installation function
main() {
    echo "🚀 LIV Document Format Setup Script v$LIV_VERSION"
    echo "=================================================="
    echo
    
    # Parse command line arguments
    INSTALL_DEV_DEPS=false
    BUILD_FROM_SOURCE=false
    SKIP_TESTS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dev)
                INSTALL_DEV_DEPS=true
                shift
                ;;
            --build)
                BUILD_FROM_SOURCE=true
                INSTALL_DEV_DEPS=true
                shift
                ;;
            --skip-tests)
                SKIP_TESTS=true
                shift
                ;;
            --help)
                echo "Usage: $0 [options]"
                echo
                echo "Options:"
                echo "  --dev         Install development dependencies"
                echo "  --build       Build from source (implies --dev)"
                echo "  --skip-tests  Skip running tests after installation"
                echo "  --help        Show this help message"
                echo
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root. This is not recommended."
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    # Run installation steps
    check_prerequisites
    install_system_deps
    
    if [ "$INSTALL_DEV_DEPS" = true ]; then
        install_dev_deps
    fi
    
    if [ "$BUILD_FROM_SOURCE" = true ]; then
        print_status "Building from source..."
        make install
        make build
        make install-binaries
    else
        install_binaries
    fi
    
    install_js_sdk
    install_python_sdk
    install_desktop_app
    configure_file_associations
    create_config
    
    if [ "$SKIP_TESTS" = false ]; then
        run_tests
    fi
    
    show_summary
}

# Run main function with all arguments
main "$@"