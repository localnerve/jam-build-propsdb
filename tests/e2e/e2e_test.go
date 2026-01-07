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
	"github.com/localnerve/propsdb/tests/helpers"
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

	/*
		authzHost, _ := authorizerContainer.Host(ctx)
		authzPort, _ := authorizerContainer.MappedPort(ctx, "8080")
		authzURL := fmt.Sprintf("http://%s:%s", authzHost, authzPort.Port())
	*/

	// Wait a bit for everything to stabilize
	time.Sleep(5 * time.Second)

	// Run E2E tests
	t.Run("HealthCheck", func(t *testing.T) {
		testHealthCheck(t, baseURL)
	})

	/*
		t.Run("PrometheusMetrics", func(t *testing.T) {
			testPrometheusMetrics(t, baseURL)
		})
	*/

	t.Run("SwaggerUI", func(t *testing.T) {
		testSwaggerUI(t, baseURL)
	})

	// Public API Access
	t.Run("PublicAPIAccess", func(t *testing.T) {
		testPublicAPIAccessEmpty(t, baseURL)
	})

	// Version Header
	/*
		t.Run("VersionHeader", func(t *testing.T) {
			testVersionHeader(t, baseURL)
		})
	*/

	// User Data 204 Behavior
	/*
		t.Run("UserData204Behavior", func(t *testing.T) {
			testUserData204Behavior(t, baseURL, authzURL, db)
		})
	*/
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

/*
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

	// Check for expected Prometheus metrics
	if !bytes.Contains(body, []byte("propsdb_http_requests_total")) {
		t.Errorf("Expected propsdb_http_requests_total metric. Body: %s", bodyStr)
	}

	if !bytes.Contains(body, []byte("go_goroutines")) {
		t.Errorf("Expected go_goroutines metric. Body: %s", bodyStr)
	}

	t.Logf("Metrics endpoint working, found %d bytes of metrics", len(bodyStr))
}
*/

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

func testPublicAPIAccessEmpty(t *testing.T, baseURL string) {
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

/*
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
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200 with version header, got %d. Body: %s", resp.StatusCode, string(body))
	}
}

func testUserData204Behavior(t *testing.T, baseURL, authzURL string, db *sql.DB) {
	// 1. Setup Auth
	email := fmt.Sprintf("user-%s@test.local", uuid.New().String()[:8])
	password := helpers.GeneratePassword()
	token := helpers.AcquireAccount(t, authzURL, email, password, []string{"user"})

	// 2. Setup Data (directly in DB since we can't easily Set user data without more complex API calls here)
	// We need to find the user ID created by authorizer
	var userID string
	var err error
	// Retry loop for user creation propagation to DB
	for i := 0; i < 10; i++ {
		err = db.QueryRow("SELECT id FROM authorizer.authorizer_users WHERE email = ?", email).Scan(&userID)
		if err == nil {
			break
		}
		t.Logf("Attempt %d: Failed to find created user ID, retrying... (%v)", i+1, err)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		// If still failing, let's list tables for debugging
		rows, listErr := db.Query("SHOW TABLES FROM authorizer")
		if listErr == nil {
			var tableName string
			t.Log("Tables in 'authorizer' database:")
			for rows.Next() {
				rows.Scan(&tableName)
				t.Logf("- %s", tableName)
			}
		}
		t.Fatalf("Failed to find created user ID after retries: %v", err)
	}

	// Insert doc and empty collection into propsdb DB
	// We use the 'db' which is connected to root, so we can access testdb
	_, err = db.Exec("INSERT INTO testdb.user_documents (user_id, document_name, document_version) VALUES (?, ?, ?)", userID, "e2e-doc", 1)
	if err != nil {
		t.Fatalf("Failed to insert user document: %v", err)
	}
	var docID int64
	err = db.QueryRow("SELECT document_id FROM testdb.user_documents WHERE user_id = ? AND document_name = ?", userID, "e2e-doc").Scan(&docID)
	if err != nil {
		t.Fatalf("Failed to get doc ID: %v", err)
	}
	_, err = db.Exec("INSERT INTO testdb.user_collections (collection_name) VALUES (?)", "e2e-emptycoll")
	if err != nil {
		t.Fatalf("Failed to insert user collection: %v", err)
	}
	var collID int64
	err = db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&collID)
	if err != nil {
		t.Fatalf("Failed to get coll ID: %v", err)
	}
	_, err = db.Exec("INSERT INTO testdb.user_documents_collections (document_id, collection_id) VALUES (?, ?)", docID, collID)
	if err != nil {
		t.Fatalf("Failed to link doc and coll: %v", err)
	}

	// 3. Verify 204
	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/api/data/user/e2e-doc/e2e-emptycoll", nil)
	req.AddCookie(&http.Cookie{Name: "cookie_session", Value: token})

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 204, got %d. Body: %s", resp.StatusCode, string(body))
	}
}
*/
