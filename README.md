# TatuScan

Distributed machine inventory.

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

## Highlights

- Monorepo with an independent `api` Go module

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

## Project Layout

```
tatuscan/
├── Makefile
├── api/
└── .make/
```

Portuguese docs: [README-PT.md](README-PT.md).

## License

This project is licensed under the GNU General Public License v3.0 — see [LICENSE](LICENSE) for details.
