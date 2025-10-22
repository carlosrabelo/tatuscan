#!/usr/bin/env python3
"""Atualiza `computer_activation` via API usando o `/tmp/relatorio.csv`."""

from __future__ import annotations

import argparse
import csv
import json
import logging
import os
import re
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen

DEFAULT_CSV_PATH = Path("inventario.csv")
DEFAULT_API_BASE = "http://localhost:8040/api"


def resolver_api_base(api_base: str | None) -> str:
    env_base = os.environ.get("TATUSCAN_URL")
    if env_base:
        base_url = env_base.rstrip("/")
        if not base_url.endswith("/api"):
            base_url += "/api"
        return base_url
    return (api_base or DEFAULT_API_BASE).rstrip("/")


def normalizar_numero(valor: str) -> str:
    """Extrai somente os dígitos e remove zeros à esquerda."""
    digitos = "".join(ch for ch in valor if ch.isdigit())
    if not digitos:
        return ""
    normalizado = digitos.lstrip("0")
    return normalizado or "0"


def carregar_relatorio(caminho_csv: Path) -> Dict[str, str]:
    """Carrega o relatório e retorna um índice NUMERO -> DATA DA CARGA."""
    numero_para_data: Dict[str, str] = {}

    with caminho_csv.open(newline="", encoding="utf-8") as arquivo:
        leitor = csv.DictReader(arquivo)
        for linha in leitor:
            numero = normalizar_numero(linha.get("NUMERO", ""))
            if not numero:
                continue

            data_carga = (linha.get("DATA DA CARGA") or "").strip()
            if not data_carga:
                continue

            numero_para_data[numero] = data_carga

    return numero_para_data


def extrair_numero_do_hostname(hostname: str) -> str:
    """Retorna o identificador numérico do hostname (IFMT-1234, m1234, etc.)."""
    if not hostname:
        return ""

    hostname = hostname.strip()

    padroes_preferenciais = (
        r"ifmt[-_]?(\d+)$",
        r"ifmt[-_]?(\d+)",
        r"m(\d+)$",
        r"m(\d+)",
        r"(\d+)$",
    )

    for regex in padroes_preferenciais:
        match = re.search(regex, hostname, re.IGNORECASE)
        if match:
            return normalizar_numero(match.group(1))

    grupos = re.findall(r"\d+", hostname)
    if grupos:
        return normalizar_numero(grupos[-1])

    return ""


def parse_data_para_iso(data: str) -> str | None:
    """Converte string de data em formato ISO (yyyy-mm-dd)."""
    data = data.strip()
    if not data:
        return None

    formatos = ("%d/%m/%Y", "%Y-%m-%d", "%d-%m-%Y")
    for fmt in formatos:
        try:
            dt = datetime.strptime(data, fmt)
            return dt.strftime("%Y-%m-%d")
        except ValueError:
            continue

    logging.warning("Não consegui converter a data '%s' para ISO.", data)
    return None


def normalizar_data_api(valor: str | None) -> str | None:
    """Normaliza a data retornada pelo API para yyyy-mm-dd."""
    if not valor:
        return None

    valor = valor.strip()
    if not valor:
        return None

    if valor.endswith("Z"):
        valor = valor[:-1] + "+00:00"

    try:
        dt = datetime.fromisoformat(valor)
        return dt.date().isoformat()
    except ValueError:
        logging.debug("Data da API em formato inesperado: %s", valor)
        return None


def _request(url: str, *, data: bytes | None = None, method: str = "GET"):
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


def atualizar_por_api(api_base: str, machine_id: str, nova_data_iso: str) -> None:
    url = f"{api_base.rstrip('/')}/machines/{machine_id}"
    body = json.dumps({"computer_activation": nova_data_iso}).encode("utf-8")
    with _request(url, data=body, method="PATCH") as response:
        # le e descarta resposta para garantir que a requisição foi completada
        response.read()


def processar(api_base: str, indice: Dict[str, str]) -> Tuple[int, int, int, int, List[str]]:
    inventarios = fetch_inventory(api_base)

    total_hostnames = len(inventarios)
    hostnames_com_numero = 0
    correspondencias = 0
    atualizados = 0
    erros: List[str] = []

    for item in inventarios:
        hostname = item.get("hostname") or ""
        machine_id = item.get("machine_id")
        atual_api = normalizar_data_api(item.get("computer_activation"))

        numero_hostname = extrair_numero_do_hostname(hostname)
        if not numero_hostname:
            continue

        hostnames_com_numero += 1
        data_relatorio = indice.get(numero_hostname)
        if not data_relatorio:
            continue

        correspondencias += 1
        nova_data_iso = parse_data_para_iso(data_relatorio)
        if not nova_data_iso:
            continue

        if atual_api == nova_data_iso:
            continue

        try:
            atualizar_por_api(api_base, machine_id, nova_data_iso)
            atualizados += 1
        except HTTPError as exc:
            msg = f"HTTPError {exc.code} ao atualizar {machine_id}: {exc.reason}"
            logging.error(msg)
            erros.append(msg)
        except URLError as exc:
            msg = f"URLError ao atualizar {machine_id}: {exc.reason}"
            logging.error(msg)
            erros.append(msg)

    return total_hostnames, hostnames_com_numero, correspondencias, atualizados, erros


def configurar_argumentos() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Atualiza o campo 'computer_activation' via API com base no relatorio.csv"
        )
    )
    parser.add_argument(
        "--csv",
        type=Path,
        default=DEFAULT_CSV_PATH,
        help=f"Caminho para o relatorio.csv (padrão: {DEFAULT_CSV_PATH})",
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

    return parser.parse_args()


def main() -> None:
    args = configurar_argumentos()
    logging.basicConfig(level=getattr(logging, args.log_level))

    if not args.csv.exists():
        raise FileNotFoundError(f"CSV não encontrado: {args.csv}")

    indice_por_numero = carregar_relatorio(args.csv)
    print(f"Total de números carregados: {len(indice_por_numero)}")

    if not indice_por_numero:
        print("Nenhum número válido encontrado no relatório. Nada a fazer.")
        return

    api_base = resolver_api_base(args.api_base)

    (
        total_hostnames,
        hostnames_com_numero,
        correspondencias,
        atualizados,
        erros,
    ) = processar(api_base, indice_por_numero)

    print(f"Hostnames analisados: {total_hostnames}")
    print(f"Hostnames com número identificado: {hostnames_com_numero}")
    print(f"Correspondências encontradas: {correspondencias}")
    print(f"Registros atualizados: {atualizados}")
    if erros:
        print("Ocorreram erros durante a atualização:")
        for linha in erros:
            print(f" - {linha}")


if __name__ == "__main__":
    main()
