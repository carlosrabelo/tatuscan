# TatuScan tools

HTTP admin CLIs against the TatuScan API (`TATUSCAN_URL`, default `http://localhost:8040`).

Portuguese: [README-PT.md](README-PT.md).

## Highlights

- Three standalone Go binaries (no Python)
- Shared HTTP client with `/machines` → `/inventory` fallback
- Optional `TATUSCAN_API_TOKEN` Bearer on write operations

## Usage

```bash
make build
export TATUSCAN_URL=http://localhost:8040
# optional when the API requires writes:
# export TATUSCAN_API_TOKEN=secret

./bin/linux/add-manual-inventory --hostname IFMT-1234 --os "Chrome OS"
./bin/linux/delete-older --dry-run
./bin/linux/update-activation --csv inventario.csv
```

| Binary | Purpose |
|--------|---------|
| `add-manual-inventory` | POST a manual inventory row |
| `delete-older` | DELETE older hostname duplicates |
| `update-activation` | PATCH activation dates from CSV (`NUMERO`, `DATA DA CARGA`) |

## Configuration

- `TATUSCAN_URL` — API host ( `/api` is appended when missing)
- `TATUSCAN_API_TOKEN` — Bearer token for POST / PATCH / DELETE
- `--api-base` — override the URL without using the environment
- `TATUSCAN_LANG` — CLI language (`en` or `pt`; default `en`) in `.env`

## Development

```bash
make test
make build
make lint
```
