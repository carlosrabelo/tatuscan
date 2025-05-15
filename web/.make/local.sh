#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export TATUSCAN_API_URL="${TATUSCAN_API_URL:-http://127.0.0.1:8040}"
export TATUSCAN_PORT="${TATUSCAN_PORT:-8050}"

echo "→ Starting tatuscan-web on :$TATUSCAN_PORT (api=$TATUSCAN_API_URL)"
go run ./web/cmd/tatuscan-web
