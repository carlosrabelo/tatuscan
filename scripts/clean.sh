#!/usr/bin/env bash
# Clean TatuScan build artifacts (binaries + Docker cache)
# NOTE: This does NOT clean the database. Use clean-db.sh for that.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "→ Cleaning TatuScan artifacts..."
echo ""

# Clean binaries
"$SCRIPT_DIR/clean-bin.sh"
echo ""

# Clean Docker resources (project-specific)
echo "→ Cleaning Docker resources for tatuscan..."
cd "$SCRIPT_DIR/../server"
docker compose down --rmi local --volumes 2>/dev/null || true
docker image rm carlosrabelo/tatuscan:latest 2>/dev/null || true
echo "  ✓ Docker resources cleaned (project-specific only)"

echo ""
echo "✓ Cleanup completed"
echo ""
echo "NOTE: Database was NOT cleaned. Run 'make clean-db' to remove database files."
