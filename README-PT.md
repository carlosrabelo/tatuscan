# TatuScan

Inventário distribuído de máquinas.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Destaques

- Lightweight Go agent types for host inventory
- REST API with SQLite persistence and upsert by `machine_id`
- Monorepo with independent `client` and `api` Go modules

## Estrutura do Projeto

```
tatuscan/
├── Makefile
├── client/
├── api/
└── .make/
```

English docs: [README.md](README.md).

## Licença

Este projeto está licenciado sob a GNU General Public License v3.0 — veja [LICENSE](LICENSE) para detalhes.
