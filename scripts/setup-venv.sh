#!/usr/bin/env bash
# Setup virtual environment for TatuScan server
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SERVER_DIR="$PROJECT_ROOT/server"
VENV_DIR="$SERVER_DIR/.venv"

echo "→ Setting up Python virtual environment..."

# Check if Python is available
if ! command -v python3 &> /dev/null; then
    echo "[ERROR] python3 not found. Please install Python 3."
    exit 1
fi

# Create venv if it doesn't exist
if [ ! -d "$VENV_DIR" ]; then
    echo "  Creating virtual environment at $VENV_DIR"
    cd "$SERVER_DIR"
    python3 -m venv .venv
else
    echo "  Virtual environment already exists at $VENV_DIR"
fi

# Activate and upgrade pip
echo "  Upgrading pip..."
"$VENV_DIR/bin/pip" install --upgrade pip -q

# Install requirements
if [ -f "$SERVER_DIR/requirements.txt" ]; then
    echo "  Installing dependencies from requirements.txt..."
    "$VENV_DIR/bin/pip" install -r "$SERVER_DIR/requirements.txt" -q
else
    echo "[WARNING] requirements.txt not found"
fi

echo "✓ Virtual environment ready at $VENV_DIR"
echo ""
echo "To activate manually:"
echo "  cd $SERVER_DIR && source .venv/bin/activate"
