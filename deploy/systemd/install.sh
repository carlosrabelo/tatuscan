#!/usr/bin/env bash
# Install TatuScan API + Web as systemd services
set -euo pipefail

SERVICE_USER="tatuscan"
INSTALL_DIR="/opt/tatuscan"
DATA_DIR="/var/lib/tatuscan"
LOG_DIR="/var/log/tatuscan"

if [[ $EUID -ne 0 ]]; then
  echo "[ERROR] This script must be run as root"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "[INFO] Building binaries..."
(cd "$PROJECT_ROOT/api" && ./.make/build.sh)
(cd "$PROJECT_ROOT/web" && ./.make/build.sh)

if ! id "$SERVICE_USER" &>/dev/null; then
  useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
fi

mkdir -p "$INSTALL_DIR/bin" "$DATA_DIR" "$LOG_DIR"
install -m 0755 "$PROJECT_ROOT/api/bin/linux/tatuscan-api" "$INSTALL_DIR/bin/tatuscan-api"
install -m 0755 "$PROJECT_ROOT/web/bin/linux/tatuscan-web" "$INSTALL_DIR/bin/tatuscan-web"
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR" "$DATA_DIR" "$LOG_DIR"

install -m 0644 "$SCRIPT_DIR/tatuscan-api.service" /etc/systemd/system/
install -m 0644 "$SCRIPT_DIR/tatuscan-web.service" /etc/systemd/system/

systemctl daemon-reload
systemctl enable --now tatuscan-api.service
systemctl enable --now tatuscan-web.service

echo "[INFO] API:  http://localhost:8040"
echo "[INFO] Web:  http://localhost:8050"
echo "[INFO] Status: systemctl status tatuscan-api tatuscan-web"
