#!/bin/bash
# Build script for multiple platforms

set -e

# Variables from Makefile
BIN="${1:-tatuscan}"
SRC="${2:-./cmd/tatuscan}"
BUILD_DIR="${3:-./bin}"
LDFLAGS="${4:--s -w}"

echo "Building $BIN for multiple platforms..."
mkdir -p "$BUILD_DIR"

# Define platforms and architectures
declare -a platforms=(
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
    "windows/arm64"
    "darwin/amd64"
    "darwin/arm64"
)

# Build for each platform
for platform in "${platforms[@]}"; do
    IFS='/' read -r os arch <<< "$platform"

    output_name="$BIN-$os-$arch"
    if [ "$os" = "windows" ]; then
        output_name="$output_name.exe"
    fi

    echo "  Building for $os/$arch -> $output_name"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
        -ldflags="$LDFLAGS" \
        -o "$BUILD_DIR/$output_name" \
        "$SRC"
done

echo "Multi-platform build completed successfully!"