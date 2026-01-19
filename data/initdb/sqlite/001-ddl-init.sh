#!/bin/sh
set -e

# SQLite Initialization Script
# This script ensures the database files exist and have the correct schema.

DB_PATH="/data/${DB_APP_DATABASE}.db"
AUTHZ_DB_PATH="/data/${AUTHZ_DATABASE}.db"

echo "Initializing SQLite databases..."

# Initialize PropsDB database
# We now rely on GORM AutoMigrate to create the schema to avoid parsing bugs with pre-existing constraints
if [ ! -f "$DB_PATH" ]; then
    echo "Creating empty PropsDB database file at $DB_PATH..."
    touch "$DB_PATH"
fi

# Initialize Authorizer database
if [ ! -f "$AUTHZ_DB_PATH" ]; then
    echo "Creating Authorizer database at $AUTHZ_DB_PATH..."
    sqlite3 "$AUTHZ_DB_PATH" "CREATE TABLE authorizer_users (id CHAR(36) NOT NULL PRIMARY KEY);"
fi

# Relax permissions so non-root containers (api, authorizer) can write
chmod 666 "$DB_PATH" "$AUTHZ_DB_PATH"
chmod 777 /data

echo "SQLite initialization complete."
