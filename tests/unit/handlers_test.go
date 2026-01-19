// handlers_test.go
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

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/jam-build-propsdb/internal/handlers"
	"github.com/localnerve/jam-build-propsdb/internal/models"
	"github.com/localnerve/jam-build-propsdb/internal/services"
	"github.com/localnerve/jam-build-propsdb/tests/helpers"
	"gorm.io/datatypes"
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
		PropertyValue: models.JSON{JSON: datatypes.JSON([]byte(`"testvalue"`))},
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

// TestGetAppProperties_Empty tests GET /api/data/app/:document/:collection with empty collection
func TestGetAppProperties_Empty(t *testing.T) {
	db := setupTestDB(t)

	helpers.CreateTestDocument(t, db, "emptydoc", 1)
	helpers.CreateTestEmptyCollection(t, db, "emptydoc", "emptycoll")

	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Get("/api/data/app/:document/:collection", handler.GetAppProperties)

	req := httptest.NewRequest("GET", "/api/data/app/emptydoc/emptycoll", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)
}

// TestGetAppCollectionsAndProperties_Empty tests multi-collection GET with only empty collections
func TestGetAppCollectionsAndProperties_Empty(t *testing.T) {
	db := setupTestDB(t)

	helpers.CreateTestDocument(t, db, "multidoc", 1)
	helpers.CreateTestEmptyCollection(t, db, "multidoc", "coll1")
	helpers.CreateTestEmptyCollection(t, db, "multidoc", "coll2")

	app := fiber.New()
	handler := &handlers.AppDataHandler{DB: db}
	app.Get("/api/data/app/:document", handler.GetAppCollectionsAndProperties)

	// Filtered multi-collection
	req := httptest.NewRequest("GET", "/api/data/app/multidoc?collections=coll1,coll2", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)

	// All collections (implicitly empty)
	req = httptest.NewRequest("GET", "/api/data/app/multidoc", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)
}
