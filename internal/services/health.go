package services

import (
	"fmt"
	"log"

	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/utils"
	"gorm.io/gorm"
)

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status       string            `json:"status"`
	Database     string            `json:"database"`
	Authorizer   string            `json:"authorizer"`
	Details      map[string]string `json:"details,omitempty"`
	ErrorMessage string            `json:"error,omitempty"`
}

// HealthCheck performs a comprehensive health check of the service
func HealthCheck(cfg *config.Config, db *gorm.DB) HealthCheckResult {
	result := HealthCheckResult{
		Status:  "healthy",
		Details: make(map[string]string),
	}

	// Check database connectivity
	sqlDB, err := db.DB()
	if err != nil {
		result.Status = "unhealthy"
		result.Database = "error"
		result.Details["database_error"] = err.Error()
		result.ErrorMessage = fmt.Sprintf("Database connection error: %v", err)
		log.Printf("Health check failed - database connection: %v", err)
	} else {
		if err := sqlDB.Ping(); err != nil {
			result.Status = "unhealthy"
			result.Database = "unreachable"
			result.Details["database_ping_error"] = err.Error()
			result.ErrorMessage = fmt.Sprintf("Database ping failed: %v", err)
			log.Printf("Health check failed - database ping: %v", err)
		} else {
			result.Database = "ok"
			result.Details["database_type"] = cfg.DBType
			result.Details["database_name"] = cfg.DBAppDatabase
		}
	}

	// Check Authorizer connectivity
	if err := utils.PingAuthorizer(cfg.AuthzURL); err != nil {
		result.Status = "unhealthy"
		result.Authorizer = "unreachable"
		result.Details["authorizer_error"] = err.Error()
		if result.ErrorMessage == "" {
			result.ErrorMessage = fmt.Sprintf("Authorizer ping failed: %v", err)
		} else {
			result.ErrorMessage += fmt.Sprintf("; Authorizer ping failed: %v", err)
		}
		log.Printf("Health check failed - authorizer ping: %v", err)
	} else {
		result.Authorizer = "ok"
		result.Details["authorizer_url"] = cfg.AuthzURL
	}

	if result.Status == "healthy" {
		log.Println("Health check passed - all systems operational")
	}

	return result
}
