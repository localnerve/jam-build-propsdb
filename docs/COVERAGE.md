# Testing Guide

## Understanding Testing

The jam-build-propsdb project has multiple test levels:

### 1. Unit Tests (`tests/unit/`)
- **Purpose**: Test handlers in isolation with mocked dependencies
- **Coverage**: Handlers layer only
- **Speed**: Very fast (in-memory SQLite)
- **Run**: `make test-unit`

### 2. Integration Tests (`tests/integration/`)
- **Purpose**: Test services with real databases
- **Coverage**: Services, database, models
- **Speed**: Medium (testcontainers)
- **Run**: `make test-integration`

### 3. End-to-End Service Tests (`tests/e2e/`)
- **Purpose**: Test the health of the service stack
- **Coverage**: Not applicable
- **Speed**: Slow (builds Docker image, starts full stack)
- **Run**: `make test-e2e`

### 4. End-to-End Full Application Tests (`tests/e2e-js/`)
- **Purpose**: Test full application service stack from all browsers
- **Coverage**: Full application including main.go, middleware, all internal packages
- **Speed**: Slow (builds Docker image, starts full stack)
- **Run**: `make test-e2e-js` for test-only or `make test-e2e-js-cover` for coverage
- **Parameters**: `make test-e2e-js DEBUG=0` (no debug), `make test-e2e-js DEBUG=1` (debug), `make test-e2e-js DEBUG=2` (debug, rebuild), `make test-e2e-js REBUILD=1` (rebuild), `make test-e2e-js HOST_DEBUG=1` (debug host)

## Current Coverage Limitations

The standard `make test-coverage` command shows **0% coverage** for `internal/` packages because:

1. **Unit tests** use isolated handlers with mocked dependencies
2. **Integration tests** directly call service functions, bypassing HTTP layer
3. **Neither exercises the actual HTTP server, middleware, or main.go**

## Getting Real Coverage

### Option 1: Run E2E Tests (Recommended)
```bash
# This starts the full service and exercises all code paths
make test-e2e
```

**What E2E tests cover:**
- ✅ `cmd/server/main.go` - Server initialization
- ✅ `internal/middleware/*` - All middleware (auth, version, prometheus)
- ✅ `internal/handlers/*` - HTTP handlers
- ✅ `internal/services/*` - Business logic
- ✅ `internal/database/*` - Database connections
- ✅ `internal/config/*` - Configuration loading
- ✅ Swagger UI endpoint
- ✅ Prometheus metrics endpoint
- ✅ Health check functionality

### Option 2: Manual Coverage with Running Service
```bash
# Start the service
docker-compose up -d

# Run manual tests and collect coverage
# (This requires instrumenting the binary, which is complex)
```

### Option 3: Add More Integration Tests
You can add integration tests that go through the HTTP layer:

```go
// Example: HTTP-level integration test
func TestHTTPEndpoint(t *testing.T) {
    app := fiber.New()
    // Setup routes...
    
    req := httptest.NewRequest("GET", "/api/data/app", nil)
    resp, _ := app.Test(req)
    // Assert...
}
```

## Coverage Best Practices

### For Development
- Run `make test-unit` frequently (fast feedback)
- Run `make test-integration` before commits
- Run `make test-e2e` before releases

### For CI/CD
```yaml
# GitHub Actions example
- name: Unit Tests
  run: make test-unit

- name: Integration Tests  
  run: make test-integration

- name: E2E Tests
  run: make test-e2e
```

### Coverage Goals
- **Unit Tests**: 80%+ of handlers
- **Integration Tests**: 80%+ of services
- **E2E Tests**: Critical user paths

## Why 0% Coverage is Misleading

The coverage report shows 0% because:

1. **Unit tests** create their own Fiber app instance - doesn't touch `main.go`
2. **Integration tests** call service functions directly - doesn't touch HTTP layer
3. **Coverage tool** only measures code executed in the same process

**Reality**: The code IS tested, just not in a way that shows up in coverage reports.

## Improving Coverage Reporting

To get accurate coverage numbers, you would need to:

1. **Instrument the binary** with coverage flags
2. **Run the service** as part of tests
3. **Collect coverage** from the running process
4. **Merge coverage** from multiple test runs

This is complex and typically not worth it for most projects. Instead:
- Trust that E2E tests exercise the code
- Use integration tests for service logic
- Use unit tests for edge cases
- Monitor production metrics for real-world coverage

## Recommended Approach

```bash
# Development workflow
make test-unit          # Fast feedback
make test-integration   # Before commit
make test-e2e          # Before PR/release

# Full test suite
make test-all          # Runs everything
```

## Notes

- E2E tests take ~2-3 minutes (builds Docker image)
- Integration tests take ~30 seconds (starts containers)
- Unit tests take <1 second (in-memory)
- Choose the right test level for what you're testing
