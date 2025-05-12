MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: api-build api-run api-test build clean clean-all clean-bin clean-db \
	client-build client-build-all client-build-windows client-run client-test help test web-test

help: ## Show available targets
	@echo "tatuscan - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-18s %s\n", $$1, $$2}'

api-test: ## Test API
	@$(MAKE) -C api test

api-build: ## Build API
	@$(MAKE) -C api build

api-run: ## Run API locally
	@$(MAKE) -C api run

client-test: ## Test client
	@$(MAKE) -C client test

client-build: ## Build client (Linux)
	@$(MAKE) -C client build

client-build-windows: ## Build client (Windows)
	@$(MAKE) -C client build-windows

client-build-all: ## Build client (all platforms)
	@$(MAKE) -C client build-all

client-run: ## Run client agent locally
	@$(MAKE) -C client run ARGS='$(ARGS)'

web-test: ## Test web panel
	@$(MAKE) -C web test

build: client-build api-build ## Build client + api
	@echo "✓ Full build completed"

test: client-test api-test web-test ## Test all components

clean: ## Clean binaries and Docker cache
	@./.make/clean.sh

clean-bin: ## Clean binaries only
	@./.make/clean-bin.sh

clean-db: ## Clean local database files
	@./.make/clean-db.sh

clean-all: clean-bin clean-db ## Clean binaries and database
	@echo "✓ Full cleanup completed (binaries + database)"
