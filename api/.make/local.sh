#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export TATUSCAN_DB_DIR="${TATUSCAN_DB_DIR:-/tmp}"
export TATUSCAN_DB_FILE="${TATUSCAN_DB_FILE:-tatuscan.db}"
export TATUSCAN_PORT="${TATUSCAN_PORT:-8040}"
export TIMEZONE="${TIMEZONE:-America/Cuiaba}"

echo "→ Starting tatuscan-api on :$TATUSCAN_PORT (db=$TATUSCAN_DB_DIR/$TATUSCAN_DB_FILE)"
go run ./api/cmd/tatuscan-api
