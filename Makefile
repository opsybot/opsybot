SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

ifneq (,$(wildcard .env))
include .env
export
endif

COMPOSE := docker compose
DC_LONG_RUNNING := postgres valkey garage

MIGRATIONS_DIR := db/migrations/postgres

.PHONY: help env require-env infra infra-up infra-down infra-restart infra-reset infra-ps infra-logs psql migration gen wire db-gen

help:
	@printf "Targets:\n"
	@printf "  make env             Create .env from .env.example (no-op if .env exists)\n"
	@printf "  make infra           Bring the local dev stack up (alias for infra-up)\n"
	@printf "  make infra-up        docker compose up -d (postgres, valkey, garage)\n"
	@printf "  make infra-down      Stop the stack, keep volumes\n"
	@printf "  make infra-restart   Restart all services\n"
	@printf "  make infra-reset     docker compose down -v (DESTROYS local data)\n"
	@printf "  make infra-ps        Show service status\n"
	@printf "  make infra-logs      Tail logs for all services\n"
	@printf "  make psql            Open a psql shell on the dev database\n"
	@printf "  make migration name=create_users\n"
	@printf "                       Create a new goose migration (never hand-write one)\n"
	@printf "  make gen             Run all go:generate directives (oapi-codegen, mockgen)\n"
	@printf "  make wire            Regenerate the DI graph\n"
	@printf "  make db-gen          Regenerate sqlboiler DAO models from the local database\n"

env:
	@if [ -f .env ]; then \
		echo ".env already exists; not overwriting"; \
	else \
		cp .env.example .env && echo "Created .env from .env.example."; \
	fi

require-env:
	@if [ ! -f .env ]; then \
		echo "Missing .env — run 'make env' first." >&2; exit 1; \
	fi

infra: infra-up

infra-up: require-env
	$(COMPOSE) up -d $(DC_LONG_RUNNING)

infra-down:
	$(COMPOSE) down

infra-restart:
	$(COMPOSE) restart $(DC_LONG_RUNNING)

infra-reset:
	@read -r -p "This will destroy all local dev data. Type 'yes' to continue: " ans; \
	if [ "$$ans" = "yes" ]; then $(COMPOSE) down -v; else echo "aborted"; fi

infra-ps:
	$(COMPOSE) ps

infra-logs:
	$(COMPOSE) logs -f --tail=100 $(DC_LONG_RUNNING)

psql: require-env
	@docker run --rm -it --network host -e PGPASSWORD=$(OPSYBOT_POSTGRES_PASSWORD) postgres:18-alpine \
		psql -h $(OPSYBOT_POSTGRES_HOST) -p $(OPSYBOT_POSTGRES_PORT) \
		     -U $(OPSYBOT_POSTGRES_USER) -d $(OPSYBOT_POSTGRES_DATABASE)

migration:
	@test -n "$(name)" || (echo "usage: make migration name=create_users"; exit 1)
	go tool goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

gen:
	go generate ./...

wire:
	go tool wire gen ./internal

db-gen: require-env
	go tool sqlboiler psql -c sqlboiler.toml
