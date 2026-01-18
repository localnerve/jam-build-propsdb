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

| Command | Description |
|---------|-------------|
| `make test-unit` | Run unit tests only. |
| `make test-integration` | Run integration tests (requires Docker). |
| `make test-e2e` | Run Go-based smoke tests (requires Docker). |
| `make test-e2e-js` | Run Playwright E2E tests (requires Docker). |
| `make test-e2e-js-cover` | Run Playwright E2E tests and generate coverage report (requires Docker). |
| `make test-coverage` | Run all Go tests and generate coverage report. |
| `make test-all` | Run the unit, integration, and Go E2E tests (requires Docker). |

### Parameters
Many targets support parameters for advanced usage:
- `test-e2e-js`: Params: DEBUG=1 (debug, no rebuild), DEBUG=2 (debug, full rebuild)
- `test-e2e-js-cover`: Params: REBUILD=1 (rebuild orchestrator), HOST_DEBUG=1 (debug host), OPEN=1 (optional)
- `test-coverage`: Params: OPEN=1 (optional)

---

## 3. Database Testing

Integration tests use `testcontainers-go` to spin up ephemeral database instances. This ensures:
1. **Parallelism**: Multiple test runs don't interfere with each other.
2. **Real Logic**: We test against actual SQL execution, not mocks or stubs.
3. **Multi-DB Support**: We verify compatibility across MariaDB and PostgreSQL automatically.

### Prerequisites for Integration/E2E
- Docker Desktop must be running.
- Port permissions (the tests will find available ports automatically).

---

## 4. Debugging Tests

### Go Tests (Unit/Integration)
You can use standard Go flags or attach with Delve:
```bash
go test -v ./tests/unit/... -run TestGetAppProperties
```

### Playwright E2E
To debug the full stack:
1. Run `make test-e2e-js DEBUG=2`.
2. The orchestrator will start the containers and pause.
3. Attach your IDE debugger to `:2345`.
4. Press `Enter` in the terminal to resume the Playwright tests.

---

## 5. Continuous Integration

In CI environments (e.g., GitHub Actions), we skip the specialized local debug targets and run:
1. `make test-coverage` (Unit, Integration, and E2e Go tests)
2. `make test-e2e-js-cover` (E2e Playwright tests)

Coverage is collected and reported locally, which can be uploaded to services like Codecov.
