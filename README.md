# propsdb - Go Fiber Data Service

A high-performance data service built with Go and Fiber, serving as a drop-in replacement for the Node.js Express data service. Supports all GORM-compatible databases including MySQL, PostgreSQL, SQLite, and SQL Server.

## Features

- 🚀 **High Performance**: Built with Go Fiber for maximum throughput
- 🗄️ **Multi-Database Support**: Works with MySQL, PostgreSQL, SQLite, SQL Server, and more via `DB_TYPE` configuration
- 🔐 **Authentication**: Integrated with Authorizer using `authorizer-go` SDK
- 📦 **API Compatibility**: Drop-in replacement for the Node.js Express service
- 🔄 **Version Control**: Optimistic locking with `E_VERSION` conflict detection
- 🐳 **Docker Ready**: Multi-stage Dockerfile for containerized deployments
- 🧪 **Testable**: Designed for integration testing with testcontainers

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Database (MySQL, PostgreSQL, SQLite, or SQL Server)
- Authorizer instance running

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd propsdb-claude
```

2. Copy the example environment file:
```bash
cp .env.example .env
```

3. Configure your environment variables in `.env`:
```env
PORT=3000
DB_TYPE=mysql  # or postgres, sqlite, sqlserver
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=jam_build
DB_APP_USER=jbadmin
DB_APP_PASSWORD=your_password
DB_USER=jbuser
DB_PASSWORD=your_password
AUTHZ_URL=http://localhost:8080
AUTHZ_CLIENT_ID=your_client_id
```

4. Install dependencies:
```bash
go mod download
```

5. Run the service:
```bash
go run cmd/server/main.go
```

The service will start on `http://localhost:3000`.

## Database Configuration

### Supported Databases

Configure the database type using the `DB_TYPE` environment variable:

- **MySQL/MariaDB**: `DB_TYPE=mysql` or `DB_TYPE=mariadb`
- **PostgreSQL**: `DB_TYPE=postgres` or `DB_TYPE=postgresql`
- **SQLite**: `DB_TYPE=sqlite` (set `DB_DATABASE` to file path)
- **SQL Server**: `DB_TYPE=sqlserver` or `DB_TYPE=mssql`

### Migrations

SQL migration files are provided for each database type in the `migrations/` directory:

- `migrations/mysql/` - MySQL/MariaDB migrations
- `migrations/postgres/` - PostgreSQL migrations
- `migrations/sqlite/` - SQLite migrations
- `migrations/sqlserver/` - SQL Server migrations

The service also supports GORM AutoMigrate as a fallback.

## API Endpoints

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

### Build Docker Image

```bash
docker build -t propsdb:latest .
```

### Run with Docker

```bash
docker run -d \
  --name propsdb \
  -p 3000:3000 \
  -e DB_TYPE=mysql \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=3306 \
  -e DB_DATABASE=jam_build \
  -e DB_APP_USER=jbadmin \
  -e DB_APP_PASSWORD=password \
  -e DB_USER=jbuser \
  -e DB_PASSWORD=password \
  -e AUTHZ_URL=http://host.docker.internal:8080 \
  -e AUTHZ_CLIENT_ID=your_client_id \
  propsdb:latest
```

### Docker Compose (Example)

```yaml
version: '3.8'

services:
  propsdb:
    build: .
    ports:
      - "3000:3000"
    environment:
      - DB_TYPE=mysql
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_DATABASE=jam_build
      - DB_APP_USER=jbadmin
      - DB_APP_PASSWORD=password
      - DB_USER=jbuser
      - DB_PASSWORD=password
      - AUTHZ_URL=http://authorizer:8080
      - AUTHZ_CLIENT_ID=your_client_id
    depends_on:
      - mysql
      - authorizer

  mysql:
    image: mariadb:latest
    environment:
      - MYSQL_ROOT_PASSWORD=rootpassword
      - MYSQL_DATABASE=jam_build
    volumes:
      - mysql_data:/var/lib/mysql

  authorizer:
    image: lakhansamani/authorizer:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_TYPE=sqlite
      - DATABASE_URL=authorizer.db

volumes:
  mysql_data:
```

## Development

### Project Structure

```
propsdb-claude/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── models/          # GORM models
│   ├── middleware/      # HTTP middleware
│   ├── handlers/        # HTTP handlers
│   ├── services/        # Business logic
│   └── utils/           # Utilities
├── migrations/          # SQL migrations
├── tests/               # Tests
└── Dockerfile           # Docker configuration
```

### Running Tests

```bash
# Unit tests
go test ./... -v

# Integration tests (requires Docker)
go test ./tests/integration -v -tags=integration
```

### Building

```bash
go build -o propsdb ./cmd/server
./propsdb
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

Copyright (c) 2025 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC

This project is licensed under the GNU Affero General Public License v3.0 or later.

## Migration from Node.js Service

This service is a drop-in replacement for the Node.js Express data service with the following improvements:

- **Performance**: Go's compiled nature provides better performance
- **Database Support**: Expanded from MySQL/PostgreSQL to all GORM-supported databases
- **Simplified Deployment**: No stored procedures needed - all logic in application code
- **Type Safety**: Strong typing with Go
- **Containerization**: Production-ready Docker support

API endpoints, request/response formats, and authentication remain identical to ensure compatibility.
