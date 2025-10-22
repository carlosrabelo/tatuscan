#!/usr/bin/env bash
# Run TatuScan server locally (without Docker)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_DIR="$PROJECT_ROOT/server"
VENV_DIR="$SERVER_DIR/.venv"

SERVER_PORT="${SERVER_PORT:-8040}"

echo "→ Running server locally..."

# Check if venv exists
if [ ! -d "$VENV_DIR" ]; then
    echo "[ERROR] Virtual environment not found at $VENV_DIR"
    echo "Run: $SCRIPT_DIR/setup-venv.sh"
    exit 1
fi

# Run server
cd "$SERVER_DIR"
source .venv/bin/activate
export PYTHONPATH="$SERVER_DIR/src"
export PORT="$SERVER_PORT"
export TATUSCAND_PORT="$SERVER_PORT"
python run.py
