#!/bin/sh
set -e

strip() { printf '%s' "$1" | tr -d '\r' | xargs; }

# Normaliza variáveis (remove \r e espaços sobrando)
TATUSCAN_PORT="$(strip "${TATUSCAN_PORT:-8040}")"
GUNICORN_WORKERS="$(strip "${GUNICORN_WORKERS:-3}")"
GUNICORN_THREADS="$(strip "${GUNICORN_THREADS:-2}")"
GUNICORN_TIMEOUT="$(strip "${GUNICORN_TIMEOUT:-30}")"
LOG_LEVEL="$(strip "${LOG_LEVEL:-info}")"
SQLALCHEMY_DATABASE_URI="$(strip "${SQLALCHEMY_DATABASE_URI:-}")"
TATUSCAN_DB_DIR="$(strip "${TATUSCAN_DB_DIR:-/data}")"

# Se for SQLite, garante o diretório
if [ -z "$SQLALCHEMY_DATABASE_URI" ] || printf '%s' "$SQLALCHEMY_DATABASE_URI" | grep -qi '^sqlite'; then
  mkdir -p "$TATUSCAN_DB_DIR"
fi

# Inicializa o banco (create_all)
python - <<'PY'
from tatuscan import create_app
from tatuscan.extensions import db
app = create_app()
with app.app_context():
    db.create_all()
    print("DB initialized (create_all).")
PY

# Sobe o Gunicorn
exec gunicorn 'tatuscan:create_app()' \
  --workers "$GUNICORN_WORKERS" \
  --threads "$GUNICORN_THREADS" \
  --bind "0.0.0.0:$TATUSCAN_PORT" \
  --timeout "$GUNICORN_TIMEOUT" \
  --log-level "$LOG_LEVEL"
