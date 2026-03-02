GO := /home/emir/sdk/go1.26.0/bin/go
BINARY := mahzen
CMD := ./cmd/mahzen

.PHONY: all build run clean test lint proto-gen sqlc-gen migrate-up migrate-down deps docker-up docker-down

all: build

## build: Build the application binary
build:
	$(GO) build -o $(BINARY) $(CMD)

## run: Run the application
run:
	$(GO) run $(CMD) -config config.yaml

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	$(GO) clean

## test: Run all tests
test:
	$(GO) test ./... -v -race

## lint: Run linter
lint:
	golangci-lint run ./...

## proto-gen: Generate protobuf and gRPC gateway stubs
proto-gen:
	buf dep update
	buf generate

## sqlc-gen: Generate sqlc database code
sqlc-gen:
	sqlc generate -f sqlc/sqlc.yaml

## migrate-up: Run database migrations up
migrate-up:
	migrate -path migrations -database "postgres://mahzen:mahzen@localhost:5432/mahzen?sslmode=disable" up

## migrate-down: Run database migrations down
migrate-down:
	migrate -path migrations -database "postgres://mahzen:mahzen@localhost:5432/mahzen?sslmode=disable" down

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
