#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export TATUSCAN_URL="${TATUSCAN_URL:-http://127.0.0.1:8040}"
export TATUSCAN_INTERVAL="${TATUSCAN_INTERVAL:-60s}"
export TATUSCAN_LOG_LEVEL="${TATUSCAN_LOG_LEVEL:-info}"

echo "→ Starting tatuscan client (url=$TATUSCAN_URL interval=$TATUSCAN_INTERVAL)"
# Extra args are forwarded (e.g. make run -- -d -l debug).
go run ./client/cmd/tatuscan "$@"
