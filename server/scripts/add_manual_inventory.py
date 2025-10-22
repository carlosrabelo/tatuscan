#!/usr/bin/env python3
"""Envia registros manuais de inventário via API do TatuScanD."""

from __future__ import annotations

import argparse
import hashlib
import json
import logging
import os
from dataclasses import dataclass
from datetime import datetime
from typing import Optional
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

DEFAULT_API_BASE = "http://localhost:8040/api"


@dataclass
class Entry:
    hostname: str
    os: str
    os_version: str
    machine_id: Optional[str] = None
    ip: str = "0.0.0.0"
    cpu_percent: float = 0.0
    memory_total_mb: int = 0
    memory_used_mb: int = 0


def resolve_api_base(api_base: str | None) -> str:
    env_base = os.environ.get("TATUSCAN_URL")
    if env_base:
        base_url = env_base.rstrip("/")
        if not base_url.endswith("/api"):
            base_url += "/api"
        return base_url
    return (api_base or DEFAULT_API_BASE).rstrip("/")


def generate_machine_id(hostname: str, salt: str | None = None) -> str:
    base = hostname.strip().encode("utf-8")
    extra = (salt or "").encode("utf-8")
    return hashlib.sha256(base + extra).hexdigest()


def normalize_activation(raw: str | None) -> Optional[str]:
    if not raw:
        return None
    value = raw.strip()
    if not value:
        return None

    known_formats = ("%Y-%m-%d", "%d/%m/%Y", "%Y/%m/%d")
    for fmt in known_formats:
        try:
            return datetime.strptime(value, fmt).strftime("%Y-%m-%d")
        except ValueError:
            continue
    return value


def entry_from_args(args: argparse.Namespace) -> Entry:
    os_name = args.os or "Chrome OS"
    os_version = args.os_version or ""

    return Entry(
        hostname=args.hostname,
        os=os_name,
        os_version=os_version,
        machine_id=args.machine_id,
        ip=args.ip,
        cpu_percent=args.cpu_percent,
        memory_total_mb=0,
        memory_used_mb=0,
    )


def send_payload(api_base: str, entry: Entry, *, dry_run: bool, salt: str | None) -> None:
    machine_id = entry.machine_id or generate_machine_id(entry.hostname, salt)
    payload = {
        "machine_id": machine_id,
        "hostname": entry.hostname,
        "ip": entry.ip,
        "os": entry.os,
        "os_version": entry.os_version,
        "cpu_percent": entry.cpu_percent,
        "memory_total_mb": entry.memory_total_mb,
        "memory_used_mb": entry.memory_used_mb,
    }

    url = f"{api_base}/machines"

    if dry_run:
        logging.info("[dry-run] POST %s -> %s", url, json.dumps(payload, ensure_ascii=False))
        return

    data = json.dumps(payload).encode("utf-8")
    req = Request(
        url,
        data=data,
        headers={"Content-Type": "application/json", "Accept": "application/json"},
        method="POST",
    )

    try:
        with urlopen(req) as resp:
            body = resp.read().decode("utf-8") or "{}"
            try:
                message = json.loads(body).get("message", "OK")
            except json.JSONDecodeError:
                message = body.strip() or resp.reason
            logging.info("Hostname %s -> %s", entry.hostname, message)
    except HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="ignore")
        logging.error("HTTP %s ao enviar %s: %s", exc.code, entry.hostname, detail or exc.reason)
    except URLError as exc:
        logging.error("Erro de rede ao enviar %s: %s", entry.hostname, exc.reason)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Registra máquinas manualmente via API do TatuScanD.",
        formatter_class=argparse.ArgumentDefaultsHelpFormatter,
    )

    parser.add_argument("--hostname", required=True, help="Hostname a ser enviado")
    parser.add_argument("--os", help="Nome do sistema operacional", default="Chrome OS")
    parser.add_argument("--os-version", dest="os_version", help="Versão do sistema operacional")
    parser.add_argument("--machine-id", help="Machine ID fixo (senão gera a partir do hostname)")
    parser.add_argument("--ip", help="IP reportado", default="0.0.0.0")
    parser.add_argument("--cpu-percent", dest="cpu_percent", type=float, default=0.0, help="Carga de CPU")
    parser.add_argument("--api-base", help="Endpoint base da API")
    parser.add_argument("--salt", help="Sal extra para gerar machine_id", default=None)
    parser.add_argument("--dry-run", action="store_true", help="Só exibe o payload, não envia")
    parser.add_argument("--log-level", default="INFO", choices=["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"], help="Nível de log")

    return parser.parse_args()


def main() -> None:
    args = parse_args()
    logging.basicConfig(level=getattr(logging, args.log_level))

    api_base = resolve_api_base(args.api_base)
    logging.debug("Usando API base: %s", api_base)

    entry = entry_from_args(args)
    send_payload(api_base, entry, dry_run=args.dry_run, salt=args.salt)
    logging.info("Máquina processada")


if __name__ == "__main__":
    main()
