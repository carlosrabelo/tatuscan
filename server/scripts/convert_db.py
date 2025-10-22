#!/usr/bin/env python3
# scripts/convert_db.py
"""
Conversor fixo (sem parâmetros) do banco legado -> novo formato do TatuScanD.

- Origem (SQLite legado): /tmp/tatuscan_legacy.db
- Destino (SQLite novo):  /tmp/tatuscan_new.db
- Recria o schema do destino (drop + create)
- Normaliza datas para America/Cuiaba

Uso:
  export PYTHONPATH="$PWD/src"
  python scripts/convert_db.py
"""

import os
import sys
import logging
from datetime import datetime
from typing import Optional

import pytz
from sqlalchemy import create_engine, MetaData, Table, select
from sqlalchemy.orm import sessionmaker
from flask import Flask

# Garante que "src" esteja no PYTHONPATH
if "src" not in sys.path:
    sys.path.append(os.path.join(os.getcwd(), "src"))

# Importa a extensão e o modelo do projeto
from tatuscan.extensions import db
from tatuscan.models.inventory import Inventory


TIMEZONE = os.getenv("TIMEZONE", "America/Cuiaba")
SRC_SQLITE = "sqlite:////tmp/tatuscan_legacy.db"
DST_SQLITE = "sqlite:////tmp/tatuscan_new.db"


def parse_dt(value, tz):
    """Converte str/naive/aware em datetime aware na timezone tz."""
    if not value:
        return None
    if isinstance(value, datetime):
        return tz.localize(value) if value.tzinfo is None else value.astimezone(tz)

    s = str(value).strip()
    for fmt in ("%Y-%m-%d %H:%M:%S", "%Y-%m-%d", "%Y-%m-%dT%H:%M:%S"):
        try:
            dt = datetime.strptime(s, fmt)
            return tz.localize(dt)
        except ValueError:
            pass
    try:
        iso = s.replace("Z", "+00:00")
        dt = datetime.fromisoformat(iso)
        return tz.localize(dt) if dt.tzinfo is None else dt.astimezone(tz)
    except Exception:
        logging.warning(f"Não consegui parsear data: {s!r}")
        return None


def main():
    logging.basicConfig(level=logging.INFO, format="%(levelname)s: %(message)s")
    tz = pytz.timezone(TIMEZONE)

    # --- garantir /tmp e arquivo de origem ---
    os.makedirs("/tmp", exist_ok=True)
    src_path = SRC_SQLITE.replace("sqlite:////", "/")
    if not os.path.exists(src_path):
        logging.error("Arquivo de origem não encontrado: /tmp/tatuscan_legacy.db")
        sys.exit(1)

    # ---------- Origem (reflexão + leitura como mappings) ----------
    logging.info(f"Origem (SQLite legado): {SRC_SQLITE}")
    src_engine = create_engine(SRC_SQLITE)
    src_md = MetaData()
    try:
        src_inventory = Table("inventory", src_md, autoload_with=src_engine)
    except Exception as e:
        logging.error(f"Não achei a tabela 'inventory' na origem: {e}")
        sys.exit(1)

    with src_engine.connect() as conn:
        # IMPORTANTE: usar mappings() para obter RowMapping (dict-like)
        result = conn.execute(select(src_inventory))
        rows = result.mappings().all()

    logging.info(f"Registros lidos da origem: {len(rows)}")

    # ---------- Destino (mini Flask app só p/ ORM apontando p/ /tmp) ----------
    dst_path = DST_SQLITE.replace("sqlite:////", "/")
    os.makedirs(os.path.dirname(dst_path), exist_ok=True)

    app = Flask("tatuscan_migrator")
    app.config.update(
        SQLALCHEMY_DATABASE_URI=DST_SQLITE,
        SQLALCHEMY_TRACK_MODIFICATIONS=False,
    )
    db.init_app(app)

    with app.app_context():
        logging.info(f"Destino (SQLite novo): {DST_SQLITE}")

        # recria do zero
        try:
            db.drop_all()
        except Exception:
            pass
        db.create_all()

        Session = sessionmaker(bind=db.engine)
        session = Session()

        inserted = updated = 0
        try:
            for r in rows:  # r é RowMapping (dict-like)
                # defensivo: se a coluna vier com outra capitalização, normaliza chaves
                if "machine_id" not in r:
                    # tenta achar insensível a maiúsculas
                    lower_map = {k.lower(): k for k in r.keys()}
                    if "machine_id" in lower_map:
                        mid_key = lower_map["machine_id"]
                    else:
                        raise KeyError(f"Coluna 'machine_id' não encontrada. Chaves disponíveis: {list(r.keys())}")
                else:
                    mid_key = "machine_id"

                machine_id = str(r[mid_key])

                # usar r.get(...) pois é mapeamento
                payload = dict(
                    machine_id=machine_id,
                    hostname=(r.get("hostname") or "")[:100],
                    ip=(r.get("ip") or "")[:45],
                    os=(r.get("os") or "")[:100],
                    os_version=(r.get("os_version") or "")[:100],
                    cpu_percent=float(r.get("cpu_percent") or 0.0),
                    memory_total_mb=int(r.get("memory_total_mb") or 0),
                    memory_used_mb=int(r.get("memory_used_mb") or 0),
                    created_at=parse_dt(r.get("created_at"), tz),
                    updated_at=parse_dt(r.get("updated_at"), tz),
                    computer_model=(r.get("computer_model") or None),
                    computer_activation=parse_dt(r.get("computer_activation"), tz),
                    activation_days=int(r.get("activation_days") or 0)
                        if r.get("activation_days") is not None else None,
                )

                if payload["created_at"] is None:
                    payload["created_at"] = datetime.now(tz)

                existing: Optional[Inventory] = session.get(Inventory, machine_id)
                if existing:
                    for k, v in payload.items():
                        setattr(existing, k, v)
                    updated += 1
                else:
                    session.add(Inventory(**payload))
                    inserted += 1

                if (inserted + updated) % 500 == 0:
                    session.flush()

            session.commit()
            logging.info(f"Feito. Inseridos: {inserted} | Atualizados: {updated}")
            logging.info("Arquivos prontos em /tmp:")
            logging.info(" - Legado (leitura): /tmp/tatuscan_legacy.db")
            logging.info(" - Novo (gerado):    /tmp/tatuscan_new.db")
        except Exception:
            session.rollback()
            logging.exception("Erro durante a conversão")
            sys.exit(2)
        finally:
            session.close()


if __name__ == "__main__":
    main()
