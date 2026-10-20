# Klyniva backend — run `make help`. The root Makefile drives Docker/k8s.

.DEFAULT_GOAL := help
export GOTOOLCHAIN := auto

.PHONY: help run build test test-integration cover fmt vet migrate seed swagger tools

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

run: ## Run the API locally (needs PostgreSQL, see .env.example)
	go run ./cmd/server

build: ## Build server and migrate binaries into bin/
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/klyniva-server ./cmd/server
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/klyniva-migrate ./cmd/migrate

test: ## Unit tests with the race detector
	go test -race -count=1 ./...

test-integration: ## PostgreSQL integration tests (Docker required)
	go test -tags integration -count=1 ./internal/integration/...

cover: ## Unit test coverage report (coverage.html)
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

fmt: ## Format code
	gofmt -w .

vet: ## Static checks
	go vet ./... && go vet -tags integration ./...

migrate: ## Run migrations against the configured database
	go run ./cmd/migrate

seed: ## Run migrations and seed reference/demo data
	go run ./cmd/migrate -seed

swagger: ## Regenerate Swagger docs (needs: go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/server/main.go -o docs

tools: ## Install developer tools
	go install github.com/swaggo/swag/cmd/swag@latest
