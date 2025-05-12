#!/usr/bin/env bash
# Installation script for TatuScan client
set -euo pipefail

BIN_NAME="${1:-tatuscan}"
SERVICE_NAME="tatuscan"
SERVICE_USER="${3:-tatuscan}"

if [ "${EUID:-$(id -u)}" -eq 0 ]; then
    INSTALL_DIR="${2:-/usr/local/bin}"
    INSTALL_TYPE="system"
else
    INSTALL_DIR="${2:-$HOME/.local/bin}"
    INSTALL_TYPE="user"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$CLIENT_DIR/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_privileges() {
    if [ "$INSTALL_TYPE" = "system" ] && [ "${EUID:-$(id -u)}" -ne 0 ]; then
        log_error "System-wide installation requires root privileges (use sudo)"
        exit 1
    fi
}

install_binary() {
    local os arch platform_binary generic_binary source_path=""
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) log_error "Unsupported architecture: $arch"; exit 1 ;;
    esac

    platform_binary="$BIN_DIR/$os/$BIN_NAME"
    generic_binary="$BIN_DIR/linux/$BIN_NAME"

    if [ -f "$platform_binary" ]; then
        source_path="$platform_binary"
    elif [ -f "$generic_binary" ]; then
        source_path="$generic_binary"
    else
        log_error "Binary not found. Tried:"
        log_error "  - $platform_binary"
        log_error "  - $generic_binary"
        log_info "Run 'make client-build' (or build-all) first"
        exit 1
    fi

    log_info "Installing binary to $INSTALL_DIR/"
    log_info "Source: $source_path"
    mkdir -p "$INSTALL_DIR"
    cp "$source_path" "$INSTALL_DIR/$BIN_NAME"
    chmod +x "$INSTALL_DIR/$BIN_NAME"
}

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

install_systemd_service() {
    if [ "$(uname -s)" != "Linux" ] || [ "$INSTALL_TYPE" != "system" ]; then
        return 0
    fi

    local env_dir="/etc/tatuscan"
    local env_file="$env_dir/agent.env"
    local service_file="/etc/systemd/system/$SERVICE_NAME.service"
    local url="${TATUSCAN_URL:-}"

    mkdir -p "$env_dir"
    if [ ! -f "$env_file" ]; then
        if [ -z "$url" ]; then
            log_error "TATUSCAN_URL is required for systemd install"
            log_info "Example: sudo TATUSCAN_URL=http://api.example:8040 make -C client install"
            exit 1
        fi
        cat > "$env_file" << EOF
TATUSCAN_URL=$url
TATUSCAN_INTERVAL=${TATUSCAN_INTERVAL:-60s}
EOF
        chmod 0640 "$env_file"
        chown root:"$SERVICE_USER" "$env_file" 2>/dev/null || chmod 0644 "$env_file"
        log_info "Wrote $env_file"
    else
        log_info "Keeping existing $env_file"
    fi

    cat > "$service_file" << EOF
[Unit]
Description=TatuScan monitoring agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
EnvironmentFile=-$env_file
ExecStart=$INSTALL_DIR/$BIN_NAME -d
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_info "Systemd unit installed: $service_file"
    log_info "Enable with: sudo systemctl enable --now $SERVICE_NAME"
}

main() {
    log_info "Starting TatuScan installation ($INSTALL_TYPE)..."
    check_privileges
    install_binary
    create_service_user
    install_systemd_service
    log_info "Binary: $INSTALL_DIR/$BIN_NAME"
    if [ "$INSTALL_TYPE" = "user" ]; then
        log_info "Ensure $INSTALL_DIR is on PATH"
        log_info "Run with: TATUSCAN_URL=http://host:8040 $INSTALL_DIR/$BIN_NAME -d"
    fi
}

main "$@"
