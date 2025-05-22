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
- **dpkg** — opcional, para `make client-package`

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

### Agente (usuário local)

```bash
make client-build
make client-install          # ~/.local/bin
make client-uninstall
```

### Agente (sistema via .deb + daemon)

```bash
make client-build
make client-package          # client/dist/*.deb
sudo dpkg -i client/dist/tatuscan_*.deb
# o postinst habilita e inicia tatuscan.service
```

Edite `/etc/tatuscan/agent.env` (conffile) para `TATUSCAN_URL` se necessário e rode `sudo systemctl restart tatuscan`.

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

Ferramentas admin (mesma API; respeitam `TATUSCAN_URL` e `TATUSCAN_API_TOKEN`):

```bash
make tools-build
./tools/bin/linux/add-manual-inventory --hostname IFMT-1234
./tools/bin/linux/delete-older --dry-run
./tools/bin/linux/update-activation --csv inventario.csv
```

## Configuração

**Client** — `TATUSCAN_URL` (obrigatório), `TATUSCAN_INTERVAL` (default `15m`), `TATUSCAN_API_TOKEN` opcional (Bearer para escritas).

**API** — `TATUSCAN_PORT` (`8040`), `TATUSCAN_DB_DIR` / `TATUSCAN_DB_FILE` (default `/tmp/tatuscan.db`; Docker usa `/data`), `TIMEZONE` (`America/Cuiaba`), `TATUSCAN_API_TOKEN` opcional (protege POST/PATCH/DELETE), `TATUSCAN_OFFLINE_AFTER` (default `2h`). `.env` é opcional — arquivo ausente usa defaults. Na subida a API reescreve datetimes legados Flask/SQLAlchemy para RFC3339Nano (idempotente); pule com `TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE=1` quando a transição terminar.

**Web** — `TATUSCAN_PORT` (`8050`), `TATUSCAN_API_URL` (`http://127.0.0.1:8040`), `TATUSCAN_OFFLINE_AFTER` (default `2h`), `TATUSCAN_LANG` (`en` ou `pt`; default `en`) no `.env`.

**API** — mensagens JSON/HTML usam `TATUSCAN_LANG` do `.env` (`en` ou `pt`; default `en`).

**Tools** — mensagens CLI usam `TATUSCAN_LANG` do `.env` (`en` ou `pt`; default `en`).

Veja o `.env.example` de cada componente.

## Layout do Projeto

```
tatuscan/
├── Makefile          # só orquestrador
├── AGENTS.md         # contratos para agentes/scaffold
├── client/           # agente (go.mod + árvore client/)
├── api/              # REST + SQLite (go.mod + árvore api/)
├── web/              # painel HTML (go.mod + árvore web/)
├── tools/            # CLIs admin (go.mod + árvore tools/)
├── deploy/           # Docker / k8s / systemd (api+web)
└── .make/            # helpers da raiz (local, clean, clean-bin, clean-db)
```

Cada componente Go mantém infra na raiz (`Makefile`, `.make/`, `bin/`, `go.mod`) e o código na pasta com o mesmo nome. A raiz do git também tem `.make/` para scripts do monorepo.

## Desenvolvimento

```bash
make help             # lista targets do orquestrador
make build            # build client + api + web + tools
make test             # client + api + web + tools
make quality          # fmt + lint em todos os componentes Go
make client-install   # agente → ~/.local/bin
make client-package   # .deb do agente (unit + enable --now)
make local-start      # API+web local (sem Docker)
make local-test       # testes de unidade + smoke HTTP (sem Docker)
make local-stop       # para processos do make local-start
make api-run          # só API
make web-run          # só painel
make client-run       # agente local
make stack-start      # Docker Compose api+web
make -C client build  # build no componente
make -C api test
make -C web fmt
make -C tools test
```

Deploy (api+web): [deploy/README-PT.md](deploy/README-PT.md).

English docs: [README.md](README.md).

## Licença

Este projeto está sob a GNU General Public License v3.0 — veja [LICENSE](LICENSE).
