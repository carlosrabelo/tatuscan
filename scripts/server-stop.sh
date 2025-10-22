#!/usr/bin/env bash
# Stop TatuScan server
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_DIR="$PROJECT_ROOT/server"

# Variables
SERVER_IMAGE="${SERVER_IMAGE:-carlosrabelo/tatuscan}"
SERVER_TAG="${SERVER_TAG:-$(git describe --tags --always --dirty 2>/dev/null || echo latest)}"
SERVER_FULL="$SERVER_IMAGE:$SERVER_TAG"
SERVER_PORT="${SERVER_PORT:-8040}"
COMPOSE="${COMPOSE:-docker compose}"

echo "→ Stopping server..."
cd "$SERVER_DIR"

# Remove any existing containers with the same name (including dead ones)
if docker ps -a --format '{{.Names}}' | grep -q '^tatuscan$'; then
    echo "  Removing existing tatuscan container..."
    docker rm -f tatuscan 2>/dev/null || true
fi

APP_IMAGE="$SERVER_FULL" PORT="$SERVER_PORT" $COMPOSE down --remove-orphans
echo "✓ Server stopped"
