# Jam-Build-PropsDB - Coverage Guide

This project implements a unified coverage reporting system that combines traditional Go unit/integration test coverage with modern binary coverage collected from containerized E2E tests.

## 1. Unified Reporting System

Coverage is managed via a generic `report-coverage` target in the `Makefile`. This target handles data ingestion, processing, terminal summaries, and HTML generation.

### Key Logic
- **Data Detection**: Automatically detects if the target directory contains a text-based `coverage.out` profile or binary Go `covdata` files.
- **Processing**: Converts binary data to text formats if necessary using `go tool covdata`.
- **Summarization**: Displays a terminal-friendly summary of the top 20 functions/packages that are NOT yet 100% covered.
- **HTML Generation**: Produces a detailed `coverage.html` report.

---

## 2. Go Unit and Integration Coverage

The project uses a specialized approach to ensure that tests located in the `tests/` package (outside the application logic) correctly report coverage for the `internal/` package.

### How it works
We use the `-coverpkg=./internal/...` flag during test execution. This tells the Go test runner to instrument and monitor all application logic in `internal/`, even if the tests are in a sibling package.

### Command
```bash
make test-coverage
```
- **Inputs**: Unit, Integration, and Go E2E tests.
- **Output**: `coverage/coverage.out` and `coverage/coverage.html`.

---

## 3. Playwright E2E Coverage (Containerized)

Collecting coverage from a full stack running in Docker is complex. We solve this by compiling the Go binary with coverage instrumentation and using a custom orchestrator.

### The Pipeline
1. **Instrumented Binary**: The `Dockerfile` builds the Go service with `-cover`.
2. **Orchestration**: The `cmd/testcontainers` tool starts the stack and sets `GOCOVERDIR` inside the container.
3. **Trigger**: When Playwright tests finish, the orchestrator gracefully shuts down the service.
4. **Flush**: The Go service flushes its binary coverage counters to the volume-mapped directory.
5. **Extraction**: The host orchestrator waits for the flush and then verifies the data.
6. **Reporting**: `make report-coverage` processes the binary data from the mapped directory.

### Command
```bash
make test-e2e-js-cover
```
- **Output**: `coverage/e2e-js/coverage.html`.

---

## 4. Viewing Reports

### Terminal View
Both coverage targets display a summary in your terminal immediately after the run. It filters out 100% covered files to help you focus on what's missing.

### Browser View
You can launch the full, line-by-line HTML report by passing `OPEN=1`:
```bash
make test-coverage OPEN=1
make test-e2e-js-cover OPEN=1
```

---

## 5. Interpreting Results

- **Blue/Green**: Code was executed during the test run.
- **Red**: Code was NOT executed.
- **Percentage**: Represents statement coverage.

### Best Practices
- **Local Dev**: Use `make test-coverage` for a quick health check of your logic.
- **Deep Dive**: Use `make test-e2e-js-cover` to see how your code behaves under real-world integration scenarios.
- **Goal**: Aim for high coverage in `internal/services/` and `internal/handlers/`.
