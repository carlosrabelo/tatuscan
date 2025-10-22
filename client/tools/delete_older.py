#!/usr/bin/env python3
"""Remove duplicatas de inventário via API mantendo o registro mais recente."""

from __future__ import annotations

import argparse
import json
import logging
import os
from collections import defaultdict
from datetime import datetime
from typing import Dict, Iterable, List, Tuple
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

DEFAULT_API_BASE = "http://localhost:8040/api"


def resolver_api_base(api_base: str | None) -> str:
    env_base = os.environ.get("TATUSCAN_URL")
    if env_base:
        base_url = env_base.rstrip("/")
        if not base_url.endswith("/api"):
            base_url += "/api"
        return base_url
    return (api_base or DEFAULT_API_BASE).rstrip("/")


def _request(url: str, *, method: str = "GET", data: bytes | None = None):
    headers = {"Accept": "application/json"}
    if data is not None:
        headers["Content-Type"] = "application/json"
    req = Request(url, data=data, headers=headers, method=method)
    return urlopen(req)


def fetch_inventory(api_base: str) -> List[Dict[str, str]]:
    endpoints = ["/machines", "/inventory"]
    last_error: HTTPError | URLError | None = None

    for endpoint in endpoints:
        url = f"{api_base.rstrip('/')}{endpoint}"
        try:
            with _request(url) as response:
                payload = json.load(response)
            return payload.get("items") or []
        except HTTPError as exc:
            last_error = exc
            if exc.code not in {404, 405}:
                raise
            logging.debug(
                "Endpoint %s retornou %s. Tentando próximo endpoint.",
                endpoint,
                exc.code,
            )
        except URLError as exc:
            last_error = exc
            logging.debug(
                "Falha ao acessar %s: %s",
                endpoint,
                exc.reason,
            )

    if last_error:
        raise last_error

    return []


def delete_inventory(api_base: str, machine_id: str) -> None:
    url = f"{api_base.rstrip('/')}/machines/{machine_id}"
    with _request(url, method="DELETE") as response:
        response.read()


def _parse_dt(value: str | None) -> datetime | None:
    if not value:
        return None
    value = value.strip()
    if not value:
        return None
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"
    try:
        return datetime.fromisoformat(value)
    except ValueError:
        logging.debug("Data em formato inesperado: %s", value)
        return None


def ordenar_registros(registros: Iterable[Dict[str, str]]) -> List[Dict[str, str]]:
    def _score(item: Dict[str, str]) -> Tuple[datetime, datetime]:
        updated = _parse_dt(item.get("updated_at"))
        created = _parse_dt(item.get("created_at"))
        # ordena pelo mais recente considerando updated, depois created
        return (
            updated or created or datetime.min,
            created or datetime.min,
        )

    return sorted(registros, key=_score, reverse=True)


def processar(api_base: str, *, dry_run: bool = False) -> Tuple[int, int, int, List[str]]:
    inventarios = fetch_inventory(api_base)

    grupos: Dict[str, List[Dict[str, str]]] = defaultdict(list)
    for item in inventarios:
        hostname = (item.get("hostname") or "").strip()
        if not hostname:
            continue
        grupos[hostname].append(item)

    duplicados = 0
    removidos = 0
    ignorados = 0
    erros: List[str] = []

    for hostname, registros in grupos.items():
        if len(registros) <= 1:
            continue

        duplicados += 1
        ordenados = ordenar_registros(registros)
        manter = ordenados[0]
        manter_dt = _parse_dt(manter.get("updated_at")) or _parse_dt(manter.get("created_at"))
        print(f"Hostname duplicado: {hostname} (total: {len(registros)})")
        print(
            f"  Manter: machine_id={manter['machine_id']} data={manter_dt or 'desconhecida'}"
        )

        for registro in ordenados[1:]:
            drop_dt = _parse_dt(registro.get("updated_at")) or _parse_dt(
                registro.get("created_at")
            )
            machine_id = registro.get("machine_id")
            print(
                f"  Apagar: machine_id={machine_id} data={drop_dt or 'desconhecida'}"
            )
            if dry_run:
                ignorados += 1
                continue
            try:
                delete_inventory(api_base, machine_id)
                removidos += 1
            except HTTPError as exc:
                msg = f"HTTPError {exc.code} ao remover {machine_id}: {exc.reason}"
                logging.error(msg)
                erros.append(msg)
            except URLError as exc:
                msg = f"URLError ao remover {machine_id}: {exc.reason}"
                logging.error(msg)
                erros.append(msg)

    return duplicados, removidos, ignorados, erros


def configurar_argumentos() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Remove registros duplicados via API, mantendo o mais recente por hostname",
    )
    parser.add_argument(
        "--api-base",
        default=None,
        help="URL base da API (padrão: TATUSCAN_URL ou localhost)",
    )
    parser.add_argument(
        "--log-level",
        default="INFO",
        choices=["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"],
        help="Nível de log (padrão: INFO)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Exibe o que seria removido sem chamar a API de DELETE",
    )

    return parser.parse_args()


def main() -> None:
    args = configurar_argumentos()
    logging.basicConfig(level=getattr(logging, args.log_level))

    api_base = resolver_api_base(args.api_base)

    (
        duplicados,
        removidos,
        ignorados,
        erros,
    ) = processar(api_base, dry_run=args.dry_run)

    print(f"Hostnames com duplicatas: {duplicados}")
    if args.dry_run:
        print(f"Registros marcados (dry-run): {ignorados}")
    else:
        print(f"Registros removidos: {removidos}")

    if erros:
        print("Ocorreram erros durante a remoção:")
        for linha in erros:
            print(f" - {linha}")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        logging.exception("Erro ao remover duplicatas")
