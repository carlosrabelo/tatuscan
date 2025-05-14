#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$ROOT_DIR/bin"
VERSION="${WEB_VERSION:-$(git -C "$ROOT_DIR" describe --tags --always --dirty 2>/dev/null || echo dev)}"
CMD_PATH="./web/cmd/tatuscan-web"

cd "$ROOT_DIR"
mkdir -p "$BIN_DIR/linux"
echo "→ Building tatuscan-web..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-s -w -X main.version=$VERSION" \
  -o "$BIN_DIR/linux/tatuscan-web" "$CMD_PATH"
echo "✓ Built: $BIN_DIR/linux/tatuscan-web"
