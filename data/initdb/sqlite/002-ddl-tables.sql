--
-- Jam-build database structure for SQLite.
--

-- Create the application_documents table
CREATE TABLE IF NOT EXISTS application_documents (
    document_id INTEGER PRIMARY KEY AUTOINCREMENT,
    document_name TEXT NOT NULL UNIQUE,
    document_version INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create the application_collections table
CREATE TABLE IF NOT EXISTS application_collections (
    collection_id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create the application_properties table
CREATE TABLE IF NOT EXISTS application_properties (
    property_id INTEGER PRIMARY KEY AUTOINCREMENT,
    property_name TEXT NOT NULL,
    property_value TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (json_valid(property_value))
);

-- Junction tables
CREATE TABLE IF NOT EXISTS application_documents_collections (
    document_id INTEGER NOT NULL,
    collection_id INTEGER NOT NULL,
    PRIMARY KEY (document_id, collection_id),
    FOREIGN KEY (document_id) REFERENCES application_documents(document_id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES application_collections(collection_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS application_collections_properties (
    collection_id INTEGER NOT NULL,
    property_id INTEGER NOT NULL,
    PRIMARY KEY (collection_id, property_id),
    FOREIGN KEY (collection_id) REFERENCES application_collections(collection_id) ON DELETE CASCADE,
    FOREIGN KEY (property_id) REFERENCES application_properties(property_id) ON DELETE CASCADE
);

-- User data
CREATE TABLE IF NOT EXISTS user_documents (
    document_id INTEGER UNIQUE,
    user_id TEXT NOT NULL,
    document_name TEXT NOT NULL,
    document_version INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, document_name)
);

CREATE TABLE IF NOT EXISTS user_collections (
    collection_id INTEGER PRIMARY KEY AUTOINCREMENT,
    collection_name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_properties (
    property_id INTEGER PRIMARY KEY AUTOINCREMENT,
    property_name TEXT NOT NULL,
    property_value TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (json_valid(property_value))
);

CREATE TABLE IF NOT EXISTS user_documents_collections (
    document_id INTEGER NOT NULL,
    collection_id INTEGER NOT NULL,
    PRIMARY KEY (document_id, collection_id),
    FOREIGN KEY (document_id) REFERENCES user_documents(document_id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES user_collections(collection_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_collections_properties (
    collection_id INTEGER NOT NULL,
    property_id INTEGER NOT NULL,
    PRIMARY KEY (collection_id, property_id),
    FOREIGN KEY (collection_id) REFERENCES user_collections(collection_id) ON DELETE CASCADE,
    FOREIGN KEY (property_id) REFERENCES user_properties(property_id) ON DELETE CASCADE
);
