GO := /home/emir/sdk/go1.26.0/bin/go
BINARY := mahzen
CMD := ./cmd/mahzen

.PHONY: all build run clean test lint sqlc-gen migrate-up migrate-down deps docker-up docker-down web-install web-dev web-build web-lint dist gen-certs

all: build

## build: Build the application binary
build:
	$(GO) build -o $(BINARY) $(CMD)

## run: Run the application (generates TLS certs if missing)
run:
	@[ -f certs/server.crt ] && [ -f certs/server.key ] || ./scripts/gen-certs.sh
	$(GO) run $(CMD) -config config.yaml

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf cmd/mahzen/dist
	$(GO) clean

## test: Run all tests
test:
	$(GO) test ./... -v -race

## lint: Run linter
lint:
	golangci-lint run ./...

## sqlc-gen: Generate sqlc database code
sqlc-gen:
	sqlc generate -f sqlc/sqlc.yaml

## migrate-up: Apply all migrations
migrate-up:
	docker exec -i mahzen-postgres psql -U mahzen -d mahzen < migrations/000001_init.up.sql

## migrate-down: Roll back all migrations
migrate-down:
	docker exec -i mahzen-postgres psql -U mahzen -d mahzen < migrations/000001_init.down.sql

## deps: Download Go module dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## docker-up: Start all infrastructure services
docker-up:
	docker compose -f deploy/docker-compose.yml up -d

## docker-down: Stop all infrastructure services
docker-down:
	docker compose -f deploy/docker-compose.yml down

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## web-install: Install frontend dependencies
web-install:
	cd web && bun install

## web-dev: Start frontend dev server (installs dependencies if missing)
web-dev:
	@[ -d web/node_modules ] || (cd web && bun install)
	cd web && bun run dev

## web-build: Build frontend for production
web-build:
	cd web && bun run build

## web-lint: Lint frontend code
web-lint:
	cd web && bun run lint

## gen-certs: Generate self-signed TLS certificate for HTTP/3 development
gen-certs:
	./scripts/gen-certs.sh

## dist: Build frontend + embed + Go binary (production)
dist: web-build
	rm -rf cmd/mahzen/dist
	cp -r web/dist cmd/mahzen/dist
	$(GO) build -o $(BINARY) $(CMD)
