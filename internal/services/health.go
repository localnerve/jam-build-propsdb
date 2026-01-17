// health.go
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

package services

import (
	"fmt"
	"log"

	"github.com/localnerve/jam-build-propsdb/internal/config"
	"github.com/localnerve/jam-build-propsdb/internal/utils"
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
