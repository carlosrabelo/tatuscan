#!/bin/bash
# Installation script for TatuScan

set -e

# Variables
BIN_NAME="${1:-tatuscan}"
SERVICE_NAME="tatuscan"
SERVICE_USER="${3:-tatuscan}"

# Detect installation directory based on privileges
if [ "$EUID" -eq 0 ]; then
    # Root installation - system-wide
    INSTALL_DIR="${2:-/usr/local/bin}"
    INSTALL_TYPE="system"
else
    # User installation - local user
    INSTALL_DIR="${2:-$HOME/.local/bin}"
    INSTALL_TYPE="user"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check privileges for system-wide installation
check_privileges() {
    if [ "$INSTALL_TYPE" = "system" ] && [ "$EUID" -ne 0 ]; then
        log_error "System-wide installation requires root privileges (use sudo)"
        log_info "For user installation, run as regular user (installs to \$HOME/.local/bin)"
        exit 1
    fi
}

# Detect current platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac

    case $os in
        linux) echo "linux-$arch" ;;
        darwin) echo "darwin-$arch" ;;
        *) log_error "Unsupported OS: $os"; exit 1 ;;
    esac
}

# Install binary
install_binary() {
    local platform=$(detect_platform)
    local arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
    esac

    # Try to find binary in order of preference:
    # 1. Local build (tatuscan or tatuscan.exe)
    # 2. Platform-specific build (tatuscan-platform-arch)
    local local_binary="./bin/$BIN_NAME"
    if [ "$platform" = "windows" ]; then
        local_binary="./bin/$BIN_NAME.exe"
    fi

    local platform_binary="./bin/$BIN_NAME-$platform-$arch"
    if [ "$platform" = "windows" ]; then
        platform_binary="./bin/$BIN_NAME-$platform-$arch.exe"
    fi

    local source_path=""
    local binary_type=""

    if [ -f "$local_binary" ]; then
        source_path="$local_binary"
        binary_type="local"
    elif [ -f "$platform_binary" ]; then
        source_path="$platform_binary"
        binary_type="platform-specific"
    else
        log_error "Binary not found. Tried:"
        log_error "  - $local_binary (from 'make build')"
        log_error "  - $platform_binary (from 'make build-multi')"
        log_info "Run 'make build' or 'make build-multi' first"
        exit 1
    fi

    log_info "Installing $binary_type binary to $INSTALL_DIR/"
    log_info "Source: $source_path"
    mkdir -p "$INSTALL_DIR"
    cp "$source_path" "$INSTALL_DIR/$BIN_NAME"
    chmod +x "$INSTALL_DIR/$BIN_NAME"

    log_info "Binary installed successfully"
}

# Create service user (Linux system installation only)
create_service_user() {
    if [ "$(uname -s)" != "Linux" ] || [ "$INSTALL_TYPE" != "system" ]; then
        return 0
    fi

    if id "$SERVICE_USER" &>/dev/null; then
        log_info "User $SERVICE_USER already exists"
    else
        log_info "Creating service user: $SERVICE_USER"
        useradd --system --shell /bin/false --home-dir /var/lib/tatuscan \
                --create-home "$SERVICE_USER"
    fi
}

# Install systemd service (Linux system installation only)
install_systemd_service() {
    if [ "$(uname -s)" != "Linux" ]; then
        log_warn "Systemd service installation skipped (not Linux)"
        return 0
    fi

    if [ "$INSTALL_TYPE" != "system" ]; then
        log_warn "Systemd service installation skipped (user installation)"
        log_info "For system service, run with sudo"
        return 0
    fi

    local service_file="/etc/systemd/system/$SERVICE_NAME.service"

    log_info "Creating systemd service: $service_file"
    cat > "$service_file" << EOF
[Unit]
Description=TatuScan monitoring agent
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=$INSTALL_DIR/$BIN_NAME -d
Restart=always
RestartSec=5
Environment=TATUSCAN_URL=
Environment=TATUSCAN_INTERVAL=60s

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_info "Systemd service installed (not enabled)"
    log_info "To enable and start: sudo systemctl enable --now $SERVICE_NAME"
    log_warn "Remember to configure TATUSCAN_URL environment variable"
}

# Main installation
main() {
    log_info "Starting TatuScan installation ($INSTALL_TYPE)..."
    log_info "Install directory: $INSTALL_DIR"

    check_privileges
    install_binary
    create_service_user
    install_systemd_service

    log_info "Installation completed successfully!"
    log_info "Binary location: $INSTALL_DIR/$BIN_NAME"

    if [ "$INSTALL_TYPE" = "user" ]; then
        log_warn "User installation completed"
        log_info "Make sure $INSTALL_DIR is in your PATH"
        log_info "Add to ~/.bashrc: export PATH=\"\$HOME/.local/bin:\$PATH\""
    elif [ "$(uname -s)" = "Linux" ] && [ "$INSTALL_TYPE" = "system" ]; then
        log_info "Service user: $SERVICE_USER"
        log_info "Service file: /etc/systemd/system/$SERVICE_NAME.service"
        echo ""
        log_warn "Next steps:"
        echo "  1. Configure TATUSCAN_URL environment variable"
        echo "  2. sudo systemctl enable --now $SERVICE_NAME"
        echo "  3. sudo systemctl status $SERVICE_NAME"
    fi
}

# Run main function
main "$@"