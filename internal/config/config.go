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
	DBType              string // mysql, postgres, sqlite, sqlserver, etc.
	DBHost              string
	DBPort              string
	DBDatabase          string
	DBAppUser           string
	DBAppPassword       string
	DBAppConnectionLimit int
	DBUser              string
	DBPassword          string
	DBConnectionLimit   int

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
		DBDatabase:           getEnv("DB_DATABASE", ""),
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
	if cfg.DBDatabase == "" {
		return nil, fmt.Errorf("DB_DATABASE is required")
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
