# TatuScan tools

HTTP admin CLIs against the TatuScan API (`TATUSCAN_URL`, default `http://localhost:8040`).

Portuguese: [README-PT.md](README-PT.md).

## Highlights

- Standalone Go binaries (no Python)
- Shared HTTP client with `/machines` → `/inventory` fallback
- Optional `TATUSCAN_API_TOKEN` Bearer on write operations

## Usage

```bash
make build
export TATUSCAN_URL=http://localhost:8040

./bin/linux/add-manual-inventory --hostname IFMT-1234 --os "Chrome OS"
```

| Binary | Purpose |
|--------|---------|
| `add-manual-inventory` | POST a manual inventory row |
