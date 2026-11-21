SHELL := /usr/bin/env bash
BINARY_NAME := layout-engine
BIN_DIR := bin
CMD_PKG := ./cmd/layout-engine
GO ?= go

.PHONY: help run build test test-race cover vet fmt fmt-check hygiene verify tidy docker-build docker-up docker-down clean

help: ## Show available Make targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nAbeta Dynamic Layout Engine — Make Targets:\n\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } END { printf "\n" }' $(MAKEFILE_LIST)

run: ## Run the layout engine locally on :8090 (zero-config SQLite + starter seed)
	$(GO) run $(CMD_PKG)

build: ## Build a static local binary into bin/layout-engine
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PKG)

test: ## Run unit and integration tests
	$(GO) test -count=1 ./...

test-race: ## Run tests with race detector and coverage summary
	$(GO) test -count=1 -race -cover ./...

cover: ## Generate HTML coverage report in coverage.html
	$(GO) test -count=1 -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

vet: ## Run go vet static analysis
	$(GO) vet ./...

fmt: ## Format all Go source files with gofmt
	gofmt -w .

fmt-check: ## Verify all Go source files are gofmt-clean
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted Go files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

hygiene: ## Run brand and machine-path hygiene checks
	./scripts/check_brand_hygiene.sh

verify: ## Run the full local verification pipeline
	./scripts/verify_all.sh

tidy: ## Synchronize go.mod and go.sum
	$(GO) mod tidy

docker-build: ## Build the local OCI container image
	docker build -t abeta-layout-engine:local .

docker-up: ## Start PostgreSQL 17 and the layout engine via Docker Compose
	docker compose up --build

docker-down: ## Stop Docker Compose services
	docker compose down

clean: ## Remove build outputs, local coverage files, and temporary SQLite databases
	rm -rf $(BIN_DIR) coverage.out coverage.html layout-engine.db layout-engine.db-wal layout-engine.db-shm
