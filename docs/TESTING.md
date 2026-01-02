# PropsDB - Testing Guide

## Test Issues and Solutions

### GORM Junction Table Column Names

**Issue**: GORM AutoMigrate creates junction table columns with full model name prefixes, which differ from the MariaDB migration schema.

**GORM Column Names**:
- `application_document_document_id` (not `document_id`)
- `application_collection_collection_id` (not `collection_id`)
- `application_property_property_id` (not `property_id`)

**Solution**: All JOIN queries have been updated to use GORM's actual column names for compatibility with both SQLite (unit tests) and production databases.

### Running Tests

#### Unit Tests
```bash
# Run all unit tests (fast, uses SQLite in-memory)
make test-unit

# Run specific test
go test -v ./tests/unit -run TestGetAppProperties
```

**Status**: ✅ All unit tests passing (4/4)

#### Integration Tests
```bash
# Run all integration tests (requires Docker)
make test-integration

# Run specific database test
go test -v ./tests/integration -run TestWithMariaDB
go test -v ./tests/integration -run TestWithPostgreSQL
```

**Requirements**:
- Docker running
- Sufficient disk space for container images
- Network access to pull images

#### Code Coverage
```bash
# Generate coverage report
make test-coverage

# View in browser (opens automatically)
# Or manually: open coverage.html

# Terminal view
go tool cover -func=coverage.out
```

### Test Structure

```
tests/
├── unit/
│   └── handlers_test.go       # Handler tests with SQLite
└── integration/
    └── integration_test.go    # Real database tests with testcontainers
```

### Common Issues

#### 1. Lint Command Not Found

**Error**: `make: golangci-lint: No such file or directory`

**Solution**: The Makefile now correctly uses `$GOPATH/bin/golangci-lint` with fallback to system path.

#### 2. Column Name Mismatches

**Error**: `no such column: dc.document_id`

**Solution**: Fixed in all JOIN queries to use GORM's prefixed column names.

#### 3. Docker Not Running

**Error**: Integration tests fail to start containers

**Solution**:
```bash
# Start Docker
open -a Docker  # macOS

# Verify Docker is running
docker ps
```

### Best Practices

1. **Run unit tests frequently** - Fast feedback loop
2. **Run integration tests before commits** - Catch database-specific issues
3. **Generate coverage reports** - Aim for >80% coverage
4. **Use `-short` flag** - Skip slow tests during development

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Unit Tests
  run: make test-unit

- name: Integration Tests
  run: make test-integration
  # Requires Docker service
```

### Debugging Tests

```bash
# Verbose output
go test -v ./tests/unit/...

# Run with race detector
go test -race ./tests/unit/...

# Show test coverage per function
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
