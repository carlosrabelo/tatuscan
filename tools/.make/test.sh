#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"
echo "→ Running tools tests..."
go test ./tools/internal/...
echo "✓ tools tests passed"
