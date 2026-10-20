# DocTheraLink — common tasks. Run `make help` for the list.

.DEFAULT_GOAL := help
COMPOSE := docker compose

.PHONY: help up down logs build test test-be test-be-integration test-fe check k8s-render

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

up: .env ## Build and start the stack (app on http://localhost:4000)
	$(COMPOSE) up -d --build --wait

down: ## Stop the stack (keeps data volumes)
	$(COMPOSE) down

logs: ## Follow logs of all services
	$(COMPOSE) logs -f

build: ## Build Docker images
	$(COMPOSE) build backend frontend

.env:
	@echo "No .env found: copy .env.example to .env and set the secrets first." && exit 1

test: test-be test-fe ## Run backend and frontend unit tests

test-be: ## Backend unit tests
	cd be && go test -race -count=1 ./...

test-be-integration: ## Backend integration tests (needs Docker)
	cd be && go test -tags integration -count=1 ./internal/integration/...

test-fe: ## Frontend unit tests
	cd fe && npm test

check: ## Everything CI runs (except image builds)
	cd be && test -z "$$(gofmt -l .)" && go vet ./... && go vet -tags integration ./...
	$(MAKE) test test-be-integration
	cd fe && npm run typecheck && npx eslint --max-warnings=0 src/components/Referral src/test

k8s-render: ## Render the Kubernetes manifests (kustomize)
	kubectl kustomize k8s
