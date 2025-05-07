#!/usr/bin/env bash
# Run TatuScan client tests
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="$(dirname "$SCRIPT_DIR")"

echo "→ Running client tests..."
cd "$CLIENT_DIR"
go test -v ./...
echo "✓ client tests passed"
