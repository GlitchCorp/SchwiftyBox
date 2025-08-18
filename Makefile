# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=schwiftybox
BINARY_UNIX=$(BINARY_NAME)_unix
SRC_DIR=./src

# Database parameters
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=user
DB_PASSWORD?=password
DB_NAME?=mydb
DB_URL?=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Migration tool
MIGRATE_VERSION=v4.18.3
MIGRATE_TOOL=migrate

.PHONY: all build clean test deps help migrate-up migrate-down migrate-version migrate-create migrate-install

# Default target
all: test build

# Build the binary
build:
	cd $(SRC_DIR) && $(GOBUILD) -o ../$(BINARY_NAME) -v ./...

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run tests
test:
	cd $(SRC_DIR) && $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	cd $(SRC_DIR) && $(GOTEST) -v -coverprofile=coverage.out ./...
	cd $(SRC_DIR) && $(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: src/coverage.html"
	@cd $(SRC_DIR) && $(GOCMD) tool cover -func=coverage.out | tail -1

# Show test coverage percentage
test-coverage-percent:
	cd $(SRC_DIR) && $(GOTEST) -coverprofile=coverage.out ./... > /dev/null 2>&1
	@cd $(SRC_DIR) && $(GOCMD) tool cover -func=coverage.out | tail -1

# Run tests with race detection
test-race:
	cd $(SRC_DIR) && $(GOTEST) -v -race ./...

# Run tests for specific package
test-package:
	cd $(SRC_DIR) && $(GOTEST) -v ./internal/$(PACKAGE)

# Run integration tests
test-integration:
	chmod +x ./scripts/run_integration_tests.sh
	./scripts/run_integration_tests.sh

# Run end-to-end tests
test-e2e:
	cd $(SRC_DIR) && $(GOTEST) -v -run TestE2ETestSuite ./e2e_test.go

# Download dependencies
deps:
	cd $(SRC_DIR) && $(GOMOD) download
	cd $(SRC_DIR) && $(GOMOD) tidy

# Run the application
run:
	cd $(SRC_DIR) && $(GOCMD) run ./...

# Build for Linux
build-linux:
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o ../$(BINARY_UNIX) -v ./...

# Install migration tool
migrate-install:
	@which $(MIGRATE_TOOL) > /dev/null || (echo "Installing golang-migrate..." && \
		curl -L https://github.com/golang-migrate/migrate/releases/download/$(MIGRATE_VERSION)/migrate.linux-amd64.tar.gz | tar xvz && \
		sudo mv migrate /usr/local/bin/$(MIGRATE_TOOL))

# Run pending migrations
migrate-up: migrate-install
	$(MIGRATE_TOOL) -path ./migrations -database "$(DB_URL)" up

# Rollback last migration
migrate-down: migrate-install
	$(MIGRATE_TOOL) -path ./migrations -database "$(DB_URL)" down 1

# Show current migration version
migrate-version: migrate-install
	$(MIGRATE_TOOL) -path ./migrations -database "$(DB_URL)" version

# Create a new migration file
# Usage: make migrate-create NAME=create_posts_table
migrate-create: migrate-install
	@if [ "$(NAME)" = "" ]; then \
		echo "Error: NAME parameter is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	$(MIGRATE_TOOL) create -ext sql -dir ./migrations -seq $(NAME)

# Force migration to specific version (dangerous!)
# Usage: make migrate-force VERSION=1
migrate-force: migrate-install
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION parameter is required. Usage: make migrate-force VERSION=1"; \
		exit 1; \
	fi
	$(MIGRATE_TOOL) -path ./migrations -database "$(DB_URL)" force $(VERSION)

# Development setup
dev-setup: deps migrate-up
	@echo "Development environment is ready!"

# Docker Compose commands
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-build:
	DOCKER_BUILDKIT=1 docker compose build

docker-build-no-cache:
	DOCKER_BUILDKIT=1 docker compose build --no-cache

docker-logs:
	docker compose logs -f

docker-restart:
	docker compose restart

# Development with Docker
docker-dev:
	DOCKER_BUILDKIT=1 docker compose -f docker-compose.yml -f docker-compose.override.yml up

docker-dev-down:
	docker compose -f docker-compose.yml -f docker-compose.override.yml down

docker-dev-logs:
	docker compose -f docker-compose.yml -f docker-compose.override.yml logs -f app

# Database commands with Docker
docker-migrate-up:
	docker-compose exec app /app/scripts/migrate.sh up

docker-migrate-down:
	docker-compose exec app /app/scripts/migrate.sh down

docker-migrate-version:
	docker-compose exec app /app/scripts/migrate.sh version

# Show help
help:
	@echo "Available commands:"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build files"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-coverage-percent - Show test coverage percentage"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-package   - Run tests for specific package (requires PACKAGE parameter)"
	@echo "  test-integration - Run integration tests with real database"
	@echo "  test-e2e        - Run end-to-end tests with real HTTP server"
	@echo "  deps           - Download dependencies"
	@echo "  run            - Run the application"
	@echo "  build-linux    - Build for Linux"
	@echo "  migrate-up     - Run pending migrations"
	@echo "  migrate-down   - Rollback last migration"
	@echo "  migrate-version - Show current migration version"
	@echo "  migrate-create - Create new migration (requires NAME parameter)"
	@echo "  migrate-force  - Force migration to specific version (requires VERSION parameter)"
	@echo "  dev-setup      - Setup development environment"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-up      - Start all services"
	@echo "  docker-down    - Stop all services"
	@echo "  docker-build   - Build Docker images"
	@echo "  docker-logs    - Show logs"
	@echo "  docker-restart - Restart services"
	@echo "  docker-dev     - Start development environment"
	@echo "  docker-dev-down - Stop development environment"
	@echo "  docker-migrate-up - Run migrations in Docker"
	@echo "  docker-migrate-down - Rollback migrations in Docker"
	@echo "  docker-migrate-version - Show migration version in Docker"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  DB_HOST        - Database host (default: localhost)"
	@echo "  DB_PORT        - Database port (default: 5432)"
	@echo "  DB_USER        - Database user (default: user)"
	@echo "  DB_PASSWORD    - Database password (default: password)"
	@echo "  DB_NAME        - Database name (default: mydb)"
