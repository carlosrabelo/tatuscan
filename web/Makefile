MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help lint run test

BINARY_NAME := tatuscan-web
export BINARY_NAME

help: ## Show available targets
	@echo "tatuscan web - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Linux binary
	@./.make/build.sh

test: ## Run tests
	@./.make/test.sh

run: ## Run web panel locally
	@./.make/local.sh

lint: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

clean: ## Remove build artifacts
	@go clean
	@rm -rf bin/linux/*
	@mkdir -p bin/linux && touch bin/linux/.gitkeep
