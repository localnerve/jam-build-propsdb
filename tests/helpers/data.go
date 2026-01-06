package helpers

import (
	"encoding/json"
	"testing"

	"github.com/localnerve/propsdb/internal/models"
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
