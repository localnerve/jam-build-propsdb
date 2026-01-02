package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/propsdb/internal/handlers"
	"github.com/localnerve/propsdb/internal/models"
	"github.com/localnerve/propsdb/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.ApplicationDocument{},
		&models.ApplicationCollection{},
		&models.ApplicationProperty{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// TestGetAppProperties tests the GET /api/data/app/:document/:collection endpoint
func TestGetAppProperties(t *testing.T) {
	db := setupTestDB(t)

	// Create test data using GORM associations
	doc := models.ApplicationDocument{
		DocumentName:    "testdoc",
		DocumentVersion: 1,
	}
	db.Create(&doc)

	coll := models.ApplicationCollection{
		CollectionName: "testcoll",
	}
	db.Create(&coll)

	prop := models.ApplicationProperty{
		PropertyName:  "testprop",
		PropertyValue: []byte(`"testvalue"`),
	}
	db.Create(&prop)

	// Associate using GORM
	if err := db.Model(&doc).Association("Collections").Append(&coll); err != nil {
		t.Fatalf("Failed to associate collection: %v", err)
	}
	if err := db.Model(&coll).Association("Properties").Append(&prop); err != nil {
		t.Fatalf("Failed to associate property: %v", err)
	}

	// Create Fiber app and handler
	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Get("/api/data/app/:document/:collection", handler.GetAppProperties)

	// Test request
	req := httptest.NewRequest("GET", "/api/data/app/testdoc/testcoll", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if result["testdoc"] == nil {
		t.Error("Expected 'testdoc' in response")
	}
}

// TestSetAppProperties tests the POST /api/data/app/:document endpoint
func TestSetAppProperties(t *testing.T) {
	db := setupTestDB(t)

	// Create Fiber app and handler
	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Post("/api/data/app/:document", handler.SetAppProperties)

	// Prepare request body
	reqBody := map[string]interface{}{
		"version": 0,
		"collections": []services.CollectionInput{
			{
				Collection: "testcoll",
				Properties: map[string]interface{}{
					"prop1": "value1",
					"prop2": 123,
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/data/app/testdoc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Check status code
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if result["ok"] != true {
		t.Error("Expected ok=true in response")
	}

	if result["newVersion"] == nil {
		t.Error("Expected newVersion in response")
	}
}

// TestVersionConflict tests version conflict detection
func TestVersionConflict(t *testing.T) {
	db := setupTestDB(t)

	// Create initial document
	doc := models.ApplicationDocument{
		DocumentName:    "testdoc",
		DocumentVersion: 1,
	}
	db.Create(&doc)

	// Create Fiber app and handler
	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Post("/api/data/app/:document", handler.SetAppProperties)

	// Try to update with wrong version
	reqBody := map[string]interface{}{
		"version": 0, // Wrong version (should be 1)
		"collections": []services.CollectionInput{
			{
				Collection: "testcoll",
				Properties: map[string]interface{}{
					"prop1": "value1",
				},
			},
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/data/app/testdoc", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Should return 409 Conflict
	if resp.StatusCode != 409 {
		t.Errorf("Expected status 409 (version conflict), got %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify version error
	if result["versionError"] != true {
		t.Error("Expected versionError=true in response")
	}
}

// TestNotFound tests 404 responses
func TestNotFound(t *testing.T) {
	db := setupTestDB(t)

	// Create Fiber app and handler
	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Get("/api/data/app/:document/:collection", handler.GetAppProperties)

	// Request non-existent document
	req := httptest.NewRequest("GET", "/api/data/app/nonexistent/collection", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Should return 404
	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}
