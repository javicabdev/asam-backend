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
	@if grep -qi microsoft /proc/version 2>/dev/null && command -v docker 2>&1 | grep -q "wincred" ; then \
		echo "🔧 Detected WSL with Docker credential issues, using alternative method..."; \
		chmod +x start-docker-wsl.sh; \
		./start-docker-wsl.sh; \
	else \
		$(DOCKER_COMPOSE) up -d; \
		echo "✅ Development environment started"; \
		echo "📝 GraphQL Playground: http://localhost:8080/playground"; \
	fi

## dev-wsl: Start development environment using direct Docker commands (for WSL issues)
.PHONY: dev-wsl
dev-wsl:
	@echo "🚀 Starting development environment with WSL workaround..."
	@chmod +x start-docker-wsl.sh
	@./start-docker-wsl.sh

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
	@if grep -qi microsoft /proc/version 2>/dev/null && [ -f stop-docker-wsl.sh ]; then \
		chmod +x stop-docker-wsl.sh; \
		./stop-docker-wsl.sh; \
	else \
		$(DOCKER_COMPOSE) down; \
		echo "✅ Development environment stopped"; \
	fi

## dev-stop-wsl: Stop development environment using direct Docker commands (for WSL issues)
.PHONY: dev-stop-wsl
dev-stop-wsl:
	@echo "🛑 Stopping development environment with WSL workaround..."
	@chmod +x stop-docker-wsl.sh
	@./stop-docker-wsl.sh

## dev-restart: Restart development environment
.PHONY: dev-restart
dev-restart: dev-stop dev

## dev-rebuild: Rebuild and restart backend with latest code changes
.PHONY: dev-rebuild
dev-rebuild:
	@echo "🔨 Rebuilding backend with latest changes..."
	@$(DOCKER_COMPOSE) build api
	@$(DOCKER_COMPOSE) up -d
	@echo "✅ Backend rebuilt and restarted"
	@echo "📝 GraphQL Playground: http://localhost:8080/playground"

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
	@$(GO) run ./cmptemp/generate
	@echo "✅ Code generation complete"

## generate-mocks: Generate mocks for testing
.PHONY: generate-mocks
generate-mocks:
	@echo "🔧 Generating mocks..."
	@$(GO) generate ./...
	@echo "✅ Mock generation complete"

## Testing commands
## ─────────────────────────────────────────────────────────────────

## test: Run only unit tests (use 'make test-integration' for integration tests)
.PHONY: test
test:
	@echo "🧪 Running unit tests..."
	@$(GOTEST) -v -race ./test/unit/...

## test-all: Run all tests (unit + integration, requires postgres)
.PHONY: test-all
test-all: test-unit test-integration

## test-unit: Run unit tests only
.PHONY: test-unit
test-unit:
	@echo "🧪 Running unit tests..."
	@$(GOTEST) -v -race ./test/unit/...

## test-integration: Run integration tests only (requires postgres running on localhost:5432)
.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests..."
	@echo "📝 Note: Integration tests require postgres running. Use 'make dev' or 'docker compose up postgres' first."
	@DB_HOST=$${DB_HOST:-localhost} DB_PORT=$${DB_PORT:-5432} DB_USER=$${DB_USER:-postgres} \
		DB_PASSWORD=$${DB_PASSWORD:-postgres} DB_NAME=$${DB_NAME:-asam_db} DB_SSL_MODE=$${DB_SSL_MODE:-disable} \
		$(GOTEST) -v -p 1 ./test/integration/...

## test-coverage: Run tests with coverage report
.PHONY: test-coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## test-unit-coverage: Run unit tests with coverage report
.PHONY: test-unit-coverage
test-unit-coverage:
	@echo "📊 Running unit tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage-unit.out ./test/unit/...
	@$(GO) tool cover -html=coverage-unit.out -o coverage-unit.html
	@echo "✅ Unit test coverage report: coverage-unit.html"

## test-ci: Run tests like CI does (unit + integration)
.PHONY: test-ci
test-ci: test-unit-coverage test-integration
	@echo "✅ All CI tests complete"

## test-coverage-view: View coverage report in browser
.PHONY: test-coverage-view
test-coverage-view: test-coverage
	@echo "🌐 Opening coverage report..."
	@start coverage.html 2>/dev/null || open coverage.html 2>/dev/null || xdg-open coverage.html

## test-auth: Run authentication system tests
.PHONY: test-auth
test-auth:
	@echo "🔐 Running authentication system tests..."
	@$(GOTEST) -v -race -coverprofile=coverage_auth.out -covermode=atomic -coverpkg=github.com/javicabdev/asam-backend/internal/domain/services ./test/internal/domain/services
	@$(GO) tool cover -func=coverage_auth.out | grep -E "auth_service.go|user_service.go|total:"
	@echo "✅ Auth tests complete"

## test-auth-coverage: Run auth tests with HTML coverage report
.PHONY: test-auth-coverage
test-auth-coverage: test-auth
	@echo "📊 Generating auth coverage report..."
	@$(GO) tool cover -html=coverage_auth.out -o coverage_auth.html
	@echo "✅ Coverage report: coverage_auth.html"

## test-auth-view: Run auth tests and view coverage
.PHONY: test-auth-view
test-auth-view: test-auth-coverage
	@echo "🌐 Opening auth coverage report..."
	@start coverage_auth.html 2>/dev/null || open coverage_auth.html 2>/dev/null || xdg-open coverage_auth.html

## Code quality commands
## ─────────────────────────────────────────────────────────────────

## lint: Run linter (same as CI/CD)
.PHONY: lint
lint:
	@echo "🔍 Running golangci-lint (CI/CD configuration)..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "${RED}❌ golangci-lint not found!${NC}"; \
		echo "${YELLOW}Install it with: make tools${NC}"; \
		exit 1; \
	fi
	@$(GOLINT) run --timeout=5m ./cmd/api/... ./internal/... ./pkg/... ./test/...
	@echo "${GREEN}✅ Linting complete - no issues found!${NC}"

## security: Run security scan with gosec (SAST)
.PHONY: security
security:
	@echo "🔒 Running gosec security scanner..."
	@if ! command -v gosec &> /dev/null; then \
		echo "${RED}❌ gosec not found!${NC}"; \
		echo "${YELLOW}Install it with: make tools${NC}"; \
		exit 1; \
	fi
	@gosec -fmt=json -out=gosec-report.json ./...
	@gosec -fmt=text ./...
	@echo "${GREEN}✅ Security scan complete!${NC}"
	@echo "${YELLOW}📄 Full report saved to: gosec-report.json${NC}"

## security-ci: Run security scan with SARIF output for CI/CD
.PHONY: security-ci
security-ci:
	@echo "🔒 Running gosec for CI/CD..."
	@gosec -no-fail -fmt sarif -out gosec-results.sarif ./...
	@echo "${GREEN}✅ Security scan complete (SARIF format)${NC}"

## lint-fix: Run linter and auto-fix issues
.PHONY: lint-fix
lint-fix:
	@echo "🛠️  Running golangci-lint with auto-fix..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "${RED}❌ golangci-lint not found!${NC}"; \
		echo "${YELLOW}Install it with: make tools${NC}"; \
		exit 1; \
	fi
	@$(GOLINT) run --timeout=5m --fix
	@echo "${GREEN}✅ Auto-fix complete!${NC}"

## lint-tests: Run linter on test files only
.PHONY: lint-tests
lint-tests:
	@echo "🔍 Running linter on tests..."
	@$(GOLINT) run ./test/...

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

## Maintenance commands
## ─────────────────────────────────────────────────────────────────

## maintenance-cleanup: Clean up expired tokens
.PHONY: maintenance-cleanup
maintenance-cleanup:
	@echo "🧹 Cleaning up expired tokens..."
	@$(GO) run ./cmptemp/maintenance -cleanup-tokens -report
	@echo "✅ Token cleanup complete"

## maintenance-limit: Enforce token limit per user
.PHONY: maintenance-limit
maintenance-limit:
	@echo "🔒 Enforcing token limits..."
	@$(GO) run ./cmptemp/maintenance -enforce-token-limit -report
	@echo "✅ Token limit enforcement complete"

## maintenance-all: Run all maintenance tasks
.PHONY: maintenance-all
maintenance-all:
	@echo "🔧 Running all maintenance tasks..."
	@$(GO) run ./cmptemp/maintenance -all -report
	@echo "✅ All maintenance tasks complete"

## maintenance-dry: Dry run of maintenance tasks
.PHONY: maintenance-dry
maintenance-dry:
	@echo "👀 Running maintenance dry run..."
	@$(GO) run ./cmptemp/maintenance -all -dry-run -report

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
	@$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
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
.PHONY: d dev-logs l c t ta
d: dev
l: dev-logs
c: clean
t: test
ta: test-auth

## verify-mailersend: Verify MailerSend migration
.PHONY: verify-mailersend
verify-mailersend:
	@echo "📧 Verifying MailerSend migration..."
	@echo "Checking dependencies..."
	@grep -q "mailersend/mailersend-go" go.mod && echo "✅ MailerSend SDK found" || echo "❌ MailerSend SDK missing"
	@echo "Checking files..."
	@test -f internal/adapters/email/mailersend_adapter.go && echo "✅ Adapter exists" || echo "❌ Adapter missing"
	@echo "Checking .env..."
	@grep -q "MAILERSEND_API_KEY" .env && echo "✅ API Key configured" || echo "❌ API Key missing"
	@echo "Building project..."
	@go build -o /tmp/asam-test ./cmd/api 2>/dev/null && echo "✅ Project builds successfully" || echo "❌ Build failed"
	@rm -f /tmp/asam-test
	@echo "✨ MailerSend migration verification complete"
