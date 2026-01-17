// config.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port string

	// Database configuration
	DBType               string // mysql, postgres, sqlite, sqlserver, etc.
	DBHost               string
	DBPort               string
	DBAppDatabase        string
	DBAppUser            string
	DBAppPassword        string
	DBAppConnectionLimit int
	DBUser               string
	DBPassword           string
	DBConnectionLimit    int

	// Authorizer configuration
	AuthzURL      string
	AuthzClientID string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:                 getEnv("PORT", "3000"),
		DBType:               getEnv("DB_TYPE", "mysql"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBAppDatabase:        getEnv("DB_APP_DATABASE", ""),
		DBAppUser:            getEnv("DB_APP_USER", ""),
		DBAppPassword:        getEnv("DB_APP_PASSWORD", ""),
		DBAppConnectionLimit: getEnvAsInt("DB_APP_CONNECTION_LIMIT", 5),
		DBUser:               getEnv("DB_USER", ""),
		DBPassword:           getEnv("DB_PASSWORD", ""),
		DBConnectionLimit:    getEnvAsInt("DB_CONNECTION_LIMIT", 5),
		AuthzURL:             getEnv("AUTHZ_URL", ""),
		AuthzClientID:        getEnv("AUTHZ_CLIENT_ID", ""),
	}

	// Validate required fields
	if cfg.DBAppDatabase == "" {
		return nil, fmt.Errorf("DB_APP_DATABASE is required")
	}
	if cfg.DBAppUser == "" {
		return nil, fmt.Errorf("DB_APP_USER is required")
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if cfg.AuthzURL == "" {
		return nil, fmt.Errorf("AUTHZ_URL is required")
	}
	if cfg.AuthzClientID == "" {
		return nil, fmt.Errorf("AUTHZ_CLIENT_ID is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
