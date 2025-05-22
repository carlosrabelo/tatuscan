MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build build-all build-windows clean deb fmt help install lint package quality run test uninstall

BINARY_NAME := tatuscan
export BINARY_NAME

help: ## Show available targets
	@echo "tatuscan client - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Linux binary
	@./.make/build.sh linux

build-windows: ## Build Windows binary
	@./.make/build.sh windows

build-all: ## Build all platforms
	@./.make/build.sh all

run: ## Run agent locally (one-shot; pass -d for daemon)
	@./.make/local.sh $(ARGS)

test: ## Run tests
	@./.make/test.sh

lint: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

quality: fmt lint ## Run all quality checks

clean: ## Remove build artifacts
	@go clean
	@rm -rf bin/linux/* bin/windows/* dist/
	@touch bin/linux/.gitkeep bin/windows/.gitkeep 2>/dev/null || true

install: ## Install to ~/.local/bin (requires prior make build)
	@./.make/install.sh

uninstall: ## Remove from ~/.local/bin (sudo only if needed)
	@./.make/uninstall.sh

package: ## Build a .deb into dist/ (requires prior make build; installs + enables daemon)
	@./.make/package.sh

deb: package ## Alias for package
