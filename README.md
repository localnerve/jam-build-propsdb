# jam-build-propsdb - Go Fiber Data Service

A high-performance data service built with Go and Fiber, serving as a drop-in replacement for the Node.js Express data service in the [Jam Build](https://github.com/localnerve/jam-build) project. Supports all GORM-compatible databases including MariaDB/MySQL, PostgreSQL, SQLite, and SQL Server.

## Features

- 🚀 **High Performance**: Built with Go Fiber for maximum throughput
- 🗄️ **Multi-Database Support**: Works with MariaDB, MySQL, PostgreSQL, SQLite, and SQL Server via `DB_TYPE` configuration. [Configuration details](docs/DATABASE.md)
- 🔐 **Authentication**: Integrated with Authorizer using `authorizer-go` SDK
- 📦 **API Compatibility**: Drop-in replacement for the Node.js Express service. [API details](#api-endpoints)
- 🔄 **Version Control**: Optimistic locking with `E_VERSION` conflict detection
- 🐳 **Docker Ready**: Multi-stage Dockerfile for containerized deployments
- 🧪 **Testable**: Designed for unit, integration, and e2e testing with testcontainers. [Testing details](docs/TESTING.md)

## Quick Start

### Public Docker Image

  * https://hub.docker.com/r/localnerve/jam-build-propsdb
    - `localnerve/jam-build-propsdb:latest`

  * Environment:
    - PORT: The exposed port to the propsdb-api
    - DB_TYPE: The database type [mariadb | mysql | mssql | postgres | sqlite]
    - DB_HOST: The hostname of the database service
    - DB_PORT: The database service port
    - DB_APP_DATABASE: The property database name
    - DB_APP_USER: The application user name
    - DB_APP_PASSWORD: The application user password
    - DB_USER: The user user name
    - DB_PASSWORD: The user user password
    - DB_APP_CONNECTION_LIMIT: The application connection pool limit
    - DB_CONNECTION_LIMIT: The user connection pool limit
    - AUTHZ_URL: The url to the authorizer service
    - AUTHZ_CLIENT_ID: The client ID of the authorizer service

### Development

#### Prerequisites

- Go 1.21 or higher
- Docker Desktop
- Node 24.12.0 or higher
- GNU Make 4.4.1

#### Installation

1. Clone the repository:
```bash
git clone https://github.com/localnerve/jam-build-propsdb.git
cd jam-build-propsdb
```

2. Install dependencies:
```bash
make deps
make install-tools
```

3. Build and run the full service:
```bash
make docker-compose-up
```

The service will start on `http://localhost:3000`.

#### Ports
These are the ports used by default:

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

# Full stack end-to-end tests (requires Docker)
make test-e2e-js # Params: DEBUG=1 (debug, no rebuild), DEBUG=2 (debug, full rebuild)

# Full stack end-to-end tests with coverage (requires Docker)
make test-e2e-js-cover # Params: REBUILD=1 (rebuild orchestrator), HOST_DEBUG=1 (debug host)

```

Many more tests are available in the Makefile, see the [testing documentation](docs/TESTING.md) for full details.

### Building

```bash
# Build the api service binary
make build

# Build the healthcheck binary
make build-healthcheck

# Build the testcontainers binary
make build-testcontainers

# Build the docker image
make docker-build

# Build and run the entire composition
make DB_TYPE=mariadb docker-compose-up # DB_TYPE = mariadb | mysql | mssql | postgres | sqlite

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