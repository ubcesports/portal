ifneq (,$(wildcard backend/.env))
    include backend/.env
    export
endif

.PHONY: be fe dev stripe-webhook build-be build-fe sqlc migration-new migration-up migration-down seed DB_CHECK

POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5433
POSTGRES_ENV = PGHOST="$(POSTGRES_HOST)" PGPORT="$(POSTGRES_PORT)" PGUSER="$(POSTGRES_USER)" PGPASSWORD="$(POSTGRES_PASSWORD)" PGDATABASE="$(POSTGRES_DB)" PGSSLMODE=disable
DOCKER_COMPOSE = docker compose --env-file backend/.env -f deploy/compose.local.yaml

# nextjs commands

fe:
	cd frontend && npm run dev

build-fe:
	cd frontend && npm run build

# go commands

be:
	cd backend && go run cmd/api/main.go

build-be:
	cd backend && go build -o bin/api cmd/api/main.go

# stripe related

stripe-webhook:
	stripe listen --forward-to localhost:8080/webhooks/stripe

# db

db:
	docker compose -f compose.yaml up -d

# run both backend + frontend locally (requires you to install concurrently)

dev:
	npx concurrently \
		"make be" \
		"make fe" \
		"make stripe-webhook"

# run both backend + frontend using docker

docker:
	docker compose -f deploy/compose.local.yaml up --build

# sqlc commands

sqlc:
	cd backend && sqlc generate

# database commands

DB_CHECK:
ifndef POSTGRES_USER
	$(error Error: POSTGRES_USER is not set. Make sure backend/.env contains the PostgreSQL settings)
endif
ifndef POSTGRES_PASSWORD
	$(error Error: POSTGRES_PASSWORD is not set. Make sure backend/.env contains the PostgreSQL settings)
endif
ifndef POSTGRES_DB
	$(error Error: POSTGRES_DB is not set. Make sure backend/.env contains the PostgreSQL settings)
endif

# (usage: make migration-new name=add_billing)
migration-new:
ifndef name
	$(error Error: Please provide a migration name. Example: make migration-new name=add_billing_table)
endif
	cd backend && goose -dir sql/migrations create $(name) sql

migration-up: DB_CHECK
	@cd backend && $(POSTGRES_ENV) goose -dir sql/migrations postgres "" up

migration-down: DB_CHECK
	@cd backend && $(POSTGRES_ENV) goose -dir sql/migrations postgres "" down

# (usage: make seed file=mock_admin_audit_logs.sql)
seed: DB_CHECK
ifndef file
	$(error Error: Please provide a seed filename. Example: make seed file=mock_admin_audit_logs.sql)
endif
	@test "$(notdir $(file))" = "$(file)" || \
		(echo "Error: file must be a filename from backend/sql/seeds"; exit 1)
	@test -f "backend/sql/seeds/$(file)" || \
		(echo "Error: seed file not found: backend/sql/seeds/$(file)"; exit 1)
	@$(POSTGRES_ENV) psql -v ON_ERROR_STOP=1 -f "backend/sql/seeds/$(file)"
