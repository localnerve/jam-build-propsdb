-- Jam-build database structure for MariaDB/MySQL
-- Migration file for propsdb Go Fiber service
--
-- This creates the schema without stored procedures
-- Business logic is implemented in the Go application layer
--
-- Jam-build, a web application practical reference.
-- Copyright (c) 2025 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS jam_build;
USE jam_build;

-- Application Documents Table
CREATE TABLE IF NOT EXISTS application_documents (
    document_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    document_name VARCHAR(255) NOT NULL UNIQUE,
    document_version BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_document_name (document_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Application Collections Table
CREATE TABLE IF NOT EXISTS application_collections (
    collection_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    collection_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_collection_name (collection_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Application Properties Table
CREATE TABLE IF NOT EXISTS application_properties (
    property_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    property_name VARCHAR(255) NOT NULL,
    property_value JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CHECK (JSON_VALID(property_value)),
    INDEX idx_property_name (property_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Application Documents-Collections Junction Table
CREATE TABLE IF NOT EXISTS application_documents_collections (
    document_id BIGINT UNSIGNED NOT NULL,
    collection_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (document_id, collection_id),
    FOREIGN KEY (document_id) REFERENCES application_documents(document_id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES application_collections(collection_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Application Collections-Properties Junction Table
CREATE TABLE IF NOT EXISTS application_collections_properties (
    collection_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (collection_id, property_id),
    FOREIGN KEY (collection_id) REFERENCES application_collections(collection_id) ON DELETE CASCADE,
    FOREIGN KEY (property_id) REFERENCES application_properties(property_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User Documents Table
-- Note: Assumes authorizer.authorizer_users table exists
CREATE TABLE IF NOT EXISTS user_documents (
    document_id BIGINT UNSIGNED AUTO_INCREMENT,
    user_id CHAR(36) NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    document_version BIGINT UNSIGNED NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, document_name),
    UNIQUE KEY unique_document_id (document_id),
    INDEX idx_user_id (user_id),
    INDEX idx_document_name (document_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User Collections Table
CREATE TABLE IF NOT EXISTS user_collections (
    collection_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    collection_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_collection_name (collection_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User Properties Table
CREATE TABLE IF NOT EXISTS user_properties (
    property_id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    property_name VARCHAR(255) NOT NULL,
    property_value JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CHECK (JSON_VALID(property_value)),
    INDEX idx_property_name (property_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User Documents-Collections Junction Table
CREATE TABLE IF NOT EXISTS user_documents_collections (
    document_id BIGINT UNSIGNED NOT NULL,
    collection_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (document_id, collection_id),
    FOREIGN KEY (document_id) REFERENCES user_documents(document_id) ON DELETE CASCADE,
    FOREIGN KEY (collection_id) REFERENCES user_collections(collection_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- User Collections-Properties Junction Table
CREATE TABLE IF NOT EXISTS user_collections_properties (
    collection_id BIGINT UNSIGNED NOT NULL,
    property_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (collection_id, property_id),
    FOREIGN KEY (collection_id) REFERENCES user_collections(collection_id) ON DELETE CASCADE,
    FOREIGN KEY (property_id) REFERENCES user_properties(property_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Migration complete
-- Note: This schema does NOT include stored procedures
-- All business logic is implemented in the Go application layer
