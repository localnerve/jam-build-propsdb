#!/bin/sh
set -e

echo "Running 002-ddl-tables.sh as $DB_APP_USER in $POSTGRES_DB..."

# Execute SQL tables creation as the application user to ensure correct ownership
psql -v ON_ERROR_STOP=1 --username "$DB_APP_USER" --dbname "$POSTGRES_DB" <<-EOSQL
-- Create the application_documents table
CREATE TABLE IF NOT EXISTS application_documents (
    document_id SERIAL PRIMARY KEY,
    document_name VARCHAR(255) NOT NULL UNIQUE,
    document_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the application_collections table
CREATE TABLE IF NOT EXISTS application_collections (
    collection_id SERIAL PRIMARY KEY,
    collection_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create the application_properties table
CREATE TABLE IF NOT EXISTS application_properties (
    property_id SERIAL PRIMARY KEY,
    property_name VARCHAR(255) NOT NULL,
    property_value JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Junction tables
CREATE TABLE IF NOT EXISTS application_documents_collections (
    document_id INTEGER NOT NULL REFERENCES application_documents(document_id) ON DELETE CASCADE,
    collection_id INTEGER NOT NULL REFERENCES application_collections(collection_id) ON DELETE CASCADE,
    PRIMARY KEY (document_id, collection_id)
);

CREATE TABLE IF NOT EXISTS application_collections_properties (
    collection_id INTEGER NOT NULL REFERENCES application_collections(collection_id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES application_properties(property_id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, property_id)
);

-- User data
CREATE TABLE IF NOT EXISTS user_documents (
    document_id SERIAL UNIQUE,
    user_id CHAR(36) NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    document_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, document_name),
    FOREIGN KEY (user_id) REFERENCES authorizer_users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_collections (
    collection_id SERIAL PRIMARY KEY,
    collection_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_properties (
    property_id SERIAL PRIMARY KEY,
    property_name VARCHAR(255) NOT NULL,
    property_value JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_documents_collections (
    document_id INTEGER NOT NULL REFERENCES user_documents(document_id) ON DELETE CASCADE,
    collection_id INTEGER NOT NULL REFERENCES user_collections(collection_id) ON DELETE CASCADE,
    PRIMARY KEY (document_id, collection_id)
);

CREATE TABLE IF NOT EXISTS user_collections_properties (
    collection_id INTEGER NOT NULL REFERENCES user_collections(collection_id) ON DELETE CASCADE,
    property_id INTEGER NOT NULL REFERENCES user_properties(property_id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, property_id)
);
EOSQL

echo "002-ddl-tables.sh complete."
