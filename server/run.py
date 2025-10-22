#!/usr/bin/env ./.venv/bin/python

import os
from dotenv import load_dotenv

# carrega variáveis do .env (opcional, usando caminho absoluto)
BASE_DIR = os.path.dirname(__file__)
ENV_PATH = os.path.join(BASE_DIR, ".env")
load_dotenv(dotenv_path=ENV_PATH)

# força SQLite local em /tmp quando rodar via run.py (dev)
os.environ["TATUSCAN_DB_DIR"] = "/tmp"
os.environ.setdefault("TATUSCAN_DB_FILE", "tatuscan.db")

from tatuscan import create_app  # importa só depois de setar o ambiente

app = create_app()

if __name__ == "__main__":
    # Porta: prioriza TATUSCAN_PORT; permite fallback para PORT
    port = int(os.environ.get("TATUSCAN_PORT") or os.environ.get("PORT", 8040))
    app.run(host="0.0.0.0", port=port, debug=True)
