#!/usr/bin/env bash
# Start, smoke-test, or stop the local API+web stack (no Docker).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="${TATUSCAN_RUN_DIR:-/tmp/tatuscan-local}"
API_PORT="${TATUSCAN_PORT:-8040}"
WEB_PORT="${TATUSCAN_WEB_PORT:-8050}"
API_URL="http://127.0.0.1:${API_PORT}"
WEB_URL="http://127.0.0.1:${WEB_PORT}"
CMD="${1:-start}"

mkdir -p "$RUN_DIR"

log() { echo "$*"; }

already_up() {
  curl -sf -o /dev/null --max-time 1 "$1"
}

wait_http() {
  local url=$1
  local i=0
  while [ "$i" -lt 90 ]; do
    if curl -sf -o /dev/null --max-time 1 "$url"; then
      return 0
    fi
    i=$((i + 1))
    sleep 0.5
  done
  echo "error: timed out waiting for $url" >&2
  echo "  log: $RUN_DIR" >&2
  return 1
}

start_api() {
  log "→ Starting tatuscan-api on :$API_PORT"
  (
    cd "$ROOT_DIR/api"
    export TATUSCAN_DB_DIR="${TATUSCAN_DB_DIR:-/tmp}"
    export TATUSCAN_DB_FILE="${TATUSCAN_DB_FILE:-tatuscan.db}"
    export TATUSCAN_PORT="$API_PORT"
    export TIMEZONE="${TIMEZONE:-America/Cuiaba}"
    exec go run ./api/cmd/tatuscan-api
  ) >"$RUN_DIR/api.log" 2>&1 &
  echo $! >"$RUN_DIR/api.pid"
  echo started >"$RUN_DIR/api.owned"
  wait_http "$API_URL/api/health"
}

start_web() {
  log "→ Starting tatuscan-web on :$WEB_PORT (api=$API_URL)"
  (
    cd "$ROOT_DIR/web"
    export TATUSCAN_API_URL="$API_URL"
    export TATUSCAN_PORT="$WEB_PORT"
    exec go run ./web/cmd/tatuscan-web
  ) >"$RUN_DIR/web.log" 2>&1 &
  echo $! >"$RUN_DIR/web.pid"
  echo started >"$RUN_DIR/web.owned"
  wait_http "$WEB_URL/healthz"
}

ensure_up() {
  if already_up "$API_URL/api/health"; then
    log "✓ API already up at $API_URL"
  else
    start_api
  fi
  if already_up "$WEB_URL/healthz"; then
    log "✓ Web already up at $WEB_URL"
  else
    start_web
  fi
}

stop_owned() {
  local name=$1
  local pidfile="$RUN_DIR/${name}.pid"
  local owned="$RUN_DIR/${name}.owned"
  if [ ! -f "$owned" ] || [ ! -f "$pidfile" ]; then
    rm -f "$pidfile" "$owned"
    return 0
  fi
  local pid
  pid="$(cat "$pidfile" 2>/dev/null || true)"
  if [ -n "${pid:-}" ] && kill -0 "$pid" 2>/dev/null; then
    pkill -P "$pid" 2>/dev/null || true
    kill "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
    log "✓ Stopped $name (pid $pid)"
  fi
  rm -f "$pidfile" "$owned"
}

stop_all() {
  stop_owned web
  stop_owned api
}

print_urls() {
  log ""
  log "Local stack (no Docker):"
  log "  API    $API_URL/"
  log "  Health $API_URL/api/health"
  log "  Panel  $WEB_URL/"
  log ""
}

smoke() {
  local fail=0
  check() {
    local url=$1
    if ! curl -sf -o /dev/null --max-time 5 "$url"; then
      echo "FAIL  $url" >&2
      fail=1
      return
    fi
    log "  OK    $url"
  }

  log "→ HTTP smoke"
  check "$API_URL/api/health"
  check "$API_URL/api/machines"
  if curl -sf -o /dev/null --max-time 5 -H "Accept: text/html" "$API_URL/"; then
    log "  OK    $API_URL/"
  else
    log "  WARN  $API_URL/  (restart the API to pick up the landing page)"
  fi
  check "$WEB_URL/healthz"
  check "$WEB_URL/"
  if [ "$fail" -ne 0 ]; then
    echo "error: smoke checks failed" >&2
    return 1
  fi
  log "✓ Smoke checks passed"
}

case "$CMD" in
  start)
    ensure_up
    print_urls
    if [ -f "$RUN_DIR/api.owned" ] || [ -f "$RUN_DIR/web.owned" ]; then
      log "Ctrl+C stops the processes started by this command."
      trap stop_all INT TERM
      if [ -f "$RUN_DIR/api.owned" ]; then
        wait "$(cat "$RUN_DIR/api.pid")" || true
      elif [ -f "$RUN_DIR/web.owned" ]; then
        wait "$(cat "$RUN_DIR/web.pid")" || true
      fi
      stop_all
    fi
    ;;
  smoke)
    ensure_up
    smoke
    print_urls
    ;;
  test)
    log "→ Unit tests"
    "$ROOT_DIR/client/.make/test.sh"
    "$ROOT_DIR/api/.make/test.sh"
    "$ROOT_DIR/web/.make/test.sh"
    "$ROOT_DIR/tools/.make/test.sh"
    ensure_up
    smoke
    print_urls
    ;;
  stop)
    stop_all
    log "✓ Local stack stopped"
    ;;
  *)
    echo "usage: $0 {start|test|smoke|stop}" >&2
    exit 2
    ;;
esac
