# ASAM Backend Makefile
# This Makefile provides convenient commands for development

# Variables
DOCKER_COMPOSE = docker-compose
GO = go
GOTEST = $(GO) test
GOVET = $(GO) vet
GOFMT = gofmt
GOLINT = golangci-lint
BINARY_NAME = asam-backend
MAIN_PATH = ./cmd/api

# Color codes for pretty output
GREEN = \033[0;32m
YELLOW = \033[0;33m
RED = \033[0;31m
NC = \033[0m # No Color

.DEFAULT_GOAL := help

## help: Show this help message
.PHONY: help
help:
	@echo 'Usage:'
	@echo '  ${GREEN}make${NC} ${YELLOW}<target>${NC}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${GREEN}%-20s${NC} %s\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${YELLOW}%s${NC}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Development commands
## ─────────────────────────────────────────────────────────────────

## dev: Start development environment with Docker
.PHONY: dev
dev:
	@echo "🚀 Starting development environment..."
	@$(DOCKER_COMPOSE) up -d
	@echo "✅ Development environment started"
	@echo "📝 GraphQL Playground: http://localhost:8080/playground"

## dev-setup: Complete setup (start, migrate, seed)
.PHONY: dev-setup
dev-setup: dev
	@echo "⏳ Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) db-migrate
	@$(MAKE) db-seed
	@echo "\n✅ Development environment ready!"
	@echo "🌐 GraphQL Playground: http://localhost:8080/playground"
	@echo "👤 Test users: admin@asam.org / admin123, user@asam.org / admin123"

## dev-logs: Show development logs
.PHONY: dev-logs
dev-logs:
	@$(DOCKER_COMPOSE) logs -f api

## dev-stop: Stop development environment
.PHONY: dev-stop
dev-stop:
	@echo "🛑 Stopping development environment..."
	@$(DOCKER_COMPOSE) down
	@echo "✅ Development environment stopped"

## dev-restart: Restart development environment
.PHONY: dev-restart
dev-restart: dev-stop dev

## clean: Clean everything (containers, volumes, build artifacts)
.PHONY: clean
clean:
	@echo "🧹 Cleaning everything..."
	@$(DOCKER_COMPOSE) down -v
	@rm -rf tmp/
	@rm -rf logs/
	@rm -rf coverage*
	@rm -rf bin/
	@echo "✅ Cleanup complete"

## Database commands
## ─────────────────────────────────────────────────────────────────

## db-migrate: Run database migrations
.PHONY: db-migrate
db-migrate:
	@echo "🔄 Running database migrations..."
	@$(DOCKER_COMPOSE) exec api go run ./cmd/migrate up
	@echo "✅ Migrations complete"

## db-rollback: Rollback last migration
.PHONY: db-rollback
db-rollback:
	@echo "⏪ Rolling back last migration..."
	@$(DOCKER_COMPOSE) exec api go run ./cmd/migrate down 1
	@echo "✅ Rollback complete"

## db-reset: Reset database (drop all tables and re-migrate)
.PHONY: db-reset
db-reset:
	@echo "⚠️  Resetting database..."
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d asam_db -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	@$(MAKE) db-migrate
	@$(MAKE) db-seed
	@echo "✅ Database reset complete"

## db-seed: Seed database with test data
.PHONY: db-seed
db-seed:
	@echo "🌱 Seeding database..."
	@$(DOCKER_COMPOSE) exec api go run scripts/user-management/auto-create-test-users.go
	@echo "✅ Database seeded"

## db-shell: Open PostgreSQL shell
.PHONY: db-shell
db-shell:
	@$(DOCKER_COMPOSE) exec postgres psql -U postgres -d asam_db

## Code generation commands
## ─────────────────────────────────────────────────────────────────

## generate: Generate GraphQL code
.PHONY: generate
generate:
	@echo "🔧 Generating GraphQL code..."
	@$(GO) run ./cmd/generate
	@echo "✅ Code generation complete"

## generate-mocks: Generate mocks for testing
.PHONY: generate-mocks
generate-mocks:
	@echo "🔧 Generating mocks..."
	@$(GO) generate ./...
	@echo "✅ Mock generation complete"

## Testing commands
## ─────────────────────────────────────────────────────────────────

## test: Run all tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@$(GOTEST) -v -race ./...

## test-unit: Run unit tests only
.PHONY: test-unit
test-unit:
	@echo "🧪 Running unit tests..."
	@$(GOTEST) -v -race -short ./...

## test-integration: Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests..."
	@$(GOTEST) -v -race -tags=integration ./...

## test-coverage: Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## test-coverage-view: View coverage report in browser
.PHONY: test-coverage-view
test-coverage-view: test-coverage
	@echo "🌐 Opening coverage report..."
	@start coverage.html 2>/dev/null || open coverage.html 2>/dev/null || xdg-open coverage.html

## Code quality commands
## ─────────────────────────────────────────────────────────────────

## lint: Run linter
.PHONY: lint
lint:
	@echo "🔍 Running linter..."
	@$(GOLINT) run

## fmt: Format code
.PHONY: fmt
fmt:
	@echo "✨ Formatting code..."
	@$(GOFMT) -s -w .
	@$(GO) mod tidy

## vet: Run go vet
.PHONY: vet
vet:
	@echo "🔍 Running go vet..."
	@$(GOVET) ./...

## security: Run security scan
.PHONY: security
security:
	@echo "🔒 Running security scan..."
	@gosec -quiet ./...

## Build commands
## ─────────────────────────────────────────────────────────────────

## build: Build the application
.PHONY: build
build:
	@echo "🔨 Building application..."
	@mkdir -p bin
	@$(GO) build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

## build-linux: Build for Linux
.PHONY: build-linux
build-linux:
	@echo "🔨 Building for Linux..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Linux build complete"

## build-windows: Build for Windows
.PHONY: build-windows
build-windows:
	@echo "🔨 Building for Windows..."
	@mkdir -p bin
	@GOOS=windows GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Windows build complete"

## build-mac: Build for macOS
.PHONY: build-mac
build-mac:
	@echo "🔨 Building for macOS..."
	@mkdir -p bin
	@GOOS=darwin GOARCH=amd64 $(GO) build -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 $(GO) build -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "✅ macOS build complete"

## build-all: Build for all platforms
.PHONY: build-all
build-all: build-linux build-windows build-mac

## Docker commands
## ─────────────────────────────────────────────────────────────────

## docker-build: Build Docker image
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	@docker build -t asam-backend:latest .
	@echo "✅ Docker image built"

## docker-run: Run Docker container
.PHONY: docker-run
docker-run:
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 --env-file .env asam-backend:latest

## Utility commands
## ─────────────────────────────────────────────────────────────────

## deps: Download dependencies
.PHONY: deps
deps:
	@echo "📦 Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Dependencies downloaded"

## update-deps: Update dependencies
.PHONY: update-deps
update-deps:
	@echo "📦 Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "✅ Dependencies updated"

## tools: Install development tools
.PHONY: tools
tools:
	@echo "🔧 Installing development tools..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	@$(GO) install github.com/air-verse/air@latest
	@$(GO) install github.com/99designs/gqlgen@v0.17.73
	@echo "✅ Tools installed"

## version: Show version information
.PHONY: version
version:
	@echo "ASAM Backend"
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $(shell date +%Y-%m-%d)"

# Quick aliases
.PHONY: d dev-logs l c t
d: dev
l: dev-logs
c: clean
t: test
