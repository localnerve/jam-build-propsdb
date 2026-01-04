.PHONY: help build build-all clean test test-unit test-integration test-coverage run docker-build docker-run swagger lint fmt vet

# Variables
BINARY_NAME=propsdb
HEALTHCHECK_BINARY=healthcheck
DOCKER_IMAGE=propsdb
DOCKER_TAG=latest
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html
SWAGGER_DIR=docs/api

# Docker parameters
DOCKER_MARIADB_IMAGE=mariadb:12.1.2
DOCKER_POSTGRES_IMAGE=postgres:18-alpine
DOCKER_AUTHZ_IMAGE=localnerve/authorizer:1.5.3

TEST_ENV=DOCKER_MARIADB_IMAGE=$(DOCKER_MARIADB_IMAGE) DOCKER_POSTGRES_IMAGE=$(DOCKER_POSTGRES_IMAGE) DOCKER_AUTHZ_IMAGE=$(DOCKER_AUTHZ_IMAGE)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(TEST_ENV) $(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_NAME)"

build-healthcheck: ## Build the healthcheck binary
	@echo "Building $(HEALTHCHECK_BINARY)..."
	$(GOBUILD) -o $(HEALTHCHECK_BINARY) ./cmd/healthcheck
	@echo "Build complete: $(HEALTHCHECK_BINARY)"

build-all: build build-healthcheck ## Build all binaries
	@echo "All binaries built successfully"
	@ls -lh $(BINARY_NAME) $(HEALTHCHECK_BINARY)

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(HEALTHCHECK_BINARY)
	rm -rf $(COVERAGE_DIR)
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies updated"

test: test-unit ## Run all tests (alias for test-unit)

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./tests/unit/...

test-integration: ## Run integration tests (requires Docker)
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

test-e2e: ## Run end-to-end tests with full stack (requires Docker)
	@echo "Running E2E tests..."
	$(GOTEST) -v ./tests/e2e/... -timeout 300s

test-e2e-rebuild: ## Run E2E tests with forced rebuild of propsdb-test image
	@echo "Forcing rebuild of propsdb-test images..."
	docker rmi propsdb-test:latest || true
	@echo "Running E2E tests..."
	$(GOTEST) -v ./tests/e2e/... -timeout 300s

test-all: ## Run all tests including integration and E2E
	@echo "Running all tests..."
	$(GOTEST) -v ./tests/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"
	@echo "Opening coverage report in browser..."
	@open $(COVERAGE_HTML) 2>/dev/null || xdg-open $(COVERAGE_HTML) 2>/dev/null || echo "Please open $(COVERAGE_HTML) manually"

coverage-report: ## Generate and display coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

run: build ## Build and run the server
	@echo "Starting $(BINARY_NAME)..."
	./$(BINARY_NAME)

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 3000:3000 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start all services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker-compose up -d
	@echo "Services started. Use 'make docker-compose-logs' to view logs"

docker-compose-down: ## Stop all Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker-compose down

docker-compose-logs: ## View Docker Compose logs
	docker-compose logs -f

swagger: ## Generate OpenAPI/Swagger documentation
	@echo "Generating Swagger documentation..."
	@mkdir -p $(SWAGGER_DIR)
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$$(go env GOPATH)/bin/swag init -g cmd/server/main.go -o $(SWAGGER_DIR)
	@echo "Swagger documentation generated in $(SWAGGER_DIR)"

swagger-serve: swagger ## Generate and serve Swagger UI
	@echo "Swagger documentation available at http://localhost:3000/swagger/index.html"
	@echo "Run 'make run' to start the server"

lint: ## Run linter
	@echo "Running linter..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@$$(go env GOPATH)/bin/golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

check: fmt vet lint ## Run all code quality checks

install: build-all ## Install binaries to $GOPATH/bin
	@echo "Installing binaries..."
	cp $(BINARY_NAME) $(GOPATH)/bin/
	cp $(HEALTHCHECK_BINARY) $(GOPATH)/bin/
	@echo "Binaries installed to $(GOPATH)/bin"

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

.DEFAULT_GOAL := help
