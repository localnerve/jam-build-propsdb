// data.go
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

package helpers

import (
	"encoding/json"
	"testing"

	"github.com/localnerve/jam-build-propsdb/internal/models"
	"gorm.io/gorm"
)

// CreateTestDocument creates a test application document
func CreateTestDocument(t *testing.T, db *gorm.DB, name string, version uint64) {
	doc := models.ApplicationDocument{
		DocumentName:    name,
		DocumentVersion: version,
	}
	if err := db.Create(&doc).Error; err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}
}

// CreateTestCollection creates a collection with properties
func CreateTestCollection(t *testing.T, db *gorm.DB, docName, colName string, properties map[string]interface{}) {
	var doc models.ApplicationDocument
	if err := db.Where("document_name = ?", docName).First(&doc).Error; err != nil {
		t.Fatalf("Failed to find document %s: %v", docName, err)
	}

	coll := models.ApplicationCollection{
		CollectionName: colName,
	}
	if err := db.Model(&doc).Association("Collections").Append(&coll); err != nil {
		t.Fatalf("Failed to associate collection: %v", err)
	}

	for k, v := range properties {
		jsonVal, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Failed to marshal property value: %v", err)
		}

		prop := models.ApplicationProperty{
			PropertyName:  k,
			PropertyValue: jsonVal,
		}
		if err := db.Model(&coll).Association("Properties").Append(&prop); err != nil {
			t.Fatalf("Failed to associate property: %v", err)
		}
	}
}

// CreateTestEmptyCollection creates a collection with no properties
func CreateTestEmptyCollection(t *testing.T, db *gorm.DB, docName, colName string) {
	var doc models.ApplicationDocument
	if err := db.Where("document_name = ?", docName).First(&doc).Error; err != nil {
		t.Fatalf("Failed to find document %s: %v", docName, err)
	}

	coll := models.ApplicationCollection{
		CollectionName: colName,
	}
	if err := db.Model(&doc).Association("Collections").Append(&coll); err != nil {
		t.Fatalf("Failed to associate collection: %v", err)
	}
}
