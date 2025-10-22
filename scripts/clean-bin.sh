#!/usr/bin/env bash
# Clean TatuScan binaries only
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_ROOT/bin"
CLIENT_DIR="$PROJECT_ROOT/client"

echo "→ Cleaning binaries..."

# Remove binaries from bin directory (preserve .gitkeep files)
if [ -d "$BIN_DIR" ]; then
    find "$BIN_DIR" -type f ! -name '.gitkeep' -delete
    find "$BIN_DIR" -type d -empty ! -name 'bin' -delete 2>/dev/null || true
    echo "  ✓ Cleaned $BIN_DIR (preserved .gitkeep files)"
fi

# Remove legacy binaries in client dir
if [ -f "$CLIENT_DIR/tatuscan" ]; then
    rm -f "$CLIENT_DIR/tatuscan"
    echo "  ✓ Removed $CLIENT_DIR/tatuscan"
fi

if [ -f "$CLIENT_DIR/tatuscan.exe" ]; then
    rm -f "$CLIENT_DIR/tatuscan.exe"
    echo "  ✓ Removed $CLIENT_DIR/tatuscan.exe"
fi

# Remove MSI installers
if compgen -G "$CLIENT_DIR/*.msi" > /dev/null 2>&1; then
    rm -f "$CLIENT_DIR"/*.msi
    echo "  ✓ Removed *.msi files"
fi

# Remove build directory
if [ -d "$CLIENT_DIR/build" ]; then
    rm -rf "$CLIENT_DIR/build"
    echo "  ✓ Removed $CLIENT_DIR/build"
fi

echo "✓ Binary cleanup completed"
