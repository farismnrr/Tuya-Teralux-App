# Teralux App Backend - Makefile for Development Automation

.PHONY: help dev start install-watch build build-docker start-docker push pull-docker start-compose stop-compose update test clean kill migrate-up migrate-down migrate-version

# Default target
help:
	@echo "Teralux App Backend - Available Commands:"
	@echo ""
	@echo "  make dev              - Run development server with hot reload (Air)"
	@echo "  make start            - Run development server without hot reload"
	@echo "  make install-watch    - Install Air for hot reload"
	@echo "  make build            - Build release binary"
	@echo "  make build-docker     - Build Docker image"
	@echo "  make start-docker     - Run Docker image (with .env)"
	@echo "  make push             - Push Docker image to GHCR"
	@echo "  make pull-docker      - Pull latest image for Docker Compose"
	@echo "  make start-compose    - Start Docker Compose stack (pulls first)"
	@echo "  make stop-compose     - Stop Docker Compose stack"
	@echo "  make update           - Update running container using Watchtower"
	@echo "  make test             - Run all unit tests"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make kill             - Kill process running on PORT 8080"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-version  - Show current migration version"
	@echo ""

# Run development server with hot reload (requires air)
dev:
	@echo "🚀 Starting development server with hot reload..."
	@echo "💡 Tip: Install air with 'make install-watch' if not installed"
	@$(shell go env GOPATH)/bin/air || (echo "❌ air not found. Installing..." && go install github.com/air-verse/air@latest && $(shell go env GOPATH)/bin/air)

# Run development server without hot reload
start:
	@echo "🚀 Starting development server (no hot reload)..."
	go run main.go

# Install air for hot reload (Go equivalent of cargo-watch)
install-watch:
	@echo "📦 Installing Air..."
	go install github.com/air-verse/air@latest
	@echo "✅ Air installed successfully"

# Build release binary
build:
	@echo "🔨 Building release binary..."
	go build -o main .

# --- Docker Configuration ---
DOCKER_IMAGE_NAME = teralux-backend
GHCR_REPO = ghcr.io/pt-perkasa-pilar-utama/teralux-backend
# GHCR_REPO = ghcr.io/farismnrr/teralux-backend
DOCKER_TAG = latest

# Build via Docker
build-docker:
	@echo "🐳 Building Docker images..."
	@echo "  Building latest (no auto-migrate)..."
	docker build -t $(GHCR_REPO):latest --build-arg AUTO_MIGRATE=false .
	@echo "  Building dev (auto-migrate)..."
	docker build -t $(GHCR_REPO):dev --build-arg AUTO_MIGRATE=true .

# Start dev container locally
start-docker-dev:
	@echo "🚀 Starting Dev Docker container..."
	docker run --rm -it -p 8080:8080 --env-file .env $(DOCKER_IMAGE_NAME):dev

# Push to GHCR (reads env vars)
push:
	@echo "🔐 Logging in to GHCR..."
	@export $$(grep -v '^#' .env | grep -v '^$$' | sed 's/#.*//g' | xargs) && \
	(echo "$${CR_PAT:-$$GITHUB_TOKEN}" | docker login ghcr.io -u farismnrr --password-stdin)
	@read -p "Enter tag to push (default: latest): " TAG; \
	TAG=$${TAG:-latest}; \
	echo "🚀 Setting up Docker Buildx..."; \
	docker buildx inspect multi-arch-builder > /dev/null 2>&1 || docker buildx create --name multi-arch-builder --use; \
	echo "🚀 Building and Pushing '$$TAG' (Multi-arch: amd64, arm64)..."; \
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg AUTO_MIGRATE=false -t $(GHCR_REPO):$$TAG --push .; \
	echo "🚀 Building and Pushing 'dev' (Multi-arch: amd64, arm64)..."; \
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg AUTO_MIGRATE=true -t $(GHCR_REPO):dev --push .; \
	echo "✅ Images pushed to GHCR (Tags: $$TAG, dev)"

# Pull latest image for Docker Compose
pull-docker:
	@echo "📥 Pulling Docker images..."
	docker compose pull
	docker compose -f docker-compose.dev.yml pull

# Start Docker Compose (Production/Latest)
start-compose:
	@echo "🚀 Starting Docker Compose (Latest)..."
	docker compose -f docker-compose.yml up -d

# Start Docker Compose (Dev)
start-compose-dev:
	@echo "🚀 Starting Docker Compose (Dev)..."
	docker compose -f docker-compose.dev.yml up -d

# Stop Docker Compose
stop-compose:
	@echo "🛑 Stopping Docker Compose usage..."
	docker compose -f docker-compose.yml down
	docker compose -f docker-compose.dev.yml down

# Update running container using Watchtower
update:
	@echo "🔄 Checking for updates with Watchtower..."
	@export $$(grep -v '^#' .env | grep -v '^$$' | xargs) && \
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		--env DOCKER_API_VERSION=1.45 \
		--env REPO_USER=farismnrr \
		--env REPO_PASS="$${CR_PAT:-$$GITHUB_TOKEN}" \
		containrrr/watchtower \
		--debug \
		--run-once \
		$(DOCKER_IMAGE_NAME)

# Run all tests
test:
	@echo "🧪 Running all tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	go clean
	rm -f main
	@echo "✅ Clean completed"

# Kill process running on port 8080 (server's default port)
kill:
	@echo "🔪 Killing processes on port 8080..."
	@lsof -ti:8080 | xargs -r kill -9 || echo "✅ No process running on port 8080"

# --- Database Migration Configuration ---
MIGRATE_VERSION = v4.18.1
MIGRATE_URL = https://github.com/golang-migrate/migrate/releases/download/$(MIGRATE_VERSION)/migrate.linux-amd64.tar.gz
MIGRATE_BIN = $(shell go env GOPATH)/bin/migrate

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string from .env
DB_URL = mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?charset=utf8mb4&parseTime=True&loc=Local

# Install golang-migrate if not found
$(MIGRATE_BIN):
	@echo "📦 Installing golang-migrate..."
	@mkdir -p $(shell go env GOPATH)/bin
	@curl -L $(MIGRATE_URL) | tar xvz -C $(shell go env GOPATH)/bin
	@chmod +x $(MIGRATE_BIN)
	@echo "✅ golang-migrate installed successfully"

# Run all pending migrations
migrate-up: $(MIGRATE_BIN)
	@echo "⬆️  Running migrations..."
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" up
	@echo "✅ Migrations completed"

# Rollback last migration
migrate-down: $(MIGRATE_BIN)
	@echo "⬇️  Rolling back last migration..."
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" down 1
	@echo "✅ Rollback completed"

# Show current migration version
migrate-version: $(MIGRATE_BIN)
	@echo "📊 Current migration version:"
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" version
