# TatuScan Deploy

Empacotamento de produção para os componentes do monorepo **api** e **web**.

English: [README.md](README.md).

## Docker Compose

```bash
# from repo root
make stack-start

# API  http://localhost:8040
# Web  http://localhost:8050
```

Arquivo Compose: [docker/docker-compose.yml](docker/docker-compose.yml).  
Overrides opcionais: copie [docker/.env.example](docker/.env.example) para `docker/.env`.

## Kubernetes

```bash
make deploy-k8s
# or: cd deploy && make k8s-deploy
```

Manifestos em [k8s/](k8s/): deployments da API e do Web, services, ConfigMap, PVC, Ingress opcional (`/api` → API, `/` → Web).

## Systemd

```bash
sudo deploy/systemd/install.sh
```

Instala `/opt/tatuscan/bin/tatuscan-api` e `tatuscan-web`, units `tatuscan-api.service` e `tatuscan-web.service`.
