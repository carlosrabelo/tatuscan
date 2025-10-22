#!/usr/bin/env bash
# Clean TatuScan database only
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_ROOT/data"

echo "→ Cleaning database..."
echo "  WARNING: This will delete all local database files!"
echo ""

# Ask for confirmation
read -p "  Are you sure you want to delete database files? [y/N] " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "✗ Database cleanup cancelled"
    exit 0
fi

# Remove data directory
if [ -d "$DATA_DIR" ]; then
    rm -rf "$DATA_DIR"
    echo "  ✓ Removed $DATA_DIR"
fi

# Remove any .db files in server directory (legacy)
if compgen -G "$PROJECT_ROOT/server/*.db" > /dev/null 2>&1; then
    rm -f "$PROJECT_ROOT/server"/*.db
    echo "  ✓ Removed server/*.db files"
fi

# Remove any .db files in client directory (legacy)
if compgen -G "$PROJECT_ROOT/client/*.db" > /dev/null 2>&1; then
    rm -f "$PROJECT_ROOT/client"/*.db
    echo "  ✓ Removed client/*.db files"
fi

echo "✓ Database cleanup completed"
