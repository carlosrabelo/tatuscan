# Makefile — TatuScan Deployment
# This Makefile orchestrates deployment across different platforms

.RECIPEPREFIX := >
.SHELL := /usr/bin/env bash

# Default target
.DEFAULT_GOAL := help
.PHONY: help docker-build docker-up docker-down docker-logs k8s-deploy k8s-delete k8s-status systemd-install systemd-uninstall clean

# =========================
# Help
# =========================
help:
> @echo ""
> @echo "TatuScan Deployment Commands:"
> @echo ""
> @echo "Docker Commands:"
> @echo "  docker-build    - Build Docker image"
> @echo "  docker-up       - Start services with Docker Compose"
> @echo "  docker-down     - Stop Docker services"
> @echo "  docker-logs     - Show Docker logs"
> @echo "  docker-clean    - Clean Docker artifacts"
> @echo ""
> @echo "Kubernetes Commands:"
> @echo "  k8s-deploy      - Deploy to Kubernetes"
> @echo "  k8s-delete      - Delete Kubernetes deployment"
> @echo "  k8s-status      - Show Kubernetes status"
> @echo "  k8s-logs        - Show Kubernetes logs"
> @echo "  k8s-port-forward - Port forward to local"
> @echo ""
> @echo "Systemd Commands:"
> @echo "  systemd-install - Install systemd service"
> @echo "  systemd-uninstall - Uninstall systemd service"
> @echo "  systemd-status  - Check systemd service status"
> @echo ""
> @echo "Utility Commands:"
> @echo "  clean           - Clean all deployment artifacts"
> @echo "  help            - Show this help message"
> @echo ""

# =========================
# Docker Commands
# =========================
docker-build:
> @echo "Building Docker image..."
> cd docker && docker-compose build

docker-up:
> @echo "Starting Docker services..."
> cd docker && docker-compose up -d

docker-down:
> @echo "Stopping Docker services..."
> cd docker && docker-compose down

docker-logs:
> @echo "Showing Docker logs..."
> cd docker && docker-compose logs -f

docker-clean:
> @echo "Cleaning Docker artifacts..."
> cd docker && docker-compose down -v
> docker system prune -f

# =========================
# Kubernetes Commands
# =========================
k8s-deploy:
> @echo "Deploying to Kubernetes..."
> kubectl apply -f k8s/namespace.yaml
> kubectl apply -f k8s/configmap.yaml
> kubectl apply -f k8s/secret.yaml
> kubectl apply -f k8s/pvc.yaml
> kubectl apply -f k8s/deployment.yaml
> kubectl apply -f k8s/service.yaml
> @echo "Kubernetes deployment completed!"
> @echo "To enable ingress: kubectl apply -f k8s/ingress.yaml"

k8s-delete:
> @echo "Deleting Kubernetes deployment..."
> kubectl delete -f k8s/ingress.yaml --ignore-not-found=true
> kubectl delete -f k8s/service.yaml --ignore-not-found=true
> kubectl delete -f k8s/deployment.yaml --ignore-not-found=true
> kubectl delete -f k8s/pvc.yaml --ignore-not-found=true
> kubectl delete -f k8s/secret.yaml --ignore-not-found=true
> kubectl delete -f k8s/configmap.yaml --ignore-not-found=true
> kubectl delete -f k8s/namespace.yaml --ignore-not-found=true

k8s-status:
> @echo "Kubernetes deployment status:"
> kubectl get all -n tatuscan

k8s-logs:
> @echo "Kubernetes pod logs:"
> kubectl logs -f deployment/tatuscan -n tatuscan

k8s-port-forward:
> @echo "Port forwarding TatuScan service to localhost:8040..."
> kubectl port-forward service/tatuscan-service 8040:8040 -n tatuscan

# =========================
# Systemd Commands
# =========================
systemd-install:
> @echo "Installing systemd service..."
> cd systemd && sudo ./install.sh

systemd-uninstall:
> @echo "Uninstalling systemd service..."
> cd systemd && sudo ./uninstall.sh

systemd-status:
> @echo "Systemd service status:"
> systemctl status tatuscan.socket || true
> journalctl -u tatuscan@.service -n 20 --no-pager || true

# =========================
# Utility Commands
# =========================
clean:
> @echo "Cleaning all deployment artifacts..."
> # Docker cleanup
> cd docker && docker-compose down -v 2>/dev/null || true
> docker system prune -f
> # Kubernetes cleanup (commented for safety)
> # kubectl delete -f k8s/ --ignore-not-found=true
> @echo "Cleanup completed!"

# Quick deployment shortcuts
.PHONY: quick-docker quick-k8s quick-systemd

quick-docker: docker-build docker-up
> @echo "Docker deployment completed!"
> @echo "Access TatuScan at: http://localhost:8040"

quick-k8s: k8s-deploy k8s-status
> @echo "Kubernetes deployment completed!"
> @echo "Use 'make k8s-port-forward' to access locally"

quick-systemd: systemd-install systemd-status
> @echo "Systemd deployment completed!"
> @echo "Access TatuScan at: http://localhost:8040"