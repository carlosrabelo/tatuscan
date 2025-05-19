# Ferramentas TatuScan

CLIs HTTP de administração contra a API do TatuScan (`TATUSCAN_URL`, padrão `http://localhost:8040`).

English: [README.md](README.md).

## Destaques

- Três binários Go independentes (sem Python)
- Cliente HTTP compartilhado com fallback `/machines` → `/inventory`
- `TATUSCAN_API_TOKEN` opcional (Bearer) nas escritas

## Uso

```bash
make build
export TATUSCAN_URL=http://localhost:8040
# opcional quando a API exige token nas escritas:
# export TATUSCAN_API_TOKEN=secret

./bin/linux/add-manual-inventory --hostname IFMT-1234 --os "Chrome OS"
./bin/linux/delete-older --dry-run
./bin/linux/update-activation --csv inventario.csv
```

| Binário | Propósito |
|---------|-----------|
| `add-manual-inventory` | POST de uma linha de inventário manual |
| `delete-older` | DELETE de duplicatas antigas por hostname |
| `update-activation` | PATCH de datas de ativação a partir de CSV (`NUMERO`, `DATA DA CARGA`) |

## Configuração

- `TATUSCAN_URL` — host da API (`/api` é acrescentado se faltar)
- `TATUSCAN_API_TOKEN` — token Bearer para POST / PATCH / DELETE
- `--api-base` — sobrescreve a URL sem usar o ambiente
- `TATUSCAN_LANG` — idioma da CLI (`en` ou `pt`; padrão `en`) no `.env`

## Desenvolvimento

```bash
make test
make build
make lint
```
