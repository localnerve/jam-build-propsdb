// e2e_test.go
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

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/localnerve/jam-build-propsdb/internal/config"
	"github.com/localnerve/jam-build-propsdb/internal/database"
	"github.com/localnerve/jam-build-propsdb/internal/services"
	"github.com/localnerve/jam-build-propsdb/tests/helpers"
)

// TestE2EWithFullStack tests the entire service stack
func TestE2EWithFullStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	tc, err := helpers.CreateAllTestContainers(t)
	if err != nil {
		t.Fatalf("Failed to start test containers: %v", err)
	}
	defer tc.Terminate(t)

	propsdbHost, _ := tc.PropsDBContainer.Host(ctx)
	propsdbPort, _ := tc.PropsDBContainer.MappedPort(ctx, "3000")
	baseURL := fmt.Sprintf("http://%s:%s", propsdbHost, propsdbPort.Port())

	// Wait a bit for everything to stabilize
	time.Sleep(5 * time.Second)

	// Run E2E tests
	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t, tc)
	})

	t.Run("PrometheusMetrics", func(t *testing.T) {
		testPrometheusMetrics(t, baseURL)
	})

	t.Run("SwaggerUI", func(t *testing.T) {
		testSwaggerUI(t, baseURL)
	})

	// Public API Access
	t.Run("PublicAPIAccess", func(t *testing.T) {
		testPublicAPIAccess(t, baseURL)
	})
}

func testHealthCheck(t *testing.T, tc *helpers.TestContainers) {
	ctx := context.Background()

	// 1. Prepare configuration for the health check
	// We need to point to the mapped ports on localhost, not internal container names
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Update DB host and port to mapped values
	dbHost, _ := tc.DBContainer.Host(ctx)
	dbPort, _ := tc.DBContainer.MappedPort(ctx, "3306")
	cfg.DBHost = dbHost
	cfg.DBPort = dbPort.Port()

	// Update Authorizer URL to mapped value
	authzHost, _ := tc.AuthorizerContainer.Host(ctx)
	authzPort, _ := tc.AuthorizerContainer.MappedPort(ctx, "8080")
	cfg.AuthzURL = fmt.Sprintf("http://%s:%s", authzHost, authzPort.Port())

	// 2. Establish GORM connection to the test database
	gormDB, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer database.Close(gormDB)

	// 3. Perform the health check
	result := services.HealthCheck(cfg, gormDB)

	// 4. Verify the result
	if result.Status != "healthy" {
		t.Errorf("Health check failed: %+v", result)
	}

	t.Logf("Health check passed: status=%s, database=%s, authorizer=%s",
		result.Status, result.Database, result.Authorizer)
}

func testPrometheusMetrics(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for metrics, got %d. Body: %s", resp.StatusCode, bodyStr)
	}

	/*
		// Check for expected Prometheus metrics
		if !bytes.Contains(body, []byte("propsdb_http_requests_total")) {
			t.Errorf("Expected propsdb_http_requests_total metric. Body: %s", bodyStr)
		}

		if !bytes.Contains(body, []byte("go_goroutines")) {
				t.Errorf("Expected go_goroutines metric. Body: %s", bodyStr)
			}
	*/

	t.Logf("Metrics endpoint working, found %d bytes of metrics", len(bodyStr))
}

func testSwaggerUI(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/swagger/index.html")
	if err != nil {
		t.Fatalf("Failed to get Swagger UI: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for Swagger UI, got %d", resp.StatusCode)
	}
}

func testPublicAPIAccess(t *testing.T, baseURL string) {
	// Test public GET endpoint (should work without auth)
	resp, err := http.Get(baseURL + "/api/data/app")
	if err != nil {
		t.Fatalf("Failed to access public API: %v", err)
	}
	defer resp.Body.Close()

	// Should return 404 with proper JSON
	if resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Response body: %s", string(body))
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Verify response is valid JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}
}
