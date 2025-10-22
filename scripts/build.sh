#!/bin/bash
# Build script for local platform only

set -e

# Variables from Makefile
BIN="${1:-tatuscan}"
SRC="${2:-./cmd/tatuscan}"
BUILD_DIR="${3:-./bin}"
LDFLAGS="${4:--s -w}"

# Detect current platform
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)

    case $arch in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac

    case $os in
        linux) echo "linux" ;;
        darwin) echo "darwin" ;;
        mingw*|msys*|cygwin*) echo "windows" ;;
        *) echo "Unsupported OS: $os" >&2; exit 1 ;;
    esac
}

# Get current platform info
OS=$(detect_platform)
ARCH=$(uname -m)
case $ARCH in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

# Set output name
output_name="$BIN"
if [ "$OS" = "windows" ]; then
    output_name="$output_name.exe"
fi

echo "Building $BIN for current platform ($OS/$ARCH)..."
mkdir -p "$BUILD_DIR"

CGO_ENABLED=0 go build \
    -ldflags="$LDFLAGS" \
    -o "$BUILD_DIR/$output_name" \
    "$SRC"

echo "Build completed: $BUILD_DIR/$output_name"