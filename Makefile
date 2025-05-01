MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: help

help: ## Show available targets
	@echo "tatuscan - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-18s %s\n", $$1, $$2}'
