.PHONY: migrate-up migrate-down migrate-status migrate-create

GOOSE ?= goose
MIGRATIONS_DIR ?= migrations

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	@test -n "$(name)" || (echo "Usage: make migrate-create name=description"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) create "$(name)" sql
