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
| `make test-e2e-js` | Run Playwright E2E tests against the test image. Full stack service debugging with parameters. |
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

Integration and E2E tests build and use `jam-build-propsdb-test` image to run the tests.

---

## 4. Debugging Tests

The project's Makefile exposes debug commands that start targeted code in the Delve debugger. Delve waits for you to connect to port 2345 with the dlv command or a comparable IDE launch configuration.

### Visual Studio Code Launch Configurations

The Visual Studio Code launch configuration to attach to Delve to debug the Unit/Integration/E2E Go tests or the Go Testcontainers Orchestrator:
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

The Visual Studio Code launch configuration to attach to Delve to debug the propsdb-api service running from the jam-build-propsdb-test image:
```json
{
  "name": "Attach to PropsDB (in Testcontainer)",
  "type": "go",
  "request": "attach",
  "mode": "remote",
  "port": 2345, // Testcontainers forces this fixed local port when env has DEBUG_CONTAINER=1
  "host": "127.0.0.1",
  "substitutePath": [
    {
      "from": "${workspaceFolder}",
      "to": "/app"
    }
  ]
}
```

### Debugging Commands for the Go Test Suites

```bash
# Targets the Go unit tests
make test-unit-debug

# Targets the Go integration tests
make test-integration-debug

# Targets the Go E2E tests 
make test-e2e-debug # add REBUILD=1 to rebuild the jam-build-propsdb-test image
```

### Debugging the propsdb-api service against the Playwright E2E Tests

To debug the propsdb-api service in a TestContainer:
1. Run `make test-e2e-js-debug` (alias for `make test-e2e-js DEBUG=2`).
2. The jam-build-propsdb-test image will be built with the Delve debugger and the Testcontainers orchestrator will start the containers and pause.
3. Attach your IDE debugger to `:2345` and set breakpoints.
4. Press `Enter` in the terminal to resume the Playwright tests.

> To restart debugging without rebuilding the test image, run `make test-e2e-js DEBUG=1`.

```bash
# Targets the propsdb-api service running from a debug build of jam-build-propsdb-test image

make test-e2e-js-debug # alias for test-e2e-js DEBUG=2, full debug rebuild, then start debugger.

# and/or:

make test-e2e-js DEBUG=1 # Same, but restarts the debugger without a rebuild
```

### Debugging Coverage extraction at the end of the Testcontainer Orchestrator process

To debug the coverage extraction in the Testcontainer Orchestrator:
1. Terminal 1: Run `make test-e2e-js-cover REBUILD=1 HOST_DEBUG=1`.
  - Follow up debugging sessions, or sessions that don't require a coverage rebuild of jam-build-propsdb-test, just use `make test-e2e-js-cover HOST_DEBUG=1`.
2. Terminal 2: Open a second terminal and run `make test-e2e-js-orchestrator-debug`.
3. Set breakpoints in `tests/helpers/testcontainers.go#collectCoverage` and attach your IDE debugger to `:2345`.
4. In Terminal 2, wait until completion - 'PropsDB testcontainer started' will be displayed.
5. In Terminal 1, Press `Enter` to resume/run the Playwright tests.
6. In Terminal 2, Press `Enter` or `Ctrl+C` to trigger the coverage collection.
7. Debug. Your breakpoints will be hit. When complete:
8. In Terminal 1, Press `Enter` to exit.
9. Kill any left over stopped `dlv` process (`pkill dlv`).

> To restart debugging without rebuilding the coverage test image, run `make test-e2e-js-cover HOST_DEBUG=1`.

```bash
# Targets the coverage extraction code in the Testcontainer Orchestrator process

make test-e2e-js-cover REBUILD=1 HOST_DEBUG=1 # Full coverage rebuild, then waits for input to start the Playwright tests. Omit REBUILD=1 for repeat runs.

make test-e2e-js-orchestrator-debug # Run in new terminal, builds debug Testcontainers and waits for debugger attachment.

# Follow the rest of the debugging instructions above.
```
---

## 5. Continuous Integration

In CI environments (e.g., GitHub Actions), we skip the specialized local debug targets and run:
1. `make test-e2e-js-cover` (E2e Playwright tests)
2. `make test-coverage` (Unit, Integration, and E2e Go tests)

Coverage is collected and reported locally, which can be uploaded to services like Codecov.

[More detail about coverage](COVERAGE.md)

---

## 6. Local Testing

There are scripts to run the Playwright js tests against a production build of the full service stack running in Docker locally.

### Prerequisites for Local Testing
- Docker Desktop must be running.
- Make, Go, and Node.js must be installed.
- Local ports used by the services to be available.

```bash
# Setup for MariaDB

make DB_TYPE=mariadb docker-compose-clean # Clear out any previous containers and data

make DB_TYPE=mariadb docker-compose-up # Start the containers, use BUILD=1 to rebuild the images

make test-e2e-js-local # Run the playwright e2e tests against the local stack
```
