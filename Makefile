MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help


help: ## Show available targets
	@echo "tatuscan - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-18s %s\n", $$1, $$2}'

.PHONY: clean clean-all clean-bin clean-db help

clean: ## Clean binaries and Docker cache
	@./.make/clean.sh

clean-bin: ## Clean binaries only
	@./.make/clean-bin.sh

clean-db: ## Clean local database files
	@./.make/clean-db.sh

clean-all: clean-bin clean-db ## Clean binaries and database
	@echo "✓ Full cleanup completed (binaries + database)"
