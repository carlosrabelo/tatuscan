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

for comp in api client web; do
  if compgen -G "$PROJECT_ROOT/$comp/*.db" > /dev/null 2>&1; then
    rm -f "$PROJECT_ROOT/$comp"/*.db
    echo "  ✓ Removed $comp/*.db files"
  fi
done
rm -f /tmp/tatuscan.db 2>/dev/null || true

echo "✓ Database cleanup completed"
