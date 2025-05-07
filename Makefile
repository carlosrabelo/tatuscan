MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: api-build api-run api-test build clean \
	clean-all clean-bin clean-db client-test help test

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

build: api-build ## Build api
	@echo "✓ Full build completed"

test: client-test api-test ## Test all components

clean: ## Clean binaries and Docker cache
	@./.make/clean.sh

clean-bin: ## Clean binaries only
	@./.make/clean-bin.sh

clean-db: ## Clean local database files
	@./.make/clean-db.sh

clean-all: clean-bin clean-db ## Clean binaries and database
	@echo "✓ Full cleanup completed (binaries + database)"
