# Klyniva — common tasks. Run `make help` for the list.

.DEFAULT_GOAL := help
COMPOSE := docker compose

.PHONY: help up down logs build test test-be test-be-integration test-fe check k8s-render \
	k8s-up k8s-start k8s-deploy k8s-e2e k8s-status k8s-down e2e-compose e2e-browser

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
	cd be && test -z "$$(gofmt -l .)" && go vet ./... && go vet -tags integration ./... && go vet -tags e2e ./...
	$(MAKE) test test-be-integration
	cd fe && npm run typecheck && npm run lint:strict

k8s-up: ## Local Kubernetes (k3d): create cluster, build, deploy → http://localhost:8088
	bash scripts/k8s-local.sh up

k8s-start: ## Start the existing local cluster after a reboot and print the logins
	bash scripts/k8s-local.sh start

k8s-deploy: ## Rebuild images and roll out to the local cluster
	bash scripts/k8s-local.sh deploy

k8s-e2e: ## End-to-end API tests against the local cluster
	bash scripts/k8s-local.sh e2e

k8s-status: ## Pods/services of the local cluster
	bash scripts/k8s-local.sh status

k8s-down: ## Delete the local cluster (and its data)
	bash scripts/k8s-local.sh down

e2e-compose: ## End-to-end API tests against the docker compose stack
	cd be && E2E_BASE_URL=http://localhost:4000 go test -tags e2e -count=1 -v ./internal/e2e/

e2e-browser: ## Browser end-to-end tests (Playwright) against the local cluster (E2E_WEB_URL to override)
	cd e2e/browser && npm ci --silent && npx playwright install chromium && E2E_WEB_URL=${E2E_WEB_URL:-http://localhost:8088} npx playwright test

k8s-render: ## Render the production overlay (kustomize)
	kubectl kustomize k8s/overlays/production
