#!/usr/bin/env bash
# Build TatuScan client
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"

CLIENT_BINARY="${CLIENT_BINARY:-tatuscan}"
CLIENT_VERSION="${CLIENT_VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
CMD_PATH="./client/cmd/tatuscan"
PLATFORM="${1:-linux}"

cd "$ROOT_DIR"

case "$PLATFORM" in
    linux)
        echo "→ Building client for Linux..."
        mkdir -p "$BIN_DIR/linux"
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
            go build -ldflags="-X main.version=$CLIENT_VERSION" \
            -o "$BIN_DIR/linux/$CLIENT_BINARY" "$CMD_PATH"
        echo "✓ Client built: $BIN_DIR/linux/$CLIENT_BINARY"
        ;;
    windows)
        echo "→ Building client for Windows..."
        mkdir -p "$BIN_DIR/windows"
        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc \
            go build -ldflags="-X main.version=$CLIENT_VERSION -H windowsgui" \
            -o "$BIN_DIR/windows/$CLIENT_BINARY.exe" "$CMD_PATH"
        echo "✓ Client built: $BIN_DIR/windows/$CLIENT_BINARY.exe"
        ;;
    all)
        echo "→ Building client for all platforms..."
        "$SCRIPT_DIR/build.sh" linux
        "$SCRIPT_DIR/build.sh" windows
        echo "✓ All builds completed"
        ;;
    *)
        echo "[ERROR] Unknown platform: $PLATFORM"
        echo "Usage: $0 {linux|windows|all}"
        exit 1
        ;;
esac
