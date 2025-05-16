# Ferramentas TatuScan

CLIs HTTP de administração contra a API TatuScan (`TATUSCAN_URL`, padrão `http://localhost:8040`).

English: [README.md](README.md).

## Destaques

- Binários Go independentes (sem Python)
- Cliente HTTP compartilhado com fallback `/machines` → `/inventory`
- `TATUSCAN_API_TOKEN` opcional (Bearer) nas escritas

## Uso

```bash
make build
export TATUSCAN_URL=http://localhost:8040

./bin/linux/add-manual-inventory --hostname IFMT-1234 --os "Chrome OS"
./bin/linux/delete-older --dry-run
```

| Binário | Propósito |
|---------|-----------|
| `add-manual-inventory` | POST de uma linha de inventário manual |
| `delete-older` | DELETE de duplicatas mais antigas do hostname |
