# TatuScan

Distributed machine inventory.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Highlights

- Lightweight Go agent for Windows and Linux with optional OS service install
- Stable machine IDs from filtered physical MAC addresses (SHA-256)
- REST API with SQLite persistence and upsert by `machine_id`
- Monorepo with independent `client` and `api` Go modules
- Separate HTML panel (in progress)

Binaries land in each component’s `bin/`:

- `client/bin/linux/tatuscan`
- `api/bin/linux/tatuscan-api`

## Usage

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

## Project Layout

```
tatuscan/
├── Makefile
├── client/
├── api/
├── web/
└── .make/
```

Portuguese docs: [README-PT.md](README-PT.md).

## License

This project is licensed under the GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
