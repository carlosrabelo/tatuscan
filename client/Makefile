MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: clean fmt help lint test

BINARY_NAME := tatuscan
export BINARY_NAME

help: ## Show available targets
	@echo "tatuscan client - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

test: ## Run tests
	@./.make/test.sh

lint: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

clean: ## Remove build artifacts
	@go clean
	@rm -rf bin/linux/* bin/windows/*
	@touch bin/linux/.gitkeep bin/windows/.gitkeep 2>/dev/null || true
