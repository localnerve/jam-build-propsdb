# Go Fiber Data Service Implementation Plan

## Overview

This plan outlines the creation of a Go Fiber-based data service that serves as a drop-in replacement for the existing Node.js Express data service. The new service will support all GORM-compatible databases through a configurable `DB_TYPE` environment variable.

### Key Requirements

- **API Compatibility**: Maintain identical REST API endpoints and response formats
- **Database Support**: Support all GORM-compatible databases (MySQL, PostgreSQL, SQLite, SQL Server, etc.) via `DB_TYPE` configuration
- **Authentication**: Integrate with Authorizer service using the `authorizer-go` SDK for admin and user role validation
- **Business Logic**: Migrate MariaDB stored procedure logic to Go code
- **Version Control**: Implement optimistic locking with version conflict detection
- **Containerization**: Docker support for easy deployment

## User Review Required

> [!NOTE]
> **Decisions Confirmed**:
> - **Project Location**: `/Users/agrant/projects/propsdb`
> - **Authorizer SDK**: Using `authorizer-go` from https://github.com/AuthorizerDev/Authorizer-go
> - **Database Support**: All GORM-compatible databases via `DB_TYPE` environment variable
> - **Containerization**: Dockerfile included for Docker deployment

> [!IMPORTANT]
> **Reorganization of Documentation**: All documentation files (TESTING.md, OBSERVABILITY.md, etc.) have been moved to the `docs/` directory. Swagger docs are in `docs/api/` and coverage output is in `coverage/`.

> [!CAUTION]
> **Project Move**: The project will be moved from `/Users/agrant/projects/propsdb` to `/Users/agrant/projects/propsdb`. After the move, you may need to re-open the workspace or update your IDE settings if it does not follow the directory rename.

> [!WARNING]
> **Breaking Change - Stored Procedures**: The MariaDB stored procedures will be replaced with Go code. This means:
> - The database schema will be simplified (no stored procedures needed)
> - All business logic will be in the application layer
> - All GORM-supported databases will use the same Go code path
> - This is more maintainable but changes the database deployment model

## Proposed Changes

### Project Structure

```
propsdb/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go              # Configuration management
│   ├── database/
│   │   ├── connection.go          # Database connection setup
│   │   └── migrations/            # SQL migrations for both DBs
│   ├── models/
│   │   ├── application.go         # Application document models
│   │   └── user.go                # User document models
│   ├── middleware/
│   │   └── auth.go                # Authorization middleware
│   ├── handlers/
│   │   ├── app_data.go            # Application data handlers
│   │   └── user_data.go           # User data handlers
│   ├── services/
│   │   ├── data_service.go        # Core business logic
│   │   └── auth_service.go        # Authorizer integration
│   └── utils/
│       ├── response.go            # Response formatting
│       └── version.go             # Version management
├── migrations/
│   ├── mysql/                     # MySQL/MariaDB migrations
│   ├── postgres/                  # PostgreSQL migrations
│   ├── sqlite/                    # SQLite migrations
│   └── sqlserver/                 # SQL Server migrations
├── tests/
│   ├── integration/               # Integration tests
│   └── unit/                      # Unit tests
├── go.mod
├── go.sum
├── .env.example
├── Dockerfile
├── .dockerignore
└── README.md
```

---

### Core Components

#### [NEW] [main.go](file:///Users/agrant/projects/propsdb/cmd/server/main.go)

Application entry point that:
- Loads configuration from environment variables
- Initializes database connection based on `DB_TYPE` (MySQL/MariaDB, PostgreSQL, SQLite, SQL Server, etc.)
- Sets up Fiber app with middleware (compression, JSON parsing, cookie parsing)
- **Implements API versioning via `X-Api-Version` header** (matching Node.js service):
  - Middleware to parse `X-Api-Version` header
  - Route versioning with version map (default → 1.0.0, "1.0" → 1.0.0)
  - Support for `useMaxVersion` behavior (routes to highest compatible version)
  - Currently only version 1.0.0 is implemented
- Registers API routes under `/api` mountpath
- Custom 404 handler with JSON response matching Node.js format
- Global error handler with special handling for `E_VERSION` errors (409 status)
- Implements graceful shutdown

> [!NOTE]
> **API Versioning**: Yes, this will use the same `X-Api-Version` header approach as the Node.js service. Clients can send `X-Api-Version: 1.0.0` or `X-Api-Version: 1.0` and the service will route to the appropriate version handler. If no header is provided, it defaults to version 1.0.0.

#### [NEW] [config.go](file:///Users/agrant/projects/propsdb/internal/config/config.go)

Configuration structure matching Node.js environment variables:
- `DB_TYPE` - Database type: "mysql", "postgres", "sqlite", "sqlserver", etc. (any GORM-supported database)
- `DB_HOST`, `DB_PORT`, `DB_APP_DATABASE`, `DB_APP_USER`, `DB_APP_PASSWORD`, `DB_APP_CONNECTION_LIMIT`
- `DB_USER`, `DB_PASSWORD`, `DB_CONNECTION_LIMIT`
- `AUTHZ_URL`, `AUTHZ_CLIENT_ID`
- `PORT` - Server port (default: 3000)

---

### Database Layer

#### [NEW] [connection.go](file:///Users/agrant/projects/propsdb/internal/database/connection.go)

Database connection management with dynamic driver selection based on `DB_TYPE`:
- MySQL/MariaDB support using `gorm.io/driver/mysql`
- PostgreSQL support using `gorm.io/driver/postgres`
- SQLite support using `gorm.io/driver/sqlite`
- SQL Server support using `gorm.io/driver/sqlserver`
- Extensible for other GORM drivers
- Connection pooling configuration
- Health check/ping functionality
- Graceful shutdown support

#### [NEW] [application.go](file:///Users/agrant/projects/propsdb/internal/models/application.go)

GORM models for application data:
```go
type ApplicationDocument struct {
    DocumentID      uint64    `gorm:"primaryKey;autoIncrement"`
    DocumentName    string    `gorm:"uniqueIndex;size:255;not null"`
    DocumentVersion uint64    `gorm:"not null;default:0"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    Collections     []ApplicationCollection `gorm:"many2many:application_documents_collections"`
}

type ApplicationCollection struct {
    CollectionID   uint64    `gorm:"primaryKey;autoIncrement"`
    CollectionName string    `gorm:"size:255;not null"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
    Properties     []ApplicationProperty `gorm:"many2many:application_collections_properties"`
}

type ApplicationProperty struct {
    PropertyID    uint64          `gorm:"primaryKey;autoIncrement"`
    PropertyName  string          `gorm:"size:255;not null"`
    PropertyValue datatypes.JSON  `gorm:"type:json"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

#### [NEW] [user.go](file:///Users/agrant/projects/propsdb/internal/models/user.go)

GORM models for user data (similar structure to application models but with `UserID` foreign key).

#### [NEW] [migrations/](file:///Users/agrant/projects/propsdb/migrations/)

SQL migration files for all supported databases:
- MySQL/MariaDB: Based on existing `mariadb-ddl-tables.sql` (without stored procedures)
- PostgreSQL: Equivalent schema adapted for PostgreSQL syntax
- SQLite: Simplified schema for SQLite
- SQL Server: Adapted schema for SQL Server syntax
- GORM AutoMigrate as fallback for other databases

---

### Middleware

#### [NEW] [version.go](file:///Users/agrant/projects/propsdb/internal/middleware/version.go)

API versioning middleware:
- Parse `X-Api-Version` header from request
- Store version in Fiber context: `c.Locals("apiVersion", version)`
- Default to "1.0.0" if header not present
- Support version aliases ("1.0" → "1.0.0")
- Version routing logic matching `express-version-route` behavior

---

### Authentication & Authorization

#### [NEW] [auth.go](file:///Users/agrant/projects/propsdb/internal/middleware/auth.go)

Authorization middleware implementing:
- `AuthAdmin()` - Validates admin role, sets `c.Locals("user", userData)`
- `AuthUser()` - Validates user role, sets `c.Locals("user", userData)`
- Cookie parsing for `cookie_session`
- Integration with Authorizer service
- Error handling with 403 responses

#### [NEW] [auth_service.go](file:///Users/agrant/projects/propsdb/internal/services/auth_service.go)

Authorizer service client using `authorizer-go` SDK:
- Initialize Authorizer client from https://github.com/AuthorizerDev/Authorizer-go
- `ValidateSession(cookie, roles)` method using SDK's session validation
- Ping/health check for Authorizer service
- Singleton pattern for Authorizer client instance

---

### API Handlers

#### [NEW] [app_data.go](file:///Users/agrant/projects/propsdb/internal/handlers/app_data.go)

Handlers for application data endpoints:
- `GetAppProperties(c *fiber.Ctx)` - GET `/api/data/app/:document/:collection`
- `GetAppCollectionsAndProperties(c *fiber.Ctx)` - GET `/api/data/app/:document?collections=...`
- `GetAppDocumentsCollectionsAndProperties(c *fiber.Ctx)` - GET `/api/data/app`
- `SetAppProperties(c *fiber.Ctx)` - POST `/api/data/app/:document` (admin only)
- `DeleteAppCollection(c *fiber.Ctx)` - DELETE `/api/data/app/:document/:collection` (admin only)
- `DeleteAppProperties(c *fiber.Ctx)` - DELETE `/api/data/app/:document` (admin only)

#### [NEW] [user_data.go](file:///Users/agrant/projects/propsdb/internal/handlers/user_data.go)

Handlers for user data endpoints (similar to app_data.go but with user context):
- All routes require user authentication
- User ID extracted from `c.Locals("user")`

---

### Business Logic

#### [NEW] [data_service.go](file:///Users/agrant/projects/propsdb/internal/services/data_service.go)

Core business logic migrated from stored procedures:

**Get Operations:**
- `GetProperties(db, document, collection, userID)` - Returns properties for a document/collection
- `GetCollectionsAndProperties(db, document, collections, userID)` - Returns multiple collections
- `GetDocumentsCollectionsAndProperties(db, userID)` - Returns all documents
- Result reduction to match Node.js output format: `{ document: { collection: { propName: propVal }}}`

**Upsert Operations:**
- `SetProperties(db, document, version, collections, userID)` - Upserts document with collections/properties
- Optimistic locking: Check version before update, return `E_VERSION` error on conflict
- Transaction management with rollback on error
- Returns new version number on success

**Delete Operations:**
- `DeleteCollection(db, document, version, collection, userID)` - Deletes a collection
- `DeleteProperties(db, document, version, collections, userID)` - Deletes properties or full document
- Cleanup of orphaned collections and properties
- Version increment on successful delete

**Version Control:**
- All mutations check current version matches input version
- Use `FOR UPDATE` row locking in transactions
- Return version conflict errors with `E_VERSION` message

---

### Utilities

#### [NEW] [response.go](file:///Users/agrant/projects/propsdb/internal/utils/response.go)

Response formatting utilities:
- `SuccessResponse(data, status)` - Standard success response
- `ErrorResponse(message, status, errorType)` - Standard error response matching Node.js format
- `VersionErrorResponse()` - Specific E_VERSION error format
- `NotFoundResponse(message)` - 404 response

#### [NEW] [version.go](file:///Users/agrant/projects/propsdb/internal/utils/version.go)

Version management utilities:
- Version conflict detection
- Version increment logic

---

### Configuration Files

#### [NEW] [go.mod](file:///Users/agrant/projects/propsdb/go.mod)

Dependencies:
- `github.com/gofiber/fiber/v2` - Web framework
- `gorm.io/gorm` - ORM
- `gorm.io/driver/mysql` - MySQL/MariaDB driver
- `gorm.io/driver/postgres` - PostgreSQL driver
- `gorm.io/driver/sqlite` - SQLite driver
- `gorm.io/driver/sqlserver` - SQL Server driver
- `gorm.io/datatypes` - JSON support
- `github.com/authorizerdev/authorizer-go` - Authorizer SDK
- Testing libraries (testcontainers, etc.)

#### [NEW] [.env.example](file:///Users/agrant/projects/propsdb/.env.example)

Example environment configuration with all required variables.

#### [NEW] [Dockerfile](file:///Users/agrant/projects/propsdb/Dockerfile)

Multi-stage Docker build:
- Stage 1: Build Go binary with all dependencies
- Stage 2: Minimal runtime image (alpine or distroless)
- Expose port 3000
- Health check endpoint
- Non-root user for security

#### [NEW] [.dockerignore](file:///Users/agrant/projects/propsdb/.dockerignore)

Exclude unnecessary files from Docker context:
- `.git/`, `tests/`, `.env`, etc.

#### [NEW] [README.md](file:///Users/agrant/projects/propsdb/README.md)

Documentation covering:
- Setup instructions
- Environment configuration (including `DB_TYPE` for all GORM databases)
- Running with different databases (MySQL, PostgreSQL, SQLite, SQL Server)
- Docker deployment instructions
- API endpoints and usage
- Testing instructions

---

## Verification Plan

### Automated Tests

1. **Unit Tests**
   - Test each handler function with mocked database
   - Test business logic functions independently
   - Test version conflict detection
   - Test input validation and transformation

2. **Integration Tests with Testcontainers**
   - Spin up MariaDB container and run full test suite
   - Spin up PostgreSQL container and run full test suite
   - Spin up SQL Server container and run full test suite
   - Test SQLite with in-memory database
   - Test all CRUD operations on each database
   - Test concurrent version conflicts
   - Test authorization middleware
   - Verify response formats match Node.js service

3. **Commands to run:**
   ```bash
   go test ./... -v
   go test ./tests/integration -v -tags=integration
   ```

### Manual Verification

1. **Database Compatibility**
   - Deploy schema to MariaDB, PostgreSQL, SQLite, and SQL Server
   - Verify migrations run successfully on all databases
   - Test data operations on all supported databases
   - Verify GORM AutoMigrate works for unsupported databases

2. **API Compatibility**
   - Compare API responses with Node.js service
   - Verify error messages and status codes match
   - Test version header support

3. **Performance Testing**
   - Load test with concurrent requests
   - Verify connection pooling works correctly
   - Test graceful shutdown under load
