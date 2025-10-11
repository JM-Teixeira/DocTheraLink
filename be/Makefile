# CARE Backend Makefile

# Variables
BINARY_NAME=care-backend
MAIN_PATH=cmd/server/main.go
MIGRATE_PATH=cmd/migrate/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_NAME)-linux $(MAIN_PATH)

# Build migration tool
build-migrate:
	$(GOBUILD) -o bin/migrate $(MIGRATE_PATH)

# Run the application
run:
	$(GOCMD) run $(MAIN_PATH)

# Run with live reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf bin/

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -cover ./...

# Run tests with coverage report
test-coverage-html:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run database migrations
migrate:
	$(GOCMD) run $(MIGRATE_PATH)

# Run database migrations with seeding
migrate-seed:
	$(GOCMD) run $(MIGRATE_PATH) -seed

# Generate Swagger documentation (requires swag: go install github.com/swaggo/swag/cmd/swag@latest)
swagger:
	swag init -g cmd/server/main.go -o docs

# Docker commands
docker-build:
	docker build -t $(BINARY_NAME) .

docker-run:
	docker run -p 9000:9000 $(BINARY_NAME)

# Docker Compose commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f api

docker-rebuild:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Database commands (using Docker Compose)
db-up:
	docker-compose up -d postgres redis

db-down:
	docker-compose down postgres redis

db-reset:
	docker-compose down postgres
	docker volume rm care-backend_postgres_data
	docker-compose up -d postgres

# Create a new migration file
create-migration:
	@read -p "Enter migration name: " name; \
	mkdir -p migrations; \
	touch migrations/$$(date +%Y%m%d%H%M%S)_$$name.sql

# Install development tools
install-tools:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/swaggo/swag/cmd/swag@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup development environment
setup-dev: install-tools deps
	cp .env.example .env
	@echo "Development environment setup complete!"
	@echo "1. Update .env file with your configuration"
	@echo "2. Run 'make docker-up' to start services"
	@echo "3. Run 'make migrate-seed' to setup database"
	@echo "4. Run 'make dev' to start development server"

# Production build
build-prod:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-w -s" -o bin/$(BINARY_NAME) $(MAIN_PATH)

# Help
help:
	@echo "Available commands:"
	@echo "  build          - Build the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  build-migrate  - Build migration tool"
	@echo "  run            - Run the application"
	@echo "  dev            - Run with live reload"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Download dependencies"
	@echo "  deps-update    - Update dependencies"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  migrate        - Run database migrations"
	@echo "  migrate-seed   - Run migrations with seeding"
	@echo "  swagger        - Generate Swagger docs"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  docker-up      - Start all services with Docker Compose"
	@echo "  docker-down    - Stop all services"
	@echo "  docker-logs    - View API logs"
	@echo "  docker-rebuild - Rebuild and restart services"
	@echo "  db-up          - Start database services"
	@echo "  db-down        - Stop database services"
	@echo "  db-reset       - Reset database"
	@echo "  setup-dev      - Setup development environment"
	@echo "  build-prod     - Production build"
	@echo "  help           - Show this help"

.PHONY: build build-linux build-migrate run dev clean test test-coverage test-coverage-html deps deps-update lint fmt migrate migrate-seed swagger docker-build docker-run docker-up docker-down docker-logs docker-rebuild db-up db-down db-reset create-migration install-tools setup-dev build-prod help