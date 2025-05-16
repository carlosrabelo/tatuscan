MAKEFLAGS += --no-print-directory

.DEFAULT_GOAL := help

.PHONY: build clean fmt help lint test

help: ## Show available targets
	@echo "tatuscan tools - Available targets"
	@echo ""
	@grep -hE '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*## "} {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Linux binaries
	@./.make/build.sh

test: ## Run tests
	@./.make/test.sh

lint: ## Run go vet
	@go vet ./...

fmt: ## Format code
	@go fmt ./...

clean: ## Remove build artifacts
	@go clean
	@rm -rf bin/linux/*
	@mkdir -p bin/linux && touch bin/linux/.gitkeep
