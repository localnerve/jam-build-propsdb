#!/bin/sh
set -e

# SQLite Initialization Script
# This script ensures the database files exist and have the correct schema.

DB_PATH="/data/${DB_APP_DATABASE}.db"
AUTHZ_DB_PATH="/data/${AUTHZ_DATABASE}.db"

echo "Initializing SQLite databases..."

# Initialize PropsDB database
if [ ! -f "$DB_PATH" ]; then
    echo "Creating PropsDB database at $DB_PATH..."
    sqlite3 "$DB_PATH" < /scripts/init/002-ddl-tables.sql
fi

# Initialize Authorizer database
if [ ! -f "$AUTHZ_DB_PATH" ]; then
    echo "Creating Authorizer database at $AUTHZ_DB_PATH..."
    sqlite3 "$AUTHZ_DB_PATH" "CREATE TABLE authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);"
fi

echo "SQLite initialization complete."
