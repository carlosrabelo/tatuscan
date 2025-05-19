# TatuScan Deploy

Production packaging for the monorepo components **api** and **web**.

Portuguese: [README-PT.md](README-PT.md).

## Docker Compose

```bash
# from repo root
make stack-start

# API  http://localhost:8040
# Web  http://localhost:8050
```

Compose file: [docker/docker-compose.yml](docker/docker-compose.yml).  
Optional overrides: copy [docker/.env.example](docker/.env.example) to `docker/.env`.

## Kubernetes

```bash
make deploy-k8s
# or: cd deploy && make k8s-deploy
```

Manifests under [k8s/](k8s/): API + Web deployments, services, ConfigMap, PVC, optional Ingress (`/api` → API, `/` → Web).

## Systemd

```bash
sudo deploy/systemd/install.sh
```

Installs `/opt/tatuscan/bin/tatuscan-api` and `tatuscan-web`, units `tatuscan-api.service` and `tatuscan-web.service`.
