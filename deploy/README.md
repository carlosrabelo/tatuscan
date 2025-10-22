# TatuScan Deployment Options

This directory contains deployment configurations for different environments and platforms.

## Directory Structure

```
deploy/
├── docker/             # Production Docker deployment
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── entrypoint.sh
│   └── .dockerignore
├── k8s/               # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── pvc.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
├── systemd/           # Systemd service files for Linux
│   ├── tatuscan.service
│   ├── tatuscan@.service
│   ├── tatuscan.socket
│   ├── install.sh
│   └── uninstall.sh
├── Makefile           # Deployment automation
└── README.md          # This file

Note: For local development, use server/docker-compose.yml instead.
```

## Docker Deployment

### Development (Recommended)
For local development, use the server's docker-compose:
```bash
# From project root
make server-start
# or
cd server
docker compose up -d
```

### Production Deployment
For production deployment:
```bash
# From project root
make deploy-docker
# or
cd deploy/docker
docker compose up -d
```

### Environment Variables
Create `.env` file in the server directory based on `server/.env.example`:
- `TATUSCAN_PORT`: Server port (default: 8040)
- `HOST_PORT`: External port mapping (default: 8040)
- `TATUSCAN_DB_DIR`: Database directory (default: /data)
- `TATUSCAN_DB_FILE`: Database filename (default: tatuscan.db)
- Database configuration for PostgreSQL/MySQL

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster (minikube, k3s, or cloud provider)
- kubectl configured
- Ingress controller (nginx, traefik, etc.)

### Deployment Steps

1. **Create namespace and apply configurations:**
```bash
cd deploy/k8s
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

2. **Configure Ingress (optional):**
```bash
# Edit ingress.yaml with your domain
kubectl apply -f ingress.yaml
```

3. **Enable autoscaling (optional):**
```bash
kubectl apply -f hpa.yaml
```

### Monitoring and Management
```bash
# Check deployment status
kubectl get all -n tatuscan

# View logs
kubectl logs -f deployment/tatuscan -n tatuscan

# Scale deployment
kubectl scale deployment tatuscan --replicas=3 -n tatuscan

# Port forward for testing
kubectl port-forward service/tatuscan-service 8040:8040 -n tatuscan
```

### Configuration
- Edit `configmap.yaml` for environment variables
- Edit `secret.yaml` for sensitive data (database URLs, keys)
- Adjust `pvc.yaml` for storage requirements
- Modify `ingress.yaml` for your domain and TLS

## Systemd Deployment

### Prerequisites
- Linux system with systemd
- Python 3.8+
- Root access

### Installation
```bash
cd deploy/systemd
sudo ./install.sh
```

### Service Management
```bash
# Check status
systemctl status tatuscan.socket

# View logs
journalctl -u tatuscan@.service -f

# Restart service
sudo systemctl restart tatuscan@.service

# Stop service
sudo systemctl stop tatuscan.socket

# Enable at boot
sudo systemctl enable tatuscan.socket
```

### Uninstallation
```bash
cd deploy/systemd
sudo ./uninstall.sh
```

### Configuration
Service configuration is handled through environment variables in the service file:
- `/opt/tatuscan` - Installation directory
- `/var/lib/tatuscan` - Data directory
- `/var/log/tatuscan` - Log directory
- Port: 8040
- User: `tatuscan`

## Security Considerations

### Docker
- Use specific image tags instead of `latest`
- Implement resource limits
- Use read-only filesystem where possible
- Run as non-root user

### Kubernetes
- Enable NetworkPolicies
- Use PodSecurityPolicies
- Implement RBAC
- Use secrets for sensitive data
- Enable TLS termination

### Systemd
- Services run as dedicated user `tatuscan`
- Filesystem restrictions enabled
- No new privileges allowed
- Private temporary directories

## Monitoring

### Health Checks
All deployments include health checks at `/api/health`

### Metrics
- Docker: Container stats
- Kubernetes: HPA metrics, resource usage
- Systemd: Journal logs, service status

### Logging
- Docker: `docker-compose logs`
- Kubernetes: `kubectl logs`
- Systemd: `journalctl`

## Troubleshooting

### Docker Issues
```bash
# Check container logs
docker-compose logs tatuscan

# Rebuild image
docker-compose build --no-cache

# Check network
docker network ls
```

### Kubernetes Issues
```bash
# Describe pod
kubectl describe pod -n tatuscan

# Check events
kubectl get events -n tatuscan --sort-by=.metadata.creationTimestamp

# Debug pod
kubectl exec -it deployment/tatuscan -n tatuscan -- /bin/bash
```

### Systemd Issues
```bash
# Check service status
systemctl status tatuscan@.service

# View detailed logs
journalctl -u tatuscan@.service -n 100

# Check configuration
systemd-analyze verify tatuscan.service
```

## Production Considerations

1. **Database**: Use PostgreSQL/MySQL instead of SQLite for production
2. **Backup**: Implement regular database backups
3. **SSL/TLS**: Enable HTTPS in production
4. **Monitoring**: Set up proper monitoring and alerting
5. **Updates**: Plan for zero-downtime deployments
6. **Scaling**: Configure horizontal scaling based on load

## Support

For deployment issues:
1. Check logs for error messages
2. Verify configuration files
3. Ensure all prerequisites are met
4. Consult the main project documentation