# TatuScan

Distributed machine inventory: a Go agent posts host metrics to a Go API, and a separate Go web panel renders dashboards.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Highlights

- Lightweight Go agent for Windows and Linux with optional OS service install
- Stable machine IDs from filtered physical MAC addresses (SHA-256)
- REST API with SQLite persistence and upsert by `machine_id`
- Separate HTML panel for OS, version, and hardware-age views
- English/Portuguese UI (i18n) via `TATUSCAN_LANG` in `.env` (`en` or `pt`)
- Monorepo with independent `client`, `api`, `web`, and `tools` Go modules
- Docker Compose, Kubernetes, and systemd packaging under `deploy/`

## Prerequisites

- **Go 1.24.2+** — language version pinned to the stable release of 2025-05-01; [download](https://go.dev/dl/)
- **Docker** — optional, for `make stack-start`

## Installation

### Build from Source

```bash
git clone https://github.com/carlosrabelo/tatuscan.git
cd tatuscan
make build
```

Binaries land in each component’s `bin/`:

- `client/bin/linux/tatuscan`
- `api/bin/linux/tatuscan-api`
- `web/bin/linux/tatuscan-web`
- `tools/bin/linux/` — `add-manual-inventory`, `delete-older`, `update-activation`

## Quick Start

```bash
make local-start      # API :8040 + panel :8050 (no Docker)
make local-test       # unit tests + HTTP smoke (no Docker)
make client-run       # one-shot inventory POST (other terminal)
# daemon: make client-run ARGS='-d'
# one process at a time: make api-run / make web-run
```

Full stack with Docker:

```bash
make stack-start
# API  http://localhost:8040
# Web  http://localhost:8050
```

## Usage

Agent (default URL `http://127.0.0.1:8040`; `.env` optional):

```bash
make client-run
make client-run ARGS='-d -l info'
# or
./client/bin/linux/tatuscan -d -l info
```

API surface (agents and scripts):

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/machines` | List inventory |
| GET | `/api/inventory` | Alias of `/api/machines` |
| POST | `/api/machines` | Create or update |
| PATCH / DELETE | `/api/machines/{id}` | Patch activation / delete |
| GET | `/api/stats/os` | OS distribution |
| GET | `/api/stats/versions?top=8` | Version distribution |
| GET | `/api/stats/age` | Age buckets |
| GET | `/api/stats/online?after=2h` | Online/offline counts |

Panel routes: `/`, `/report/`, `/charts/`.

Admin tools (same API; honor `TATUSCAN_URL` and `TATUSCAN_API_TOKEN`):

```bash
make tools-build
./tools/bin/linux/add-manual-inventory --hostname IFMT-1234
./tools/bin/linux/delete-older --dry-run
./tools/bin/linux/update-activation --csv inventario.csv
```

## Configuration

**Client** — `TATUSCAN_URL` (required), `TATUSCAN_INTERVAL` (default `60s`), optional `TATUSCAN_API_TOKEN` (Bearer for writes).

**API** — `TATUSCAN_PORT` (`8040`), `TATUSCAN_DB_DIR` / `TATUSCAN_DB_FILE` (default `/tmp/tatuscan.db`; Docker sets `/data`), `TIMEZONE` (`America/Cuiaba`), optional `TATUSCAN_API_TOKEN` (protects POST/PATCH/DELETE), `TATUSCAN_OFFLINE_AFTER` (default `2h`). `.env` is optional — missing file means built-in defaults. On startup the API rewrites legacy Flask/SQLAlchemy datetime strings to RFC3339Nano (idempotent); skip with `TATUSCAN_SKIP_LEGACY_DATETIME_MIGRATE=1` when the transition is done.

**Web** — `TATUSCAN_PORT` (`8050`), `TATUSCAN_API_URL` (`http://127.0.0.1:8040`), `TATUSCAN_OFFLINE_AFTER` (default `2h`), `TATUSCAN_LANG` (`en` or `pt`; default `en`) in `.env`.

**API** — user-facing JSON/HTML messages use `TATUSCAN_LANG` from `.env` (`en` or `pt`; default `en`).

**Tools** — CLI messages use `TATUSCAN_LANG` from `.env` (`en` or `pt`; default `en`).

See each component’s `.env.example`.

## Project Layout

```
tatuscan/
├── Makefile          # orchestrator only
├── client/           # agent (go.mod + client/ source tree)
├── api/              # REST + SQLite (go.mod + api/ source tree)
├── web/              # HTML panel (go.mod + web/ source tree)
├── tools/            # admin CLIs (go.mod + tools/ source tree)
├── deploy/           # Docker / k8s / systemd
└── .make/            # root helpers (local, clean, clean-bin, clean-db)
```

Each Go component keeps infrastructure at its root (`Makefile`, `.make/`, `bin/`, `go.mod`) and all source under a same-named package directory (`client/`, `api/`, `web/`, `tools/`). The git root also has `.make/` for monorepo-wide scripts.

## Development

```bash
make help             # list orchestrator targets
make build            # build client + api + web + tools
make test             # client + api + web + tools
make local-start      # API+web locally (no Docker)
make local-test       # unit tests + HTTP smoke (no Docker)
make local-stop       # stop processes started by make local-start
make api-run          # run API only
make web-run          # run panel only
make client-run       # run agent locally
make stack-start      # Docker Compose api+web
make -C client build  # component-scoped build
make -C api test
make -C web fmt
make -C tools test
```

Deploy details: [deploy/README.md](deploy/README.md).

Portuguese docs: [README-PT.md](README-PT.md).

## License

This project is licensed under the GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
