# TatuScan

Inventário distribuído de máquinas: um agente Go envia métricas do host para uma API Go, e um painel web Go separado renderiza os dashboards.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Destaques

- Agente Go leve para Windows e Linux, com instalação opcional como serviço
- IDs estáveis a partir de MACs físicos filtrados (SHA-256)
- API REST com SQLite e upsert por `machine_id`
- Painel HTML separado para SO, versões e idade de hardware
- Interface inglês/português (i18n) via `TATUSCAN_LANG` no `.env` (`en` ou `pt`)
- Monorepo com módulos Go independentes `client`, `api`, `web` e `tools`
- Empacotamento Docker Compose, Kubernetes e systemd em `deploy/`

## Pré-requisitos

- **Go 1.24.2+** — versão de linguagem fixada na estável de 2025-05-01; [download](https://go.dev/dl/)
- **Docker** — opcional, para `make stack-start`

## Instalação

### Build a partir do código

```bash
git clone https://github.com/carlosrabelo/tatuscan.git
cd tatuscan
make build
```

Binários em cada componente:

- `client/bin/linux/tatuscan`
- `api/bin/linux/tatuscan-api`
- `web/bin/linux/tatuscan-web`
- `tools/bin/linux/` — `add-manual-inventory`, `delete-older`, `update-activation`

## Início Rápido

```bash
make local-start      # API :8040 + painel :8050 (sem Docker)
make local-test       # testes de unidade + smoke HTTP (sem Docker)
make client-run       # POST de inventário (one-shot, outro terminal)
# daemon: make client-run ARGS='-d'
# um processo por vez: make api-run / make web-run
```

Stack completa com Docker:

```bash
make stack-start
# API  http://localhost:8040
# Web  http://localhost:8050
```

## Uso

Agente (URL default `http://127.0.0.1:8040`; `.env` opcional):

```bash
make client-run
make client-run ARGS='-d -l info'
# ou
./client/bin/linux/tatuscan -d -l info
```

API (agentes e scripts):

| Método | Path | Descrição |
|--------|------|-----------|
| GET | `/api/health` | Healthcheck |
| GET | `/api/machines` | Lista inventário |
| GET | `/api/inventory` | Alias de `/api/machines` |
| POST | `/api/machines` | Cria ou atualiza |
| PATCH / DELETE | `/api/machines/{id}` | Atualiza ativação / remove |
| GET | `/api/stats/os` | Distribuição por SO |
| GET | `/api/stats/versions?top=8` | Distribuição por versão |
| GET | `/api/stats/age` | Faixas de idade |
| GET | `/api/stats/online?after=2h` | Contagens online/offline |

Rotas do painel: `/`, `/report/`, `/charts/`.

Ferramentas de administração (mesma API; respeitam `TATUSCAN_URL` e `TATUSCAN_API_TOKEN`):

```bash
make tools-build
./tools/bin/linux/add-manual-inventory --hostname IFMT-1234
./tools/bin/linux/delete-older --dry-run
./tools/bin/linux/update-activation --csv inventario.csv
```

## Configuração

**Cliente** — `TATUSCAN_URL` (obrigatório), `TATUSCAN_INTERVAL` (padrão `60s`), `TATUSCAN_API_TOKEN` opcional (Bearer para escrita).

**API** — `TATUSCAN_PORT` (`8040`), `TATUSCAN_DB_DIR` / `TATUSCAN_DB_FILE` (default `/tmp/tatuscan.db`; no Docker fica `/data`), `TIMEZONE` (`America/Cuiaba`), `TATUSCAN_API_TOKEN` opcional (protege POST/PATCH/DELETE), `TATUSCAN_OFFLINE_AFTER` (padrão `2h`). `.env` é opcional — sem o arquivo valem os defaults do binário. Na subida, a API regrava datas legadas do Flask/SQLAlchemy para RFC3339Nano (idempotente); desligue com `TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE=1` quando a transição terminar.

**Web** — `TATUSCAN_PORT` (`8050`), `TATUSCAN_API_URL` (`http://127.0.0.1:8040`), `TATUSCAN_OFFLINE_AFTER` (padrão `2h`), `TATUSCAN_LANG` (`en` ou `pt`; padrão `en`) no `.env`.

**API** — mensagens JSON/HTML usam `TATUSCAN_LANG` do `.env` (`en` ou `pt`; padrão `en`).

**Ferramentas** — mensagens da CLI usam `TATUSCAN_LANG` do `.env` (`en` ou `pt`; padrão `en`).

Veja o `.env.example` de cada componente.

## Estrutura do Projeto

```
tatuscan/
├── Makefile          # orquestrador apenas
├── client/           # agente (go.mod + árvore client/)
├── api/              # REST + SQLite (go.mod + árvore api/)
├── web/              # painel HTML (go.mod + árvore web/)
├── tools/            # CLIs de administração (go.mod + árvore tools/)
├── deploy/           # Docker / k8s / systemd
└── .make/            # helpers da raiz (local, clean, clean-bin, clean-db)
```

Cada componente Go mantém infraestrutura na raiz (`Makefile`, `.make/`, `bin/`, `go.mod`) e todo o código-fonte numa pasta homônima (`client/`, `api/`, `web/`, `tools/`). A raiz do git também tem `.make/` para scripts do monorepo.

## Desenvolvimento

```bash
make help             # lista alvos do orquestrador
make build            # build client + api + web + tools
make test             # client + api + web + tools
make local-start      # API+web local (sem Docker)
make local-test       # testes de unidade + smoke HTTP (sem Docker)
make local-stop       # para processos iniciados por make local-start
make api-run          # só a API
make web-run          # só o painel
make client-run       # agente local
make stack-start      # Docker Compose api+web
make -C client build  # build de um componente
make -C api test
make -C web fmt
make -C tools test
```

Detalhes de deploy: [deploy/README.md](deploy/README.md).

Documentação em inglês: [README.md](README.md).

## Licença

Este projeto está sob a GNU General Public License v3.0 — veja [LICENSE](LICENSE) para detalhes.
