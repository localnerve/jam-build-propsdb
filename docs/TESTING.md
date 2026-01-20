# Jam-Build-PropsDB - Testing Guide

This project follows a 4-layered testing strategy to ensure reliability across unit, integration, and end-to-end scenarios. All tests run against real or in-memory databases using standard Go tools, `testcontainers-go`, and Playwright.

## 1. Testing Strategy

### 🟢 Layer 1: Unit Tests
- **Package**: `tests/unit/`
- **Scope**: HTTP Handlers and request/response logic.
- **Database**: In-memory SQLite.
- **Run**: `make test-unit`
- **Speed**: Very Fast (< 1s).

### 🔵 Layer 2: Integration Tests
- **Package**: `tests/integration/`
- **Scope**: Business logic in `internal/services/` and database interactions.
- **Database**: Real MariaDB and PostgreSQL containers via `testcontainers-go`.
- **Run**: `make test-integration`
- **Speed**: Medium (15-30s).

### 🟠 Layer 3: Go End-to-End (E2E)
- **Package**: `tests/e2e/`
- **Scope**: Service smoke tests, health checks, Swagger UI, and Prometheus metrics.
- **Database**: Full stack orchestration.
- **Run**: `make test-e2e`
- **Speed**: Slow (30-60s).

### 🔴 Layer 4: Playwright End-to-End (JS)
- **Package**: `test-e2e-js/`
- **Scope**: Full application functional tests from a client perspective.
- **Orchestration**: Custom Go orchestrator (`cmd/testcontainers`) manages the life-cycle of the API, Database, Cache, and Authorizer.
- **Run**: `make test-e2e-js`
- **Speed**: Slowest (1-2m).

---

## 2. Common Makefile Commands

> Integration and E2E tests require Docker to be running.

| Command | Description |
|---------|-------------|
| `make test-unit` | Run unit tests only. |
| `make test-integration` | Run integration tests. |
| `make test-e2e` | Run Go-based smoke tests. |
| `make test-all` | Run the unit, integration, and Go E2E tests. |
| `make test-coverage` | Run all Go tests and generate coverage report. |
| `make test-e2e-js` | Run Playwright E2E tests against the test image. |
| `make test-e2e-js-cover` | Run Playwright E2E tests against the test image and generate coverage report. |

### Parameters
Some targets support parameters for advanced usage:
- `test-e2e`: Params: **REBUILD=1** (rebuild test image)
- `test-e2e-js`: Params: **DEBUG=1** (debug against debug test image, no rebuild), **DEBUG=2** (debug against debug test image, full rebuild test image with debugger)
- `test-e2e-js-cover`: Params: **REBUILD=1** (rebuild orchestrator), **HOST_DEBUG=1** (debug host), **OPEN=1** (open in browser)
- `test-coverage`: Params: **OPEN=1** (open in browser)

---

## 3. Database Testing

Integration and E2E tests use `testcontainers-go` to spin up ephemeral database instances. This ensures:
1. **Parallelism**: Multiple test runs don't interfere with each other.
2. **Real Logic**: We test against actual SQL execution, not mocks or stubs.
3. **Multi-DB Support**: We verify compatibility across MariaDB and PostgreSQL automatically.

### Prerequisites for Integration/E2E Tests
- Docker Desktop must be running.
- Port permissions (the tests will find available ports automatically).

---

## 4. Debugging Tests

The debug commands start the test suites in the Delve debugger. Delve waits for you to connect to port 2345 with the dlv command or a comparable IDE launch configuration.

### Debug Go Tests (Unit/Integration/E2E)

Here is the Visual Studio Code launch configuration to attach to Delve:
```json
{
  "name": "Attach to Delve (in Test)",
  "type": "go",
  "request": "attach",
  "mode": "remote",
  "remotePath": "${workspaceFolder}",
  "port": 2345,
  "host": "127.0.0.1"
}
```

Here are the debugging commands. They each target a specific set of code, do the prep work and pause for debugger attachment on port 2345. Some debugging scenarios are more complex, and require a sequence of commands.
```bash
# Targets the Go unit tests
make test-unit-debug

# Targets the Go integration tests
make test-integration-debug

# Targets the Go E2E tests 
make test-e2e-debug # add REBUILD=1 to rebuild the jam-build-propsdb-test image

# Targets the propsdb-api service running in a debug build of jam-build-propsdb-test image
make test-e2e-js-debug # alias for test-e2e-js DEBUG=2, full debug rebuild, then start debugger.

# Same but restarts the debugger without a jam-build-propsdb-test image rebuild
make test-e2e-js DEBUG=1
```

### Debug the propsdb-api service with Playwright E2E Tests

To debug the propsdb-api service in a TestContainer:
1. Run `make test-e2e-js DEBUG=2`.
2. The jam-build-propsdb-test image will be built with the Delve debugger and the orchestrator will start the containers and pause.
3. Attach your IDE debugger to `:2345`.
4. Press `Enter` in the terminal to resume the Playwright tests.

> To restart debugging without rebuilding the test image, run `make test-e2e-js DEBUG=1`.

---

## 5. Continuous Integration

In CI environments (e.g., GitHub Actions), we skip the specialized local debug targets and run:
1. `make test-coverage` (Unit, Integration, and E2e Go tests)
2. `make test-e2e-js-cover` (E2e Playwright tests)

Coverage is collected and reported locally, which can be uploaded to services like Codecov.

---

## 6. Local Testing

There are scripts to run the Playwright js tests against a build of the services running in Docker locally.

### Prerequisites for Local Testing
- Docker Desktop must be running.
- Make, Go, and Node.js must be installed.
- Local ports used by the services to be available.

```bash
# Example for MariaDB
make DB_TYPE=mariadb docker-compose-clean # Clear out any previous containers and data
make DB_TYPE=mariadb docker-compose-up # Start the containers, use BUILD=1 to rebuild the images
make test-e2e-js-local # Run the playwright e2e tests against the local stack
```
