#!/usr/bin/env bash
# Uninstallation script for TatuScan client

set -e

# Variables
BIN_NAME="${1:-tatuscan}"
INSTALL_DIR="${2:-/usr/local/bin}"
SERVICE_NAME="tatuscan"
SERVICE_USER="${3:-tatuscan}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

remove_systemd_service() {
    if [ "$(uname -s)" != "Linux" ]; then
        log_info "Systemd service removal skipped (not Linux)"
        return 0
    fi

    local service_file="/etc/systemd/system/$SERVICE_NAME.service"
    if [ -f "$service_file" ]; then
        log_info "Stopping and disabling $SERVICE_NAME service..."
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            systemctl stop "$SERVICE_NAME"
            log_info "Service stopped"
        fi
        if systemctl is-enabled --quiet "$SERVICE_NAME" 2>/dev/null; then
            systemctl disable "$SERVICE_NAME"
            log_info "Service disabled"
        fi
        rm -f "$service_file"
        systemctl daemon-reload
        log_info "Service file removed"
    else
        log_info "No systemd service found"
    fi
}

remove_service_user() {
    if [ "$(uname -s)" != "Linux" ]; then
        return 0
    fi
    if id "$SERVICE_USER" &>/dev/null; then
        log_info "Removing service user: $SERVICE_USER"
        userdel --remove "$SERVICE_USER" 2>/dev/null || {
            log_warn "Could not remove user home directory, removing user only"
            userdel "$SERVICE_USER" 2>/dev/null || log_warn "Could not remove user $SERVICE_USER"
        }
        if getent group "$SERVICE_USER" &>/dev/null; then
            groupdel "$SERVICE_USER" 2>/dev/null || log_warn "Could not remove group $SERVICE_USER"
        fi
    else
        log_info "Service user $SERVICE_USER does not exist"
    fi
}

remove_binary() {
    local binary_path="$INSTALL_DIR/$BIN_NAME"
    if [ -f "$binary_path" ]; then
        log_info "Removing binary: $binary_path"
        rm -f "$binary_path"
    else
        log_info "Binary not found: $binary_path"
    fi
}

remove_data() {
    local data_dirs=(
        "/var/lib/tatuscan"
        "/var/log/tatuscan"
        "/etc/tatuscan"
    )
    for dir in "${data_dirs[@]}"; do
        if [ -d "$dir" ]; then
            read -p "Remove data directory $dir? [y/N]: " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -rf "$dir"
                log_info "Removed: $dir"
            else
                log_info "Kept: $dir"
            fi
        fi
    done
}

main() {
    log_info "Starting TatuScan uninstallation..."
    check_root
    remove_systemd_service
    remove_binary
    remove_service_user
    remove_data
    log_info "Uninstall completed"
    if command -v "$BIN_NAME" &>/dev/null; then
        log_warn "$BIN_NAME is still available in PATH"
        log_warn "Location: $(which $BIN_NAME)"
        log_warn "You may need to remove it manually"
    fi
}

main "$@"
