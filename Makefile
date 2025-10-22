# Makefile — TatuScan (Simplified)
# Gerencia build, deploy e execução via scripts

.RECIPEPREFIX := >
SHELL := /usr/bin/env bash

# Directories
SCRIPTS_DIR := scripts
DEPLOY_DIR  := deploy

# Default target
.DEFAULT_GOAL := help
.PHONY: help all clean

# =========================
# Help
# =========================
help:
> @echo ""
> @echo "TatuScan - Build & Deploy"
> @echo ""
> @echo "SETUP:"
> @echo "  setup-venv          - cria e configura .venv para o servidor"
> @echo ""
> @echo "SERVER:"
> @echo "  server-build        - build da imagem Docker"
> @echo "  server-start        - inicia servidor Docker"
> @echo "  server-stop         - para servidor Docker"
> @echo "  server-local        - roda servidor local (.venv)"
> @echo "  server-logs         - mostra logs do container"
> @echo ""
> @echo "CLIENT:"
> @echo "  client-build        - build cliente Linux"
> @echo "  client-build-windows - build cliente Windows"
> @echo "  client-build-all    - build todas plataformas"
> @echo "  client-test         - testa cliente"
> @echo ""
> @echo "DEPLOY:"
> @echo "  deploy-docker       - deploy via Docker"
> @echo "  deploy-k8s          - deploy no Kubernetes"
> @echo "  deploy-systemd      - instala systemd service"
> @echo ""
> @echo "UTILITY:"
> @echo "  clean               - limpa binários e cache Docker"
> @echo "  clean-bin           - limpa apenas binários"
> @echo "  clean-db            - limpa apenas banco de dados (pede confirmação)"
> @echo "  clean-all           - limpa tudo (binários + banco)"
> @echo "  all                 - build completo"
> @echo ""

all: server-build client-build
> @echo "✓ Build completo finalizado!"

# =========================
# SETUP
# =========================
.PHONY: setup-venv

setup-venv:
> @$(SCRIPTS_DIR)/setup-venv.sh

# =========================
# SERVER
# =========================
.PHONY: server-build server-start server-stop server-restart server-logs server-local

server-build:
> @$(SCRIPTS_DIR)/server-build.sh

server-start: server-build
> @$(SCRIPTS_DIR)/server-start.sh

server-stop:
> @$(SCRIPTS_DIR)/server-stop.sh

server-restart: server-stop server-start

server-logs:
> @cd server && docker compose logs -f --tail=200

server-local:
> @$(SCRIPTS_DIR)/server-local.sh

# =========================
# CLIENT
# =========================
.PHONY: client-build client-build-windows client-build-all client-test

client-build:
> @cd client && go mod download
> @$(SCRIPTS_DIR)/client-build.sh linux

client-build-windows:
> @cd client && go mod download
> @$(SCRIPTS_DIR)/client-build.sh windows

client-build-all:
> @cd client && go mod download
> @$(SCRIPTS_DIR)/client-build.sh all

client-test:
> @$(SCRIPTS_DIR)/client-test.sh

# =========================
# DEPLOY
# =========================
.PHONY: deploy-docker deploy-k8s deploy-systemd

deploy-docker:
> @cd $(DEPLOY_DIR) && $(MAKE) quick-docker

deploy-k8s:
> @cd $(DEPLOY_DIR) && $(MAKE) quick-k8s

deploy-systemd:
> @cd $(DEPLOY_DIR) && $(MAKE) quick-systemd

# =========================
# CLEANUP
# =========================
.PHONY: clean clean-bin clean-db clean-all

clean:
> @$(SCRIPTS_DIR)/clean.sh

clean-bin:
> @$(SCRIPTS_DIR)/clean-bin.sh

clean-db:
> @$(SCRIPTS_DIR)/clean-db.sh

clean-all: clean-bin clean-db
> @echo "✓ Full cleanup completed (binaries + database)"
