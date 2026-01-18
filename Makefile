.PHONY: help build build-healthcheck build-testcontainers build-testcontainers-debug clean deps test test-unit test-integration test-e2e test-e2e-debug test-e2e-rebuild test-e2e-js test-e2e-js-debug test-e2e-js-cover test-e2e-js-host-debug test-cache-clean test-all test-coverage report-coverage docker-compose-up docker-compose-down docker-compose-logs obs-up obs-down obs-logs swagger lint fmt vet check install-tools

# Variables
BINARY_NAME=jam-build-propsdb
HEALTHCHECK_BINARY=healthcheck
TESTCONTAINERS_BINARY=testcontainers
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html
TESTCONTAINERS_LOG=testcontainers.log
SWAGGER_DIR=docs/api

# Docker parameters
ENV_FILE=.env.dev
ENV_DOCKER_FILE=.env.docker

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
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

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

build-testcontainers-debug: ## Prepare for a new jam-build-propsdb-test image build
	@echo "Rebuilding $(TESTCONTAINERS_BINARY) for debug..."
	docker rmi jam-build-propsdb-test:latest || true
	$(MAKE) build-testcontainers

clean: ## Remove build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME) $(HEALTHCHECK_BINARY) $(TESTCONTAINERS_BINARY)
	rm -f $(ENV_DOCKER_FILE)
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

test-e2e: ## Run end-to-end go tests with full stack (requires Docker)
	@echo "Running E2E tests..."
	$(GOTEST) -v ./tests/e2e/... -timeout 300s

test-e2e-debug: ## Start debugger for E2E go tests, attach with 'dlv connect :2345' or comparable IDE launch configuration
	@echo "Running E2E tests in debug mode..."
	$(DLVTEST) ./tests/e2e/... --headless --listen=:2345 --api-version=2 --log

test-e2e-rebuild: ## Run E2E go tests with forced rebuild of jam-build-propsdb-test image
	@echo "Forcing rebuild of jam-build-propsdb-test images..."
	docker rmi jam-build-propsdb-test:latest || true
	@$(MAKE) test-e2e

test-e2e-js: ## Run end-to-end tests with full stack. Params: DEBUG=1 (debug, no rebuild), DEBUG=2 (debug, full rebuild)
	@{ \
		DEBUG_VAL=$(DEBUG); \
		[ -z "$$DEBUG_VAL" ] && DEBUG_VAL=0; \
		if [ "$$DEBUG_VAL" -gt 0 ]; then \
			echo "Setting up .env.debug..." ; \
			rm -f .env.debug; \
			cp $(ENV_FILE) .env.debug; \
			printf '\n' >> .env.debug; \
			echo "DEBUG_CONTAINER=true" >> .env.debug; \
			echo "TESTCONTAINERS_BUILD_CONTEXT=." >> .env.debug; \
			ENV_FILE_TO_USE=.env.debug; \
			TIMEOUT=120; \
		else \
			ENV_FILE_TO_USE=$(ENV_FILE); \
			TIMEOUT=30; \
		fi; \
		if [ "$$DEBUG_VAL" -eq 2 ]; then \
			$(MAKE) build-testcontainers-debug; \
		else \
			$(MAKE) build-testcontainers; \
		fi; \
		echo "Starting testcontainers with $$ENV_FILE_TO_USE..." ; \
		./$(TESTCONTAINERS_BINARY) -f $$ENV_FILE_TO_USE > $(TESTCONTAINERS_LOG) 2>&1 & \
		TCPID=$$!; \
		count=0; \
		while ! grep -q "PropsDB testcontainer started" $(TESTCONTAINERS_LOG); do \
			if [ $$count -ge $$TIMEOUT ]; then \
				echo "Timeout: Failed to start"; kill $$TCPID 2>/dev/null; exit 1; \
			fi; \
			if [ "$$DEBUG_VAL" -gt 0 -a "$$count" -ne 0 -a "`expr $$count % 20`" -eq 0 ]; then \
				echo ""; \
			fi; \
			printf '%s' "."; \
			sleep 1; count=`expr $$count + 1`; \
		done; \
		if [ "$$DEBUG_VAL" -gt 0 ]; then \
			echo "\nContainers ready!"; \
			echo "Attach debugger to :2345 and press enter to start E2E tests..."; \
			read -r dummy; \
			echo "Starting E2E tests..."; \
		else \
			echo "\nReady! Running E2E tests..."; \
		fi; \
		echo $$(awk -F'=' '/AUTHZ_URL/ {print $$1"=""http://"$$2; exit}' $(TESTCONTAINERS_LOG)) > .env.test; \
		echo $$(awk -F'=' '/BASE_URL/ {print $$1"=""http://"$$2; exit}' $(TESTCONTAINERS_LOG)) >> .env.test; \
		$(GODOTENVCMD) -f .env.test,$$ENV_FILE_TO_USE $(NPXCMD) playwright test --project api-chromium; \
		EXIT_CODE=$$?; \
		\
		echo "Cleaning up..."; \
		kill $$TCPID 2>/dev/null || pkill -f $(TESTCONTAINERS_BINARY) || true; \
		\
		exit $$EXIT_CODE; \
	}

test-e2e-js-debug: ## Run E2E tests in debug mode (alias for DEBUG=2)
	@$(MAKE) test-e2e-js DEBUG=2

test-e2e-js-cover: ## Run E2E tests with coverage collection. Params: REBUILD=1 (rebuild orchestrator), HOST_DEBUG=1 (debug host)
	@{ \
	  rm -rf $(COVERAGE_DIR)/e2e-js; \
		REBUILD_VAL=$(REBUILD); \
		[ -z "$$REBUILD_VAL" ] && REBUILD_VAL=0; \
		echo "Setting up .env.cover..." ; \
		rm -f .env.cover; \
		cp $(ENV_FILE) .env.cover; \
		printf '\n' >> .env.cover; \
		echo "COVERAGE_DIR=$(COVERAGE_DIR)/e2e-js" >> .env.cover; \
		echo "COLLECT_COVERAGE=true" >> .env.cover; \
		echo "TESTCONTAINERS_BUILD_CONTEXT=." >> .env.cover; \
		if [ "$(HOST_DEBUG)" = "1" ]; then \
			echo "HOST_DEBUG=true" >> .env.cover; \
		fi; \
		ENV_FILE_TO_USE=.env.cover; \
		TIMEOUT=120; \
		if [ "$$REBUILD_VAL" -eq 1 ]; then \
			$(MAKE) build-testcontainers-debug; \
		else \
			$(MAKE) build-testcontainers; \
		fi; \
		if [ "$(HOST_DEBUG)" = "1" ]; then \
			echo "\nHOST_DEBUG enabled!"; \
			echo "1. Open a new terminal and run: make test-e2e-js-orchestrator-debug"; \
			echo "2. Set your breakpoints in tests/helpers/testcontainers.go (e.g., collectCoverage)"; \
			echo "3. Type 'continue' in the debugger."; \
			echo "4. Wait for 'PropsDB testcontainer started' and then press enter here..."; \
			read -r dummy; \
		else \
			echo "Starting testcontainers with $$ENV_FILE_TO_USE..." ; \
			./$(TESTCONTAINERS_BINARY) -f $$ENV_FILE_TO_USE > $(TESTCONTAINERS_LOG) 2>&1 & \
			TCPID=$$!; \
			count=0; \
			while ! grep -q "PropsDB testcontainer started" $(TESTCONTAINERS_LOG); do \
				if [ $$count -ge $$TIMEOUT ]; then \
					echo "Timeout: Failed to start"; kill $$TCPID 2>/dev/null; exit 1; \
				fi; \
				printf '%s' "."; \
				sleep 1; count=`expr $$count + 1`; \
			done; \
		fi; \
		echo "\nReady! Running E2E tests with coverage..."; \
		echo $$(awk -F'=' '/AUTHZ_URL/ {print $$1"=""http://"$$2; exit}' $(TESTCONTAINERS_LOG)) > .env.test; \
		echo $$(awk -F'=' '/BASE_URL/ {print $$1"=""http://"$$2; exit}' $(TESTCONTAINERS_LOG)) >> .env.test; \
		$(GODOTENVCMD) -f .env.test,$$ENV_FILE_TO_USE $(NPXCMD) playwright test --project api-chromium; \
		EXIT_CODE=$$?; \
		\
		echo "Cleaning up and collecting coverage..."; \
		if [ "$(HOST_DEBUG)" = "1" ]; then \
			echo "Please stop the debugger (Ctrl+C and 'quit' or 'exit') to trigger coverage collection, then press enter here."; \
			read -r dummy; \
		else \
			kill $$TCPID 2>/dev/null || pkill -f $(TESTCONTAINERS_BINARY) || true; \
		fi; \
		sleep 3; # wait for coverage collection to finish\
		echo "Coverage extraction log can be found in $(TESTCONTAINERS_LOG)"; \
		$(MAKE) report-coverage DIR=$(COVERAGE_DIR)/e2e-js TITLE="E2E JS COVERAGE SUMMARY" OPEN=$(OPEN); \
		exit $$EXIT_CODE; \
	}

report-coverage: ## Generate and display coverage report. Params: DIR=path (required), TITLE=header (required), OPEN=1|0 (optional)
	@if [ -z "$(DIR)" ] || [ -z "$(TITLE)" ]; then \
		echo "Error: DIR and TITLE are required parameters."; \
		exit 1; \
	fi; \
	OUT_FILE=$(DIR)/coverage.out; \
	HTML_FILE=$(DIR)/coverage.html; \
	echo "\n================================================================================" ; \
	echo "$(TITLE)" ; \
	echo "================================================================================" ; \
	if [ -f "$$OUT_FILE" ]; then \
		echo "Using existing coverage profile: $$OUT_FILE"; \
	elif [ -d "$(DIR)" ] && [ "$$(ls -A $(DIR) 2>/dev/null)" ]; then \
		echo "Processing binary coverage data in $(DIR)..."; \
		$(GOCMD) tool covdata percent -i=$(DIR) ; \
		$(GOCMD) tool covdata textfmt -i=$(DIR) -o=$$OUT_FILE ; \
	else \
		echo "No coverage data found in $(DIR)"; \
		echo "================================================================================\n" ; \
		exit 0; \
	fi; \
	echo "--------------------------------------------------------------------------------" ; \
	$(GOCMD) tool cover -func=$$OUT_FILE | grep -v "100.0%" | head -n 20 ; \
	echo "--------------------------------------------------------------------------------" ; \
	echo "Generating HTML coverage report..." ; \
	$(GOCMD) tool cover -html=$$OUT_FILE -o=$$HTML_FILE ; \
	echo "Coverage report available at: $$HTML_FILE" ; \
	if [ "$(OPEN)" = "1" ]; then \
		echo "Opening coverage report in browser..." ; \
		open $$HTML_FILE 2>/dev/null || xdg-open $$HTML_FILE 2>/dev/null || echo "Please open file://$(PWD)/$$HTML_FILE manually" ; \
	fi; \
	echo "================================================================================\n"

test-e2e-js-orchestrator-debug: build-testcontainers ## Run the testcontainers binary under Delve (headless) for VS Code attachment.
	@ENV_TO_USE=$(ENV_FILE); \
	if [ -f .env.cover ]; then ENV_TO_USE=.env.cover; fi; \
	echo "Starting testcontainers orchestrator in headless debug mode with $$ENV_TO_USE..." ; \
	echo "Listening on :2345. Use VS Code 'Attach to Delve (in Test)' to begin." ; \
	$(DLVCMD) debug ./cmd/testcontainers --headless --listen=:2345 --api-version=2 --accept-multiclient --log -- -f $$ENV_TO_USE | tee $(TESTCONTAINERS_LOG)

test-e2e-js-host-debug: build-testcontainers ## Debug the testcontainers host process itself using Delve.
	@echo "Starting testcontainers host process in debug mode..."
	$(DLVCMD) debug ./cmd/testcontainers -- -f $(ENV_FILE)

test-cache-clean: ## force test cache cleanup
	@echo "Cleaning test cache..."
	go clean -testcache

test-all: ## Run all tests including integration and E2E
	@echo "Running all tests..."
	$(GOTEST) -v ./tests/...

test-coverage: ## Run tests with coverage report. Params: OPEN=1 (optional)
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic -coverpkg=./internal/... ./...
	@$(MAKE) report-coverage DIR=$(COVERAGE_DIR) TITLE="GO TEST COVERAGE SUMMARY" OPEN=$(OPEN)

$(ENV_DOCKER_FILE): $(ENV_FILE)
	@echo "Generating $(ENV_DOCKER_FILE) from $(ENV_FILE)..."
	@sed 's/localhost/host.docker.internal/g' $(ENV_FILE) > $(ENV_DOCKER_FILE)

docker-compose-up: $(ENV_DOCKER_FILE) ## Start all services with Docker Compose
	@echo "Starting Docker Compose services..."
	docker-compose --env-file $(ENV_DOCKER_FILE) up -d
	@echo "Services started. Use 'make docker-compose-logs' to view logs"

docker-compose-down: $(ENV_DOCKER_FILE) ## Stop all Docker Compose services
	@echo "Stopping Docker Compose services..."
	docker-compose --env-file $(ENV_DOCKER_FILE) down

docker-compose-logs: $(ENV_DOCKER_FILE) ## View Docker Compose logs
	docker-compose --env-file $(ENV_DOCKER_FILE) logs -f

obs-up: ## Start observability services
	@echo "Starting observability services..."
	docker-compose -f docker-compose.observability.yml up -d

obs-down: ## Stop observability services
	@echo "Stopping observability services..."
	docker-compose -f docker-compose.observability.yml down

obs-logs: ## View observability logs
	docker-compose -f docker-compose.observability.yml logs -f

swagger: ## Generate OpenAPI/Swagger documentation
	@echo "Generating Swagger documentation..."
	@mkdir -p $(SWAGGER_DIR)
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$$(go env GOPATH)/bin/swag init -g cmd/server/main.go -o $(SWAGGER_DIR)
	@echo "Swagger documentation generated in $(SWAGGER_DIR)"

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
	go install github.com/joho/godotenv/cmd/godotenv@latest

.DEFAULT_GOAL := help
