MIGRATIONS_DIR := db/migrations/postgres

.PHONY: migration
migration:
	@test -n "$(name)" || (echo "usage: make migration name=create_users"; exit 1)
	go tool goose -dir $(MIGRATIONS_DIR) -s create $(name) sql
