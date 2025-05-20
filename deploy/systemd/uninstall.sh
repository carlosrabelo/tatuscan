#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "[ERROR] This script must be run as root"
  exit 1
fi

systemctl disable --now tatuscan-web.service 2>/dev/null || true
systemctl disable --now tatuscan-api.service 2>/dev/null || true
# legacy units
systemctl disable --now tatuscan.socket 2>/dev/null || true
systemctl disable --now tatuscan.service 2>/dev/null || true

rm -f /etc/systemd/system/tatuscan-api.service
rm -f /etc/systemd/system/tatuscan-web.service
rm -f /etc/systemd/system/tatuscan.service
rm -f /etc/systemd/system/tatuscan.socket
rm -f /etc/systemd/system/tatuscan@.service
systemctl daemon-reload

echo "[INFO] Systemd units removed (binaries/data under /opt/tatuscan and /var/lib/tatuscan left intact)"
