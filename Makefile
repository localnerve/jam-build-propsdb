.PHONY: help build build-healthcheck build-testcontainers build-testcontainers-debug build-all check clean coverage-report deps dev install install-tools test test-unit test-integration test-e2e test-e2e-js test-e2e-js-debug test-e2e-debug test-e2e-rebuild test-cache-clean test-all test-coverage run run-testcontainers docker-build docker-run docker-compose-up docker-compose-down docker-compose-logs swagger swagger-serve lint fmt vet

# Variables
BINARY_NAME=propsdb
HEALTHCHECK_BINARY=healthcheck
TESTCONTAINERS_BINARY=testcontainers
DOCKER_IMAGE=propsdb
DOCKER_TAG=latest
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html
SWAGGER_DIR=docs/api

# Docker parameters
ENV_FILE=.env.dev

# Commands
GOCMD=go
DLVCMD=dlv
NPXCMD=npx
GODOTENVCMD=godotenv
GODOTENV=$(GODOTENVCMD) -f $(ENV_FILE)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GODOTENV) $(GOCMD) test
DLVTEST=$(GODOTENV) $(DLVCMD) test
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

build-testcontainers: ## Build the testcontainers binary
	@echo "Building $(TESTCONTAINERS_BINARY)..."
	$(GOBUILD) -o $(TESTCONTAINERS_BINARY) ./cmd/testcontainers
	@echo "Build complete: $(TESTCONTAINERS_BINARY)"

build-testcontainers-debug: ## Prepare for a new propsdb-test image build
	@echo "Building $(TESTCONTAINERS_BINARY) with debug propsdb container test image..."
	docker rmi propsdb-test:latest || true
	@rm -f .env.debug
	@cp $(ENV_FILE) .env.debug
	@printf '\n' >> .env.debug
	@echo "DEBUG_CONTAINER=true" >> .env.debug
	@echo "TESTCONTAINERS_BUILD_CONTEXT=." >> .env.debug
	$(GOBUILD) -o $(TESTCONTAINERS_BINARY) ./cmd/testcontainers
	@echo "Build complete: $(TESTCONTAINERS_BINARY)"

build-all: build build-healthcheck build-testcontainers ## Build all binaries
	@echo "All binaries built successfully"
	@ls -lh $(BINARY_NAME) $(HEALTHCHECK_BINARY) $(TESTCONTAINERS_BINARY)

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(HEALTHCHECK_BINARY) $(TESTCONTAINERS_BINARY)
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

test-e2e-debug: ## Start debugger for E2E tests, attach with 'dlv connect :2345' or comparable IDE launch configuration
	@echo "Running E2E tests in debug mode..."
	$(DLVTEST) ./tests/e2e/... --headless --listen=:2345 --api-version=2 --log

test-e2e-rebuild: ## Run E2E tests with forced rebuild of propsdb-test image
	@echo "Forcing rebuild of propsdb-test images..."
	docker rmi propsdb-test:latest || true
	@echo "Running E2E tests..."
	$(GOTEST) -v ./tests/e2e/... -timeout 300s

test-e2e-js: build-testcontainers ## Run end-to-end tests with full stack (requires Docker) FROM Playwright browsers
	@{ \
		echo "Starting testcontainers..." ; \
		./$(TESTCONTAINERS_BINARY) -f $(ENV_FILE) > testcontainers.log 2>&1 & \
		TCPID=$$!; \
		\
		# Portable polling loop \
		count=0; \
		while ! grep -q "PropsDB testcontainer started" testcontainers.log; do \
			if [ $$count -ge 30 ]; then \
				echo "Timeout: Failed to start"; kill $$TCPID 2>/dev/null; exit 1; \
			fi; \
			printf '%s' "."; \
			sleep 1; count=`expr $$count + 1`; \
		done; \
		\
		echo "\nReady! Running E2E tests..."; \
		echo $$(awk -F'=' '/AUTHZ_URL/ {print $$1"=""http://"$$2; exit}' testcontainers.log) > .env.test; \
		echo $$(awk -F'=' '/BASE_URL/ {print $$1"=""http://"$$2; exit}' testcontainers.log) >> .env.test; \
		$(GODOTENVCMD) -f .env.test,$(ENV_FILE) $(NPXCMD) playwright test --project api-chromium; \
		EXIT_CODE=$$?; \
		\
		echo "Cleaning up..."; \
		kill $$TCPID 2>/dev/null || pkill -f $(TESTCONTAINERS_BINARY) || true; \
		\
		exit $$EXIT_CODE; \
	}

test-e2e-js-debug: build-testcontainers-debug ## Run end-to-end tests with full stack (requires Docker) FROM Playwright browsers IN DEBUG MODE
	@{ \
		echo "Rebuilding propsdb-test image as debug container and starting testcontainers..." ; \
		./$(TESTCONTAINERS_BINARY) -f .env.debug > testcontainers.log 2>&1 & \
		TCPID=$$!; \
		\
		# Portable polling loop \
		count=0; \
		while ! grep -q "PropsDB testcontainer started" testcontainers.log; do \
			if [ $$count -ge 120 ]; then \
				echo "Timeout: Failed to start"; kill $$TCPID 2>/dev/null; exit 1; \
			fi; \
			if [ "$$count" -ne 0 -a "`expr $$count % 20`" -eq 0 ]; then \
				echo ""; \
			fi; \
			printf '%s' "."; \
			sleep 1; count=`expr $$count + 1`; \
		done; \
		\
		echo "\nContainers ready!"; \
		echo "Attach debugger to :2345 and press enter to start E2E tests..."; \
		read -r dummy; \
		echo "Starting E2E tests..."; \
		echo $$(awk -F'=' '/AUTHZ_URL/ {print $$1"=""http://"$$2; exit}' testcontainers.log) > .env.test; \
		echo $$(awk -F'=' '/BASE_URL/ {print $$1"=""http://"$$2; exit}' testcontainers.log) >> .env.test; \
		$(GODOTENVCMD) -f .env.test,.env.debug $(NPXCMD) playwright test --project api-chromium; \
		EXIT_CODE=$$?; \
		\
		echo "Cleaning up..."; \
		kill $$TCPID 2>/dev/null || pkill -f $(TESTCONTAINERS_BINARY) || true; \
		\
		exit $$EXIT_CODE; \
	}

test-cache-clean: ## force test cache cleanup
	@echo "Cleaning test cache..."
	go clean -testcache

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

run-testcontainers: build-testcontainers ## Build and run the testcontainers binary
	@echo "Running testcontainers binary..."
	$(GODOTENV) ./$(TESTCONTAINERS_BINARY)

docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 3000:3000 --env-file $(ENV_FILE) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start all services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker-compose --env-file $(ENV_FILE) up -d
	@echo "Services started. Use 'make docker-compose-logs' to view logs"

docker-compose-down: ## Stop all Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker-compose --env-file $(ENV_FILE) down

docker-compose-logs: ## View Docker Compose logs
	docker-compose --env-file $(ENV_FILE) logs -f

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

install-tools: ## Install tools
	@echo "Installing tools..."
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/joho/godotenv/cmd/godotenv@latest

install: build-all ## Install binaries to $GOPATH/bin
	@echo "Installing binaries..."
	cp $(BINARY_NAME) $(GOPATH)/bin/
	cp $(HEALTHCHECK_BINARY) $(GOPATH)/bin/
	@echo "Binaries installed to $(GOPATH)/bin"

dev: ## Run in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

.DEFAULT_GOAL := help
