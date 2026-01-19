#!/bin/bash
set -e

# MSSQL Initialization Script
# This script waits for SQL Server to start and then executes initialization SQL.

echo "Waiting for SQL Server to start..."
for i in {1..100}; do
    /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -Q "SELECT 1" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "SQL Server is ready."
        break
    fi
    echo "SQL Server is not ready yet... ($i/100)"
    sleep 2
done

echo "Starting initialization for MSSQL..."

# 1. Create databases and logins (Server Level)
/opt/mssql-tools18/bin/sqlcmd -b -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -Q "
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = '$DB_APP_DATABASE') CREATE DATABASE [$DB_APP_DATABASE];
IF NOT EXISTS (SELECT * FROM sys.databases WHERE name = '$AUTHZ_DATABASE') CREATE DATABASE [$AUTHZ_DATABASE];
IF NOT EXISTS (SELECT * FROM sys.server_principals WHERE name = '$DB_APP_USER') CREATE LOGIN [$DB_APP_USER] WITH PASSWORD = '$DB_APP_PASSWORD';
IF NOT EXISTS (SELECT * FROM sys.server_principals WHERE name = '$DB_USER') CREATE LOGIN [$DB_USER] WITH PASSWORD = '$DB_PASSWORD';
"

# 2. Create tables in application database
echo "Creating tables in $DB_APP_DATABASE..."
/opt/mssql-tools18/bin/sqlcmd -b -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -d "$DB_APP_DATABASE" -Q "
-- Authorizer shadow table
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'authorizer_users')
BEGIN
    CREATE TABLE authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);
END
"

# 3. Create tables in authorizer database
echo "Creating tables in $AUTHZ_DATABASE..."
/opt/mssql-tools18/bin/sqlcmd -b -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -d "$AUTHZ_DATABASE" -Q "
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'authorizer_users')
BEGIN
    CREATE TABLE authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);
END
"

# 4. Grant privileges (Database Level)
echo "Granting privileges..."
/opt/mssql-tools18/bin/sqlcmd -b -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -d "$DB_APP_DATABASE" -Q "
IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = '$DB_APP_USER') CREATE USER [$DB_APP_USER] FOR LOGIN [$DB_APP_USER];
IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = '$DB_USER') CREATE USER [$DB_USER] FOR LOGIN [$DB_USER];
ALTER ROLE [db_owner] ADD MEMBER [$DB_APP_USER];
ALTER ROLE [db_datareader] ADD MEMBER [$DB_USER];
ALTER ROLE [db_datawriter] ADD MEMBER [$DB_USER];
"

/opt/mssql-tools18/bin/sqlcmd -b -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -C -d "$AUTHZ_DATABASE" -Q "
IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = '$DB_APP_USER') CREATE USER [$DB_APP_USER] FOR LOGIN [$DB_APP_USER];
IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = '$DB_USER') CREATE USER [$DB_USER] FOR LOGIN [$DB_USER];
ALTER ROLE [db_owner] ADD MEMBER [$DB_APP_USER];
ALTER ROLE [db_datareader] ADD MEMBER [$DB_USER];
"

echo "MSSQL initialization complete."
