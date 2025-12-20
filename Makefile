.PHONY: help test test-docker test-dev e2e dev clean down logs

COMPOSE_DEV = docker compose -f docker-compose.yml
COMPOSE_TEST = docker compose -f docker-compose.test.yml
COMPOSE_E2E = docker compose -f docker-compose.e2e.yml

help:
	@echo ""
	@echo "Available targets:"
	@echo "  make test         Run unit + integration tests (no Docker)"
	@echo "  make test-docker  Run canonical integration tests in Docker"
	@echo "  make test-dev     Run tests with bind mount (fast iteration)"
	@echo "  make e2e          Run full end-to-end tests"
	@echo "  make dev          Run API + Postgres for local development"
	@echo "  make logs         Tail dev logs"
	@echo "  make down         Stop dev containers"
	@echo "  make clean        Remove containers, networks, volumes"
	@echo ""

# ---------- Tests ----------

test:
	go test ./internal/... ./cmd/...

test-docker:
	$(COMPOSE_TEST) run --build --rm tests

test-dev:
	$(COMPOSE_TEST) --profile dev run --rm test-dev

e2e:
	$(COMPOSE_E2E) run --build --rm e2e

# ---------- Development ----------

dev:
	$(COMPOSE_DEV) up --build -d

logs:
	$(COMPOSE_DEV) logs -f

down:
	$(COMPOSE_DEV) down

# ---------- Cleanup ----------

clean:
	$(COMPOSE_DEV) down -v --remove-orphans
	$(COMPOSE_TEST) down -v --remove-orphans
	$(COMPOSE_E2E) down -v --remove-orphans
