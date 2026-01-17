// user_handlers_test.go
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
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/localnerve/jam-build-propsdb/internal/handlers"
	"github.com/localnerve/jam-build-propsdb/internal/models"
	"github.com/localnerve/jam-build-propsdb/tests/helpers"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupUserTestDB creates an in-memory SQLite database for user testing
func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.UserDocument{},
		&models.UserCollection{},
		&models.UserProperty{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// CreateTestUserDocument creates a user document directly via GORM
func CreateTestUserDocument(t *testing.T, db *gorm.DB, userID, docName string, version uint64) {
	doc := models.UserDocument{
		UserID:          userID,
		DocumentName:    docName,
		DocumentVersion: version,
	}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create user document: %v", err)
	}
}

// CreateTestUserEmptyCollection creates a user collection with no properties
func CreateTestUserEmptyCollection(t *testing.T, db *gorm.DB, userID, docName, colName string) {
	var doc models.UserDocument
	if err := db.Where("user_id = ? AND document_name = ?", userID, docName).First(&doc).Error; err != nil {
		t.Fatalf("Failed to find user document %s: %v", docName, err)
	}

	coll := models.UserCollection{
		CollectionName: colName,
	}
	if err := db.Model(&doc).Association("Collections").Append(&coll); err != nil {
		t.Fatalf("Failed to associate collection: %v", err)
	}
}

func TestGetUserProperties_Empty(t *testing.T) {
	db := setupUserTestDB(t)
	userID := "user-123"
	docName := "emptydoc"
	colName := "emptycoll"

	CreateTestUserDocument(t, db, userID, docName, 1)
	CreateTestUserEmptyCollection(t, db, userID, docName, colName)

	app := fiber.New()
	// Mock auth middleware to set user in context
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"id": userID,
		})
		return c.Next()
	})

	handler := &handlers.UserDataHandler{DB: db}
	app.Get("/api/data/user/:document/:collection", handler.GetUserProperties)

	req := httptest.NewRequest("GET", "/api/data/user/"+docName+"/"+colName, nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	helpers.AssertStatus(t, resp, 204)
	helpers.AssertNoContent(t, resp)
}

func TestGetUserCollectionsAndProperties_Empty(t *testing.T) {
	db := setupUserTestDB(t)
	userID := "user-456"
	docName := "multidoc"

	CreateTestUserDocument(t, db, userID, docName, 1)
	CreateTestUserEmptyCollection(t, db, userID, docName, "coll1")
	CreateTestUserEmptyCollection(t, db, userID, docName, "coll2")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", map[string]interface{}{
			"id": userID,
		})
		return c.Next()
	})

	handler := &handlers.UserDataHandler{DB: db}
	app.Get("/api/data/user/:document", handler.GetUserCollectionsAndProperties)

	// Single empty collection requested
	req := httptest.NewRequest("GET", "/api/data/user/"+docName+"?collections=coll1", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	helpers.AssertStatus(t, resp, 204)

	// Multiple empty collections requested
	req = httptest.NewRequest("GET", "/api/data/user/"+docName+"?collections=coll1,coll2", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	helpers.AssertStatus(t, resp, 204)

	// All collections (implicitly empty)
	req = httptest.NewRequest("GET", "/api/data/user/"+docName, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}
	helpers.AssertStatus(t, resp, 204)
}
