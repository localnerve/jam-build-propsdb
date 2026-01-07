package integration_test

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/database"
	"github.com/localnerve/propsdb/internal/handlers"
	"github.com/localnerve/propsdb/internal/services"
	"github.com/localnerve/propsdb/tests/helpers"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

// TestWithMariaDB tests the service with a real MariaDB container
func TestWithMariaDB(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("DB_IMAGE"),
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
		t.Fatalf("Failed to start MariaDB container: %v", err)
	}
	defer func() {
		if err := mariadbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MariaDB container: %v", err)
		}
	}()

	// Get container host and port
	host, err := mariadbContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := mariadbContainer.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Create config
	cfg := &config.Config{
		DBType:               "mysql",
		DBHost:               host,
		DBPort:               port.Port(),
		DBAppDatabase:        "testdb",
		DBAppUser:            "testuser",
		DBAppPassword:        "testpass",
		DBAppConnectionLimit: 5,
	}

	// Wait for database to be ready
	time.Sleep(5 * time.Second)

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Run tests
	t.Run("CreateAndRetrieveDocument", func(t *testing.T) {
		testCreateAndRetrieveDocument(t, db)
	})

	t.Run("VersionControl", func(t *testing.T) {
		testVersionControl(t, db)
	})

	t.Run("DeleteOperations", func(t *testing.T) {
		testDeleteOperations(t, db)
	})

	t.Run("Handler204Behavior", func(t *testing.T) {
		testHandler204Behavior(t, db)
	})
}

// TestWithPostgreSQL tests the service with a real PostgreSQL container
func TestWithPostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("POSTGRES_IMAGE"),
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_USER":     "testuser",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	// Get container host and port
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Create config
	cfg := &config.Config{
		DBType:               "postgres",
		DBHost:               host,
		DBPort:               port.Port(),
		DBAppDatabase:        "testdb",
		DBAppUser:            "testuser",
		DBAppPassword:        "testpass",
		DBAppConnectionLimit: 5,
	}

	// Wait for database to be ready
	time.Sleep(2 * time.Second)

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Run migrations
	if err := database.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Run tests
	t.Run("CreateAndRetrieveDocument", func(t *testing.T) {
		testCreateAndRetrieveDocument(t, db)
	})

	t.Run("VersionControl", func(t *testing.T) {
		testVersionControl(t, db)
	})

	t.Run("Handler204Behavior", func(t *testing.T) {
		testHandler204Behavior(t, db)
	})
}

// testCreateAndRetrieveDocument tests creating and retrieving a document
func testCreateAndRetrieveDocument(t *testing.T, db *gorm.DB) {
	// Create document with collections and properties
	collections := []services.CollectionInput{
		{
			Collection: "config",
			Properties: map[string]interface{}{
				"theme":    "dark",
				"language": "en",
				"count":    42,
			},
		},
		{
			Collection: "settings",
			Properties: map[string]interface{}{
				"enabled": true,
			},
		},
	}

	newVersion, _, err := services.SetApplicationProperties(db, "app1", 0, collections)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	if newVersion != 1 {
		t.Errorf("Expected version 1, got %d", newVersion)
	}

	// Retrieve document
	result, err := services.GetApplicationCollectionsAndProperties(db, "app1", []string{})
	if err != nil {
		t.Fatalf("Failed to retrieve document: %v", err)
	}

	// Verify structure
	if result["app1"] == nil {
		t.Fatal("Expected app1 in result")
	}

	docMap := result["app1"].(map[string]interface{})
	if docMap["__version"] != "1" {
		t.Errorf("Expected version 1, got %v", docMap["__version"])
	}

	if docMap["config"] == nil {
		t.Error("Expected config collection")
	}
}

// testVersionControl tests optimistic locking
func testVersionControl(t *testing.T, db *gorm.DB) {
	// Create initial document
	collections := []services.CollectionInput{
		{
			Collection: "data",
			Properties: map[string]interface{}{
				"value": "initial",
			},
		},
	}

	_, _, err := services.SetApplicationProperties(db, "versiontest", 0, collections)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Try to update with wrong version
	collections[0].Properties["value"] = "updated"
	_, _, err = services.SetApplicationProperties(db, "versiontest", 0, collections)
	if err == nil {
		t.Error("Expected version conflict error")
	}

	if err.Error() != "E_VERSION" {
		t.Errorf("Expected E_VERSION error, got: %v", err)
	}

	// Update with correct version
	_, _, err = services.SetApplicationProperties(db, "versiontest", 1, collections)
	if err != nil {
		t.Errorf("Failed to update with correct version: %v", err)
	}
}

// testDeleteOperations tests delete functionality
func testDeleteOperations(t *testing.T, db *gorm.DB) {
	// Create document
	collections := []services.CollectionInput{
		{
			Collection: "coll1",
			Properties: map[string]interface{}{
				"prop1": "value1",
				"prop2": "value2",
			},
		},
		{
			Collection: "coll2",
			Properties: map[string]interface{}{
				"prop3": "value3",
			},
		},
	}

	_, _, err := services.SetApplicationProperties(db, "deletetest", 0, collections)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Delete a collection
	_, _, err = services.DeleteApplicationCollection(db, "deletetest", 1, "coll1")
	if err != nil {
		t.Fatalf("Failed to delete collection: %v", err)
	}

	// Verify collection is deleted
	result, err := services.GetApplicationCollectionsAndProperties(db, "deletetest", []string{})
	if err != nil {
		t.Fatalf("Failed to retrieve document: %v", err)
	}

	docMap := result["deletetest"].(map[string]interface{})
	if docMap["coll1"] != nil {
		t.Error("Expected coll1 to be deleted")
	}

	if docMap["coll2"] == nil {
		t.Error("Expected coll2 to still exist")
	}
}

// TestHealthCheck tests the health check functionality
func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start MariaDB container
	mariadbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        os.Getenv("DB_IMAGE"),
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
		t.Fatalf("Failed to start MariaDB container: %v", err)
	}
	defer func() {
		if err := mariadbContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate MariaDB container: %v", err)
		}
	}()

	host, err := mariadbContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := mariadbContainer.MappedPort(ctx, "3306")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	cfg := &config.Config{
		DBType:        "mysql",
		DBHost:        host,
		DBPort:        port.Port(),
		DBAppDatabase: "testdb",
		DBAppUser:     "testuser",
		DBAppPassword: "testpass",
		AuthzURL:      "http://localhost:9999", // Non-existent service
	}

	time.Sleep(5 * time.Second)

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Run health check
	result := services.HealthCheck(cfg, db)

	// Database should be healthy
	if result.Database != "ok" {
		t.Errorf("Expected database to be ok, got: %s", result.Database)
	}

	// Authorizer should be unreachable
	if result.Authorizer != "unreachable" {
		t.Errorf("Expected authorizer to be unreachable, got: %s", result.Authorizer)
	}

	// Overall status should be unhealthy
	if result.Status != "unhealthy" {
		t.Errorf("Expected status to be unhealthy, got: %s", result.Status)
	}
}

// testHandler204Behavior tests the handler's 204 No Content response with a real database
func testHandler204Behavior(t *testing.T, db *gorm.DB) {
	docName := "int-emptydoc"
	colName := "int-emptycoll"

	helpers.CreateTestDocument(t, db, docName, 1)
	helpers.CreateTestEmptyCollection(t, db, docName, colName)

	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Get("/api/data/app/:document/:collection", handler.GetAppProperties)

	// Single empty collection -> 204
	req := httptest.NewRequest("GET", "/api/data/app/"+docName+"/"+colName, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)

	// Multi collection (filtered) all empty -> 204
	app.Get("/api/data/app/:document", handler.GetAppCollectionsAndProperties)
	req = httptest.NewRequest("GET", "/api/data/app/"+docName+"?collections="+colName, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)
}
