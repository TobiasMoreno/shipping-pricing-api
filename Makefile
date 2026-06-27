.PHONY: run build up down logs test test-integration lint tidy migrate-up migrate-down seed swagger

# Run the API locally (loads env from your shell or .env exported beforehand).
run:
	go run ./cmd/api

# Build the API binary.
build:
	go build -trimpath -o bin/api ./cmd/api

# Bring the local stack up (API + PostgreSQL + Redis).
up:
	docker compose up -d --build

# Tear the local stack down.
down:
	docker compose down

# Follow API logs.
logs:
	docker compose logs -f api

# Unit tests.
test:
	go test ./...

# Integration tests (added in a later change; tagged build).
test-integration:
	go test -tags=integration ./tests/integration/...

# Static checks.
lint:
	go vet ./...

tidy:
	go mod tidy

# --- Placeholders wired in later changes ---

migrate-up:
	@echo "migrate-up: implemented in add-postgres-rules change"

migrate-down:
	@echo "migrate-down: implemented in add-postgres-rules change"

seed:
	@echo "seed: implemented in add-postgres-rules change"

swagger:
	@echo "swagger: implemented in add-ci-openapi change"
