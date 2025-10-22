#!/usr/bin/env bash
# Run TatuScan client tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CLIENT_DIR="$PROJECT_ROOT/client"

echo "→ Running client tests..."
cd "$CLIENT_DIR"
go test -v ./...
echo "✓ Tests completed"
