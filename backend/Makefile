# Teralux App Backend - Makefile for Development Automation

.PHONY: help dev start install-watch build build-docker start-docker push pull-docker start-compose stop-compose update test clean kill

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
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME) .

# Run via Docker (with .env)
start-docker:
	@echo "🚀 Starting Docker container..."
	docker run --rm -it -p 8080:8080 --env-file .env $(DOCKER_IMAGE_NAME)

# Push to GHCR (reads env vars)
push:
	@echo "🔐 Logging in to GHCR..."
	@export $$(grep -v '^#' .env | grep -v '^$$' | sed 's/#.*//g' | xargs) && \
	(echo "$${CR_PAT:-$$GITHUB_TOKEN}" | docker login ghcr.io -u farismnrr --password-stdin)
	@echo "🚀 Setting up Docker Buildx..."
	# Create a new builder instance if it doesn't exist, or use the existing one
	docker buildx inspect multi-arch-builder > /dev/null 2>&1 || docker buildx create --name multi-arch-builder --use
	@echo "🚀 Building and Pushing to GHCR (Multi-arch: amd64, arm64)..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(GHCR_REPO):$(DOCKER_TAG) --push .
	@echo "✅ Image pushed to $(GHCR_REPO):$(DOCKER_TAG)"

# Pull latest image for Docker Compose
pull-docker:
	@echo "📥 Pulling latest Docker image..."
	docker compose pull

# Start Docker Compose (pulls first)
start-compose: pull-docker
	@echo "🚀 Starting Docker Compose stack..."
	docker compose up -d

# Stop Docker Compose
stop-compose:
	@echo "🛑 Stopping Docker Compose stack..."
	docker compose down

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
