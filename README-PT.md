# TatuScan

Inventário distribuído de máquinas.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Destaques

- Lightweight Go agent for Windows and Linux with optional OS service install
- Stable machine IDs from filtered physical MAC addresses (SHA-256)
- REST API with SQLite persistence and upsert by `machine_id`
- Separate HTML panel for OS, version, and hardware-age views
- English/Portuguese UI (i18n) via `TATUSCAN_LANG` in `.env` (`en` or `pt`)
- Monorepo with independent `client`, `api`, and `web` Go modules

## Estrutura do Projeto

```
tatuscan/
├── Makefile
├── client/
├── api/
├── web/
├── tools/
└── .make/
```

English docs: [README.md](README.md).

## Licença

Este projeto está licenciado sob a GNU General Public License v3.0 — veja [LICENSE](LICENSE) para detalhes.
