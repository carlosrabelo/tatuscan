#!/usr/bin/env bash
# Clean TatuScan binaries only
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "→ Cleaning binaries..."

for comp in client api web tools; do
  BIN_DIR="$PROJECT_ROOT/$comp/bin"
  if [ -d "$BIN_DIR" ]; then
    find "$BIN_DIR" -type f ! -name '.gitkeep' -delete
    find "$BIN_DIR" -type d -empty ! -name 'bin' -delete 2>/dev/null || true
    echo "  ✓ Cleaned $BIN_DIR"
  fi
done

CLIENT_DIR="$PROJECT_ROOT/client"
rm -f "$CLIENT_DIR/tatuscan" "$CLIENT_DIR/tatuscan.exe"
if [ -d "$CLIENT_DIR/build" ]; then
  rm -rf "$CLIENT_DIR/build"
fi
if compgen -G "$CLIENT_DIR/*.msi" > /dev/null 2>&1; then
  rm -f "$CLIENT_DIR"/*.msi
fi

echo "✓ Binary cleanup completed"
