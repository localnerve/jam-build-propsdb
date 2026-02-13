# PropsDB - Database Configuration Matrix

PropsDB is designed to be database-agnostic. However, since different database engines have different requirements for Docker initialization, this document serves as the source of truth for the variables and configurations required for each type.

## Configuration Requirements Matrix

| DB_TYPE | Image | Default Port | Internal Data Path | Primary Env Vars | Healthcheck Command |
|---------|-------|--------------|--------------------|------------------|---------------------|
| **mariadb** | `mariadb:12.2.2` | 3306 | `/var/lib/mysql` | `MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD` | `healthcheck.sh --connect --innodb_initialized` |
| **mysql** | `mysql:8.4` | 3306 | `/var/lib/mysql` | `MYSQL_ROOT_PASSWORD`, `MYSQL_DATABASE`, `MYSQL_USER`, `MYSQL_PASSWORD` | `mysqladmin ping -h localhost -u root -p${MYSQL_ROOT_PASSWORD}` |
| **postgres** | `postgres:18-alpine` | 5432 | `/var/lib/postgresql/data` | `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_USER` | `pg_isready -U ${POSTGRES_USER} -d ${POSTGRES_DB}` |
| **mssql** | `mcr.microsoft.com/mssql/server:2022-latest` | 1433 | `/var/opt/mssql` | `ACCEPT_EULA=Y`, `MSSQL_SA_PASSWORD` | `/opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P ${MSSQL_SA_PASSWORD} -Q "SELECT 1"` |
| **sqlite** | `keinos/sqlite3` | N/A | `/data` | N/A | `test -f /data/${DB_APP_DATABASE}.db` |

## Implementation Challenges

When switching between these types in a single `docker-compose.yml`, several issues arise (environment names, healthchecks, internal paths).

## Solution: Specialized Fragments

We use a "base" `docker-compose.yml` for core services and specialized "fragments" in `data/compose/` for each database type. The `Makefile` automatically selects the correct fragment based on `DB_TYPE`.

Available Fragments:
- `data/compose/mariadb.yml`
- `data/compose/mysql.yml`
- `data/compose/postgres.yml`
- `data/compose/mssql.yml`

This allows adding support for new database types by simply creating a new YAML fragment and matching `data/initdb/` scripts.
