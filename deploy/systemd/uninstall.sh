#!/bin/bash
# Uninstall TatuScan server systemd service

set -e

# Configuration
SERVICE_USER="tatuscan"
SERVICE_GROUP="tatuscan"
INSTALL_DIR="/opt/tatuscan"
DATA_DIR="/var/lib/tatuscan"
LOG_DIR="/var/log/tatuscan"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo_error "This script must be run as root"
   exit 1
fi

echo_info "Uninstalling TatuScan server systemd service..."

# Stop and disable services
echo_info "Stopping and disabling services..."
systemctl stop tatuscan.socket 2>/dev/null || true
  systemctl stop tatuscan@.service 2>/dev/null || true
  systemctl disable tatuscan.socket 2>/dev/null || true
  systemctl disable tatuscan@.service 2>/dev/null || true

# Remove service files
echo_info "Removing systemd service files..."
rm -f /etc/systemd/system/tatuscan.service
  rm -f /etc/systemd/system/tatuscan.socket
  rm -f /etc/systemd/system/tatuscan@.service

# Reload systemd
echo_info "Reloading systemd daemon..."
systemctl daemon-reload

# Ask about data removal
echo_warn "Do you want to remove TatuScan data directory ($DATA_DIR)? [y/N]"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo_info "Removing data directory..."
    rm -rf "$DATA_DIR"
fi

# Ask about log removal
echo_warn "Do you want to remove TatuScan log directory ($LOG_DIR)? [y/N]"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo_info "Removing log directory..."
    rm -rf "$LOG_DIR"
fi

# Ask about installation removal
echo_warn "Do you want to remove TatuScan installation directory ($INSTALL_DIR)? [y/N]"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo_info "Removing installation directory..."
    rm -rf "$INSTALL_DIR"
fi

# Ask about user removal
echo_warn "Do you want to remove TatuScan service user ($SERVICE_USER)? [y/N]"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo_info "Removing service user..."
    userdel "$SERVICE_USER" 2>/dev/null || true
fi

echo_info "TatuScan server service uninstalled successfully!"