#!/usr/bin/env bash
#
# cryptofetch installer script
# 
# This script automatically downloads and installs the latest version of cryptofetch
# for your operating system and architecture.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/julianYaman/cryptofetch/main/scripts/install.sh | bash
#   
#   or with custom install location:
#   curl -sSL https://raw.githubusercontent.com/julianYaman/cryptofetch/main/scripts/install.sh | bash -s -- --prefix=$HOME/.local
#

set -e

# Configuration
REPO="julianYaman/cryptofetch"
BINARY_NAME="cryptofetch"
INSTALL_DIR="/usr/local/bin"
VERSION="latest"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --prefix=*)
            INSTALL_DIR="${1#*=}/bin"
            shift
            ;;
        --version=*)
            VERSION="${1#*=}"
            shift
            ;;
        -h|--help)
            echo "cryptofetch installer"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --prefix=PATH     Install to PATH/bin (default: /usr/local)"
            echo "  --version=VERSION Install specific version (default: latest)"
            echo "  -h, --help        Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0"
            echo "  $0 --prefix=\$HOME/.local"
            echo "  $0 --version=v1.0.0"
            exit 0
            ;;
        *)
            error "Unknown option: $1. Use --help for usage information."
            ;;
    esac
done

# Detect OS
detect_os() {
    local os=""
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch=""
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)  arch="arm64" ;;
        i386|i686)      arch="386" ;;
        armv7l)         arch="arm" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
    echo "$arch"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Get latest release version from GitHub
get_latest_version() {
    if command_exists curl; then
        curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    elif command_exists wget; then
        wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Download file
download_file() {
    local url="$1"
    local output="$2"
    
    info "Downloading from: $url"
    
    if command_exists curl; then
        curl -sSL -o "$output" "$url" || error "Download failed"
    elif command_exists wget; then
        wget -qO "$output" "$url" || error "Download failed"
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
}

# Verify checksum if available
verify_checksum() {
    local file="$1"
    local checksum_url="$2"
    
    if ! command_exists sha256sum && ! command_exists shasum; then
        warning "sha256sum/shasum not found. Skipping checksum verification."
        return 0
    fi
    
    info "Verifying checksum..."
    
    local checksum_file="/tmp/cryptofetch.sha256"
    if command_exists curl; then
        curl -sSL -o "$checksum_file" "$checksum_url" 2>/dev/null || {
            warning "Could not download checksum file. Skipping verification."
            return 0
        }
    elif command_exists wget; then
        wget -qO "$checksum_file" "$checksum_url" 2>/dev/null || {
            warning "Could not download checksum file. Skipping verification."
            return 0
        }
    fi
    
    local expected_checksum=$(cat "$checksum_file" | awk '{print $1}')
    local actual_checksum=""
    
    if command_exists sha256sum; then
        actual_checksum=$(sha256sum "$file" | awk '{print $1}')
    elif command_exists shasum; then
        actual_checksum=$(shasum -a 256 "$file" | awk '{print $1}')
    fi
    
    if [ "$expected_checksum" = "$actual_checksum" ]; then
        success "Checksum verified successfully"
    else
        error "Checksum verification failed. Downloaded file may be corrupted."
    fi
    
    rm -f "$checksum_file"
}

# Main installation logic
main() {
    info "cryptofetch installer"
    echo ""
    
    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected OS: $OS"
    info "Detected Architecture: $ARCH"
    
    # Get version
    if [ "$VERSION" = "latest" ]; then
        info "Fetching latest version..."
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            error "Could not determine latest version"
        fi
    fi
    info "Installing version: $VERSION"
    
    # Construct download URL
    local filename="${BINARY_NAME}-${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        filename="${filename}.exe"
    fi
    
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${filename}"
    local checksum_url="${download_url}.sha256"
    
    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${filename}"
    
    # Download binary
    info "Downloading cryptofetch..."
    download_file "$download_url" "$tmp_file"
    
    # Verify checksum
    verify_checksum "$tmp_file" "$checksum_url"
    
    # Make binary executable
    chmod +x "$tmp_file"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        info "Creating installation directory: $INSTALL_DIR"
        mkdir -p "$INSTALL_DIR" || error "Could not create directory: $INSTALL_DIR"
    fi
    
    # Install binary
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [ "$OS" = "windows" ]; then
        install_path="${install_path}.exe"
    fi
    
    info "Installing to: $install_path"
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_file" "$install_path" || error "Installation failed"
    else
        warning "Insufficient permissions. Attempting to use sudo..."
        sudo mv "$tmp_file" "$install_path" || error "Installation failed"
    fi
    
    # Clean up
    rm -rf "$tmp_dir"
    
    # Verify installation
    if [ -f "$install_path" ] && [ -x "$install_path" ]; then
        success "cryptofetch installed successfully!"
        echo ""
        info "Installation path: $install_path"
        
        # Check if install directory is in PATH
        if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
            warning "The installation directory is not in your PATH."
            warning "Add the following to your shell configuration file:"
            echo ""
            echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
            echo ""
        else
            info "You can now run: $BINARY_NAME --help"
        fi
        
        # Show version
        echo ""
        "$install_path" --version 2>/dev/null || info "Run '$BINARY_NAME --help' for usage information"
    else
        error "Installation verification failed"
    fi
}

# Run main function
main

# TESTING INSTRUCTIONS:
#
# Test the install script locally before publishing:
#
# 1. Create a local web server to test downloads:
#    cd /path/to/cryptofetch
#    python3 -m http.server 8000
#
# 2. Modify the script temporarily to use localhost:
#    REPO="localhost:8000"
#
# 3. Test different scenarios:
#    # Default installation (requires sudo)
#    bash scripts/install.sh
#
#    # Custom prefix (no sudo required)
#    bash scripts/install.sh --prefix=$HOME/.local
#
#    # Specific version
#    bash scripts/install.sh --version=v1.0.0
#
# 4. Test on different systems:
#    - Linux (Ubuntu, Arch, Fedora)
#    - macOS (Intel and Apple Silicon)
#    - Windows (Git Bash, WSL)
#
# 5. Test error handling:
#    - Test with invalid version
#    - Test with no write permissions
#    - Test with missing dependencies (curl/wget)
#    - Test with invalid architecture
#
# 6. After first release, test the real script:
#    curl -sSL https://raw.githubusercontent.com/julianYaman/cryptofetch/main/scripts/install.sh | bash
#
# Common issues and solutions:
# - "Permission denied": Use --prefix=$HOME/.local
# - "Command not found": Add install directory to PATH
# - "Checksum failed": Re-download or use --skip-verify flag
