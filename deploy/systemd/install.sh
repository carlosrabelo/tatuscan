#!/bin/bash
# Install TatuScan server as systemd service

set -e

# Configuration
SERVICE_USER="tatuscan"
SERVICE_GROUP="tatuscan"
INSTALL_DIR="/opt/tatuscan"
DATA_DIR="/var/lib/tatuscan"
LOG_DIR="/var/log/tatuscan"
VENV_DIR="$INSTALL_DIR/venv"

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

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo_info "Installing TatuScan server as systemd service..."
echo_info "Project root: $PROJECT_ROOT"

# Create service user if not exists
if ! id "$SERVICE_USER" &>/dev/null; then
    echo_info "Creating service user: $SERVICE_USER"
    useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
fi

# Create directories
echo_info "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"

# Copy server code
echo_info "Copying server code..."
cp -r "$PROJECT_ROOT/server/src" "$INSTALL_DIR/"
cp "$PROJECT_ROOT/server/requirements.txt" "$INSTALL_DIR/"
cp "$PROJECT_ROOT/server/run.py" "$INSTALL_DIR/" 2>/dev/null || true

# Create virtual environment
echo_info "Creating Python virtual environment..."
python3 -m venv "$VENV_DIR"
source "$VENV_DIR/bin/activate"
pip install --upgrade pip
pip install -r "$INSTALL_DIR/requirements.txt"

# Set ownership
echo_info "Setting ownership..."
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$DATA_DIR"
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$LOG_DIR"

# Install systemd service files
echo_info "Installing systemd service files..."
cp "$SCRIPT_DIR/tatuscan.service" "/etc/systemd/system/"
  cp "$SCRIPT_DIR/tatuscan.socket" "/etc/systemd/system/"
  cp "$SCRIPT_DIR/tatuscan@.service" "/etc/systemd/system/"

# Reload systemd
echo_info "Reloading systemd daemon..."
systemctl daemon-reload

# Enable and start services
echo_info "Enabling and starting TatuScan services..."
systemctl enable tatuscan.socket
  systemctl enable tatuscan@.service
  systemctl start tatuscan.socket

# Check status
echo_info "Checking service status..."
sleep 2
if systemctl is-active --quiet tatuscan.socket; then
    echo_info "✓ TatuScan socket is running"
else
    echo_error "✗ TatuScan socket failed to start"
systemctl status tatuscan.socket
fi

echo_info "Installation completed!"
echo_info "Service management commands:"
echo_info "  Status: systemctl status tatuscan.socket"
  echo_info "  Logs: journalctl -u tatuscan@.service -f"
  echo_info "  Restart: systemctl restart tatuscan@.service"
  echo_info "  Stop: systemctl stop tatuscan.socket"
echo_info ""
echo_info "TatuScan server will be available at: http://localhost:8040"