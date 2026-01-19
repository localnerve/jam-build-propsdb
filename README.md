# jam-build-propsdb - Go Fiber Data Service

A high-performance data service built with Go and Fiber, serving as a drop-in replacement for the Node.js Express data service in the [Jam Build](https://github.com/localnerve/jam-build) project. Supports all GORM-compatible databases including MariaDB/MySQL, PostgreSQL, SQLite, and SQL Server.

## Features

- 🚀 **High Performance**: Built with Go Fiber for maximum throughput
- 🗄️ **Multi-Database Support**: Works with MariaDB, MySQL, PostgreSQL, SQLite, and SQL Server via `DB_TYPE` configuration. See [this document](docs/DATABASE.md) for configuration details.
- 🔐 **Authentication**: Integrated with Authorizer using `authorizer-go` SDK
- 📦 **API Compatibility**: Drop-in replacement for the Node.js Express service
- 🔄 **Version Control**: Optimistic locking with `E_VERSION` conflict detection
- 🐳 **Docker Ready**: Multi-stage Dockerfile for containerized deployments
- 🧪 **Testable**: Designed for unit, integration, and e2e testing with testcontainers

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Node 24.12.0 or higher
- GNU Make 4.4.1
- Docker Desktop

### Installation

1. Clone the repository:
```bash
git clone https://github.com/localnerve/jam-build-propsdb.git
cd jam-build-propsdb
```

2. Install dependencies:
```bash
make deps
```

3. Build and run the service suite:
```bash
make docker-compose-up
```

The service will start on `http://localhost:3000`.

#### All Ports Used

* Service ports: 3000 (api), 6379 (cache), 8080 (authorizer)
* Typical database ports (depends on `DB_TYPE`, defaults to `mariadb`):
    * 3306 (mariadb, mysql)
    * 5432 (postgres)
    * 1433 (mssql)
* Monitoring ports: 3001 (grafana), 9090 (prometheus)

## Database Configuration

### Supported Databases

Configure the database type using the `DB_TYPE` environment variable (defaults to `mariadb`). For a detailed matrix of configuration requirements (ports, images, healthchecks), see [the database documentation](docs/DATABASE.md).

- **MySQL/MariaDB**: `DB_TYPE=mysql` or `DB_TYPE=mariadb`
- **PostgreSQL**: `DB_TYPE=postgres` or `DB_TYPE=postgresql`
- **SQLite**: `DB_TYPE=sqlite` (set `DB_APP_DATABASE` to file path)
- **SQL Server**: `DB_TYPE=sqlserver` or `DB_TYPE=mssql`

### Initialization

SQL migration files will be applied for each database type in the `data/initdb/` directory:

- `data/initdb/mariadb/` - MariaDB initialization
- `data/initdb/mysql/` - MySQL initialization
- `data/initdb/postgres/` - PostgreSQL initialization
- `data/initdb/sqlite/` - SQLite initialization
- `data/initdb/sqlserver/` - SQL Server initialization

### Migrations

SQL migration files will be applied for each database type in the `data/migrations/` directory:

- `data/migrations/mariadb/` - MariaDB migrations
- `data/migrations/mysql/` - MySQL migrations
- `data/migrations/postgres/` - PostgreSQL migrations
- `data/migrations/sqlite/` - SQLite migrations
- `data/migrations/sqlserver/` - SQL Server migrations

The service also supports GORM AutoMigrate as a fallback.

## API Endpoints

Swagger documentation is available at `http://localhost:3000/swagger/` and updated with `make swagger`.

### Application Data (Public GET, Admin POST/DELETE)

- `GET /api/data/app/:document/:collection` - Get properties for a document/collection
- `GET /api/data/app/:document?collections=col1,col2` - Get collections and properties
- `GET /api/data/app` - Get all documents, collections, and properties
- `POST /api/data/app/:document` - Upsert document (requires admin role)
- `DELETE /api/data/app/:document/:collection` - Delete collection (requires admin role)
- `DELETE /api/data/app/:document` - Delete document or properties (requires admin role)

### User Data (All require user authentication)

- `GET /api/data/user/:document/:collection` - Get user properties
- `GET /api/data/user/:document?collections=col1,col2` - Get user collections
- `GET /api/data/user` - Get all user documents
- `POST /api/data/user/:document` - Upsert user document
- `DELETE /api/data/user/:document/:collection` - Delete user collection
- `DELETE /api/data/user/:document` - Delete user document or properties

### API Versioning

The service supports API versioning via the `X-Api-Version` header:

```bash
curl -H "X-Api-Version: 1.0.0" http://localhost:3000/api/data/app
```

Supported versions:
- `1.0.0` (default)
- `1.0` (alias for 1.0.0)

## Docker Deployment

### Service Stack Docker Compose

This will start the API service, database, cache, and authorizer.

```bash
make docker-compose-up
```

### Observability Stack Docker Compose

This will start the observability services for the API service.

```bash
make obs-up
```

For information on Prometheus metrics and Grafana dashboards, see [OBSERVABILITY](docs/OBSERVABILITY.md).

## Development

### Running Tests

Full information on running tests is available in the [testing documentation](docs/TESTING.md).

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# End-to-end service health tests (requires Docker)
make test-e2e

# Unit, Integration, and End-to-end tests with coverage (requires Docker)
make test-coverage

# Full stack end-to-end tests (requires Docker)
make test-e2e-js # Params: DEBUG=1 (debug, no rebuild), DEBUG=2 (debug, full rebuild)

# Full stack end-to-end tests with coverage (requires Docker)
make test-e2e-js-cover # Params: REBUILD=1 (rebuild orchestrator), HOST_DEBUG=1 (debug host)

```

Many more tests are available in the Makefile, see the [testing documentation](docs/TESTING.md) for full details.

### Building

```bash

# build the api service
make build 

# build the healthcheck binary
make build-healthcheck

# build the testcontainers binary
make build-testcontainers

```

## Authentication

The service integrates with [Authorizer](https://authorizer.dev/) for authentication and authorization.

- **Admin routes**: Require `admin` role
- **User routes**: Require `user` role
- **Session cookie**: `cookie_session`

## Version Control

All mutation operations (POST, DELETE) use optimistic locking:

1. Client sends current `version` in request body
2. Server checks if version matches current document version
3. If mismatch, returns `409 Conflict` with `E_VERSION` error
4. Client must refresh, reconcile, and retry

## License

Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC

This project is licensed under the GNU Affero General Public License v3.0 or later.

## Acknowledgments

Special thanks to the Antigravity AI assistant for help with the Go migration, testing architecture, and documentation.