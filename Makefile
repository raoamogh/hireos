.DEFAULT_GOAL := help

GO ?= go
DOCKER_COMPOSE ?= docker compose
SQLC ?= sqlc

.PHONY: help test test-integration build fmt vet sqlc-generate \
        migrate-up migrate-down migrate-version \
        infra-up infra-down

help: ## Show available development commands
	@echo ""
	@echo "HIREOS development commands"
	@echo "--------------------------"
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  make %-20s %s\n", $$1, $$2}'
	@echo ""

test: ## Run all Go tests
	$(GO) test ./...

test-integration: ## Run PostgreSQL integration tests against hireos_test
	@test -f .env || (echo "Error: .env file not found."; exit 1)
	@set -a && . ./.env && set +a && \
		POSTGRES_DB=hireos_test $(GO) test -count=1 -tags=integration ./internal/identity/postgres/...

build: ## Compile all Go packages
	$(GO) build ./...

fmt: ## Format all Go source files
	@gofmt -w $$(find . -type f -name '*.go' -not -path './.git/*' -not -path './vendor/*')

vet: ## Run Go static analysis
	$(GO) vet ./...

sqlc-generate: ## Generate Go database code from SQL queries
	$(SQLC) generate

migrate-up: ## Apply all pending database migrations
	@test -f .env || (echo "Error: .env file not found. Copy .env.example to .env and configure it."; exit 1)
	@set -a && . ./.env && set +a && $(GO) run ./cmd/migrate up

migrate-down: ## Roll back exactly one database migration
	@test -f .env || (echo "Error: .env file not found. Copy .env.example to .env and configure it."; exit 1)
	@echo "Warning: this rolls back exactly one migration."
	@set -a && . ./.env && set +a && $(GO) run ./cmd/migrate down

migrate-version: ## Show the current database migration version
	@test -f .env || (echo "Error: .env file not found. Copy .env.example to .env and configure it."; exit 1)
	@set -a && . ./.env && set +a && $(GO) run ./cmd/migrate version

infra-up: ## Start PostgreSQL, Redis, and MinIO
	$(DOCKER_COMPOSE) up -d

infra-down: ## Stop HIREOS Docker services without deleting volumes
	$(DOCKER_COMPOSE) down
