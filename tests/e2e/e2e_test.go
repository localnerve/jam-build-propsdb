package e2e_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestE2EWithFullStack tests the entire service stack
func TestE2EWithFullStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	ctx := context.Background()

	// Create a network
	nw, err := network.New(ctx)
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	networkName := nw.Name
	defer func() {
		if err := nw.Remove(ctx); err != nil {
			t.Logf("Failed to remove network: %v", err)
		}
	}()

	// Start MariaDB
	mariadbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("DOCKER_MARIADB_IMAGE"),
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": "rootpass",
			},
			WaitingFor: wait.ForListeningPort("3306/tcp").WithStartupTimeout(60 * time.Second),
			Networks:   []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"mariadb"},
			},
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

	// Create databases
	mariadbHost, _ := mariadbContainer.Host(ctx)
	mariadbPort, _ := mariadbContainer.MappedPort(ctx, "3306")
	db, err := sql.Open("mysql", fmt.Sprintf("root:rootpass@tcp(%s:%s)/", mariadbHost, mariadbPort.Port()))
	if err != nil {
		t.Fatalf("Failed to connect to MariaDB for setup: %v", err)
	}
	defer db.Close()

	// Wait for connection to be really ready
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("MariaDB not ready after 30 seconds: %v", err)
	}

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS testdb")
	if err != nil {
		t.Fatalf("Failed to create testdb: %v", err)
	}
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS authorizer")
	if err != nil {
		t.Fatalf("Failed to create authorizer db: %v", err)
	}
	_, err = db.Exec("GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' IDENTIFIED BY 'rootpass' WITH GRANT OPTION")
	if err != nil {
		t.Fatalf("Failed to grant privileges: %v", err)
	}
	_, err = db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		t.Fatalf("Failed to flush privileges: %v", err)
	}

	// Start Authorizer
	authorizerContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("DOCKER_AUTHZ_IMAGE"),
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"ENV":           "production",
				"DATABASE_TYPE": "mariadb",
				"DATABASE_URL":  "root:rootpass@tcp(mariadb:3306)/authorizer",
				"ADMIN_SECRET":  "admin_secret",
				"JWT_SECRET":    "jwt_secret",
				"ROLES":         "admin,user",
				"DEFAULT_ROLES": "user",
			},
			WaitingFor: wait.ForLog("Authorizer running at PORT:").WithStartupTimeout(60 * time.Second),
			Networks:   []string{networkName},
			NetworkAliases: map[string][]string{
				networkName: {"authorizer"},
			},
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

	// Check if propsdb-test image exists
	imageName := "propsdb-test:latest"
	imageExists, err := imageExists(ctx, imageName)
	if err != nil {
		t.Fatalf("Failed to check if image exists: %v", err)
	}

	var propsdbBuilderContainer testcontainers.Container
	var propsdbContainer testcontainers.Container

	if !imageExists {
		// Build and start PropsDB service
		propsdbResourceReaperSessionID := uuid.New().String()

		t.Logf("Image %s does not exist, building...", imageName)
		propsdbBuilderContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context:    "../..",
					Dockerfile: "Dockerfile",
					Repo:       "propsdb-test-builder",
					Tag:        "latest",
					BuildArgs: map[string]*string{
						"RESOURCE_REAPER_SESSION_ID": &propsdbResourceReaperSessionID,
					},
					BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
						opts.Target = "builder" // Build specific stage
					},
					PrintBuildLog: true,
				},
			},
			Started: false,
		})
		if err != nil {
			t.Fatalf("Failed to build propsdb-test-builder: %v", err)
		}
		defer func() {
			if err := propsdbBuilderContainer.Terminate(ctx); err != nil {
				t.Logf("Failed to terminate PropsDB Builder: %v", err)
			}
		}()

		propsdbContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				FromDockerfile: testcontainers.FromDockerfile{
					Context:    "../..",
					Dockerfile: "Dockerfile",
					Repo:       "propsdb-test",
					Tag:        "latest",
					KeepImage:  true, // Keep the image so we can reuse it
					BuildArgs: map[string]*string{
						"RESOURCE_REAPER_SESSION_ID": &propsdbResourceReaperSessionID,
					},
					BuildOptionsModifier: func(opts *build.ImageBuildOptions) {
						opts.Target = "runtime" // Build specific stage
					},
					PrintBuildLog: true,
				},
				ExposedPorts: []string{"3000/tcp"},
				Env: map[string]string{
					"DB_TYPE":                 "mysql",
					"DB_HOST":                 "mariadb",
					"DB_PORT":                 "3306",
					"DB_APP_DATABASE":         "testdb",
					"DB_APP_USER":             "root",
					"DB_APP_PASSWORD":         "rootpass",
					"DB_USER":                 "root",
					"DB_PASSWORD":             "rootpass",
					"DB_APP_CONNECTION_LIMIT": "5",
					"DB_CONNECTION_LIMIT":     "5",
					"AUTHZ_URL":               "http://authorizer:8080",
					"AUTHZ_CLIENT_ID":         "test_client",
					"PORT":                    "3000",
				},
				WaitingFor: wait.ForHTTP("/metrics").WithPort("3000").WithStartupTimeout(120 * time.Second),
				Networks:   []string{networkName},
			},
			Started: true,
		})
	} else {
		t.Logf("Image %s exists, reusing...", imageName)
		propsdbContainer, err = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        imageName,
				ExposedPorts: []string{"3000/tcp"},
				Env: map[string]string{
					"DB_TYPE":                 "mysql",
					"DB_HOST":                 "mariadb",
					"DB_PORT":                 "3306",
					"DB_APP_DATABASE":         "testdb",
					"DB_APP_USER":             "root",
					"DB_APP_PASSWORD":         "rootpass",
					"DB_USER":                 "root",
					"DB_PASSWORD":             "rootpass",
					"DB_APP_CONNECTION_LIMIT": "5",
					"DB_CONNECTION_LIMIT":     "5",
					"AUTHZ_URL":               "http://authorizer:8080",
					"AUTHZ_CLIENT_ID":         "test_client",
					"PORT":                    "3000",
				},
				WaitingFor: wait.ForHTTP("/metrics").WithPort("3000").WithStartupTimeout(120 * time.Second),
				Networks:   []string{networkName},
			},
			Started: true,
		})
	}
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

func imageExists(ctx context.Context, imageName string) (bool, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}

	return false, nil
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
