package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestE2EWithFullStack tests the entire service stack
func TestE2EWithFullStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Start MariaDB
	mariadbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "mariadb:11.2",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "rootpass",
				"MYSQL_DATABASE":      "testdb",
				"MYSQL_USER":          "testuser",
				"MYSQL_PASSWORD":      "testpass",
			},
			WaitingFor: wait.ForLog("ready for connections").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start MariaDB: %v", err)
	}
	defer func() {
		if err := mariadbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MariaDB: %v", err)
		}
	}()

	mariadbHost, _ := mariadbContainer.Host(ctx)
	mariadbPort, _ := mariadbContainer.MappedPort(ctx, "3306")

	// Start Authorizer
	authorizerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "localnerve/authorizer:1.5.3",
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"DATABASE_TYPE":     "mysql",
				"DATABASE_HOST":     mariadbHost,
				"DATABASE_PORT":     mariadbPort.Port(),
				"DATABASE_NAME":     "testdb",
				"DATABASE_USERNAME": "testuser",
				"DATABASE_PASSWORD": "testpass",
				"ADMIN_SECRET":      "admin_secret",
				"JWT_SECRET":        "jwt_secret",
			},
			WaitingFor: wait.ForHTTP("/").WithPort("8080").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start Authorizer: %v", err)
	}
	defer func() {
		if err := authorizerContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Authorizer: %v", err)
		}
	}()

	authorizerHost, _ := authorizerContainer.Host(ctx)
	authorizerPort, _ := authorizerContainer.MappedPort(ctx, "8080")
	authorizerURL := fmt.Sprintf("http://%s:%s", authorizerHost, authorizerPort.Port())

	// Build and start PropsDB service
	propsdbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    "../..",
				Dockerfile: "Dockerfile",
			},
			ExposedPorts: []string{"3000/tcp"},
			Env: map[string]string{
				"DB_TYPE":                 "mysql",
				"DB_HOST":                 mariadbHost,
				"DB_PORT":                 mariadbPort.Port(),
				"DB_DATABASE":             "testdb",
				"DB_APP_USER":             "testuser",
				"DB_APP_PASSWORD":         "testpass",
				"DB_USER":                 "testuser",
				"DB_PASSWORD":             "testpass",
				"DB_APP_CONNECTION_LIMIT": "5",
				"DB_CONNECTION_LIMIT":     "5",
				"AUTHZ_URL":               authorizerURL,
				"AUTHZ_CLIENT_ID":         "test_client",
				"PORT":                    "3000",
			},
			WaitingFor: wait.ForHTTP("/metrics").WithPort("3000").WithStartupTimeout(120 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start PropsDB: %v", err)
	}
	defer func() {
		if err := propsdbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate PropsDB: %v", err)
		}
	}()

	propsdbHost, _ := propsdbContainer.Host(ctx)
	propsdbPort, _ := propsdbContainer.MappedPort(ctx, "3000")
	baseURL := fmt.Sprintf("http://%s:%s", propsdbHost, propsdbPort.Port())

	// Wait a bit for everything to stabilize
	time.Sleep(5 * time.Second)

	// Run E2E tests
	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t, baseURL)
	})

	t.Run("PrometheusMetrics", func(t *testing.T) {
		testPrometheusMetrics(t, baseURL)
	})

	t.Run("SwaggerUI", func(t *testing.T) {
		testSwaggerUI(t, baseURL)
	})

	t.Run("PublicAPIAccess", func(t *testing.T) {
		testPublicAPIAccess(t, baseURL)
	})

	t.Run("VersionHeader", func(t *testing.T) {
		testVersionHeader(t, baseURL)
	})
}

func testHealthCheck(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func testPrometheusMetrics(t *testing.T, baseURL string) {
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Check for expected Prometheus metrics
	if !bytes.Contains(body, []byte("propsdb_http_requests_total")) {
		t.Error("Expected propsdb_http_requests_total metric")
	}

	if !bytes.Contains(body, []byte("go_goroutines")) {
		t.Error("Expected go_goroutines metric")
	}

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

	// Should return 200 with empty data or proper JSON
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Logf("Response body: %s", string(body))
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify response is valid JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}
}

func testVersionHeader(t *testing.T, baseURL string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/api/data/app", nil)
	req.Header.Set("X-Api-Version", "1.0")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request with version header: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 with version header, got %d", resp.StatusCode)
	}
}
