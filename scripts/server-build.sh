#!/usr/bin/env bash
# Build TatuScan server Docker image
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_DIR="$PROJECT_ROOT/server"

# Variables
SERVER_IMAGE="${SERVER_IMAGE:-carlosrabelo/tatuscand}"
SERVER_TAG="${SERVER_TAG:-$(git describe --tags --always --dirty 2>/dev/null || echo latest)}"
SERVER_FULL="$SERVER_IMAGE:$SERVER_TAG"
SERVER_PORT="${SERVER_PORT:-8040}"
COMPOSE="${COMPOSE:-docker compose}"

echo "→ Building server image: $SERVER_FULL"
cd "$SERVER_DIR"
APP_IMAGE="$SERVER_FULL" PORT="$SERVER_PORT" $COMPOSE build
echo "✓ Server image built successfully"
