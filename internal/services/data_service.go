package services

import (
	"encoding/json"
	"fmt"

	"github.com/localnerve/propsdb/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DocumentResult represents the API output format
// Structure: { documentName: { "__version": "1", collectionName: { propName: propValue }}}
type DocumentResult map[string]interface{}

// CollectionInput represents input for upsert operations
type CollectionInput struct {
	Collection string                 `json:"collection"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// DeleteCollectionInput represents input for delete operations
type DeleteCollectionInput struct {
	Collection string   `json:"collection"`
	Properties []string `json:"properties,omitempty"`
}

// GetApplicationProperties retrieves properties for a specific document and collection
func GetApplicationProperties(db *gorm.DB, documentName, collectionName string) (DocumentResult, error) {
	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := db.Table("application_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN application_documents_collections dc ON d.document_id = dc.application_document_document_id").
		Joins("JOIN application_collections c ON dc.application_collection_collection_id = c.collection_id").
		Joins("JOIN application_collections_properties cp ON c.collection_id = cp.application_collection_collection_id").
		Joins("JOIN application_properties p ON cp.application_property_property_id = p.property_id").
		Where("d.document_name = ? AND c.collection_name = ?", documentName, collectionName).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// GetApplicationCollectionsAndProperties retrieves collections and properties for a document
func GetApplicationCollectionsAndProperties(db *gorm.DB, documentName string, collections []string) (DocumentResult, error) {
	query := db.Table("application_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN application_documents_collections dc ON d.document_id = dc.application_document_document_id").
		Joins("JOIN application_collections c ON dc.application_collection_collection_id = c.collection_id").
		Joins("LEFT JOIN application_collections_properties cp ON c.collection_id = cp.application_collection_collection_id").
		Joins("LEFT JOIN application_properties p ON cp.application_property_property_id = p.property_id").
		Where("d.document_name = ?", documentName)

	if len(collections) > 0 && collections[0] != "" {
		query = query.Where("c.collection_name IN ?", collections)
	}

	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// GetApplicationDocumentsCollectionsAndProperties retrieves all documents, collections, and properties
func GetApplicationDocumentsCollectionsAndProperties(db *gorm.DB) (DocumentResult, error) {
	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := db.Table("application_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN application_documents_collections dc ON d.document_id = dc.application_document_document_id").
		Joins("JOIN application_collections c ON dc.application_collection_collection_id = c.collection_id").
		Joins("LEFT JOIN application_collections_properties cp ON c.collection_id = cp.application_collection_collection_id").
		Joins("LEFT JOIN application_properties p ON cp.application_property_property_id = p.property_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// GetUserProperties retrieves properties for a specific user document and collection
func GetUserProperties(db *gorm.DB, userID, documentName, collectionName string) (DocumentResult, error) {
	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := db.Table("user_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN user_documents_collections dc ON d.document_id = dc.user_document_document_id").
		Joins("JOIN user_collections c ON dc.user_collection_collection_id = c.collection_id").
		Joins("JOIN user_collections_properties cp ON c.collection_id = cp.user_collection_collection_id").
		Joins("JOIN user_properties p ON cp.user_property_property_id = p.property_id").
		Where("d.user_id = ? AND d.document_name = ? AND c.collection_name = ?", userID, documentName, collectionName).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// GetUserCollectionsAndProperties retrieves collections and properties for a user document
func GetUserCollectionsAndProperties(db *gorm.DB, userID, documentName string, collections []string) (DocumentResult, error) {
	query := db.Table("user_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN user_documents_collections dc ON d.document_id = dc.user_document_document_id").
		Joins("JOIN user_collections c ON dc.user_collection_collection_id = c.collection_id").
		Joins("LEFT JOIN user_collections_properties cp ON c.collection_id = cp.user_collection_collection_id").
		Joins("LEFT JOIN user_properties p ON cp.user_property_property_id = p.property_id").
		Where("d.user_id = ? AND d.document_name = ?", userID, documentName)

	if len(collections) > 0 && collections[0] != "" {
		query = query.Where("c.collection_name IN ?", collections)
	}

	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// GetUserDocumentsCollectionsAndProperties retrieves all documents, collections, and properties for a user
func GetUserDocumentsCollectionsAndProperties(db *gorm.DB, userID string) (DocumentResult, error) {
	var results []struct {
		DocumentName    string
		DocumentVersion uint64
		CollectionName  string
		PropertyName    string
		PropertyValue   datatypes.JSON
	}

	err := db.Table("user_documents d").
		Select("d.document_name, d.document_version, c.collection_name, p.property_name, p.property_value").
		Joins("JOIN user_documents_collections dc ON d.document_id = dc.user_document_document_id").
		Joins("JOIN user_collections c ON dc.user_collection_collection_id = c.collection_id").
		Joins("LEFT JOIN user_collections_properties cp ON c.collection_id = cp.user_collection_collection_id").
		Joins("LEFT JOIN user_properties p ON cp.user_property_property_id = p.property_id").
		Where("d.user_id = ?", userID).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceResults(results), nil
}

// reduceResults converts database results to the API output format
func reduceResults(results []struct {
	DocumentName    string
	DocumentVersion uint64
	CollectionName  string
	PropertyName    string
	PropertyValue   datatypes.JSON
}) DocumentResult {
	output := make(DocumentResult)

	for _, row := range results {
		// Get or create document map
		var docMap map[string]interface{}
		if output[row.DocumentName] == nil {
			docMap = make(map[string]interface{})
			docMap["__version"] = fmt.Sprintf("%d", row.DocumentVersion)
			output[row.DocumentName] = docMap
		} else {
			docMap = output[row.DocumentName].(map[string]interface{})
		}

		// Get or create collection map
		var collMap map[string]interface{}
		if docMap[row.CollectionName] == nil {
			collMap = make(map[string]interface{})
			docMap[row.CollectionName] = collMap
		} else {
			collMap = docMap[row.CollectionName].(map[string]interface{})
		}

		// Add property if present
		if row.PropertyName != "" {
			var value interface{}
			if err := json.Unmarshal(row.PropertyValue, &value); err == nil {
				collMap[row.PropertyName] = value
			}
		}
	}

	return output
}

// SetApplicationProperties upserts application document with collections and properties
func SetApplicationProperties(db *gorm.DB, documentName string, version uint64, collections []CollectionInput) (uint64, int64, error) {
	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.ApplicationDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Document doesn't exist, version should be 0
				if version != 0 {
					return fmt.Errorf("E_VERSION")
				}
			} else {
				return err
			}
		} else {
			// Document exists, check version
			if doc.DocumentVersion != version {
				return fmt.Errorf("E_VERSION")
			}
		}

		// Insert or update document
		doc = models.ApplicationDocument{DocumentName: documentName}
		if err := tx.Where("document_name = ?", documentName).
			Assign(models.ApplicationDocument{DocumentName: documentName}).
			FirstOrCreate(&doc).Error; err != nil {
			return err
		}

		documentUpdated := false

		// Process collections
		for _, coll := range collections {
			var collection models.ApplicationCollection

			// Find or create collection
			if err := tx.Where("collection_name = ?", coll.Collection).
				FirstOrCreate(&collection, models.ApplicationCollection{CollectionName: coll.Collection}).Error; err != nil {
				return err
			}

			// Associate collection with document if not already associated
			var existingAssoc models.ApplicationDocument
			err := tx.Preload("Collections", "collection_id = ?", collection.CollectionID).
				Where("document_id = ?", doc.DocumentID).
				First(&existingAssoc).Error

			if err == gorm.ErrRecordNotFound || len(existingAssoc.Collections) == 0 {
				if err := tx.Model(&doc).Association("Collections").Append(&collection); err != nil {
					return err
				}
				documentUpdated = true
			}

			// Process properties
			for propName, propValue := range coll.Properties {
				jsonValue, err := json.Marshal(propValue)
				if err != nil {
					return err
				}

				var property models.ApplicationProperty

				// Check if property exists in this collection
				var existingProp models.ApplicationCollection
				err = tx.Preload("Properties", "property_name = ?", propName).
					Where("collection_id = ?", collection.CollectionID).
					First(&existingProp).Error

				if err == gorm.ErrRecordNotFound || len(existingProp.Properties) == 0 {
					// Create new property
					property = models.ApplicationProperty{
						PropertyName:  propName,
						PropertyValue: jsonValue,
					}
					if err := tx.Create(&property).Error; err != nil {
						return err
					}

					// Associate property with collection
					if err := tx.Model(&collection).Association("Properties").Append(&property); err != nil {
						return err
					}
					documentUpdated = true
				} else {
					// Property exists, check if value changed
					property = existingProp.Properties[0]
					if string(property.PropertyValue) != string(jsonValue) {
						if err := tx.Model(&property).Update("property_value", jsonValue).Error; err != nil {
							return err
						}
						documentUpdated = true
					}
				}
			}
		}

		// Update document version if changes were made
		if documentUpdated {
			newVersion = doc.DocumentVersion + 1
			result := tx.Model(&doc).Where("document_version = ?", doc.DocumentVersion).
				Update("document_version", newVersion)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("E_VERSION - Failed to update document due to concurrent modification")
			}
			affectedRows = result.RowsAffected
		} else {
			newVersion = doc.DocumentVersion
		}

		return nil
	})

	return newVersion, affectedRows, err
}

// SetUserProperties upserts user document with collections and properties
func SetUserProperties(db *gorm.DB, userID, documentName string, version uint64, collections []CollectionInput) (uint64, int64, error) {
	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.UserDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND document_name = ?", userID, documentName).
			First(&doc).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if version != 0 {
					return fmt.Errorf("E_VERSION")
				}
			} else {
				return err
			}
		} else {
			if doc.DocumentVersion != version {
				return fmt.Errorf("E_VERSION")
			}
		}

		// Insert or update document
		doc = models.UserDocument{UserID: userID, DocumentName: documentName}
		if err := tx.Where("user_id = ? AND document_name = ?", userID, documentName).
			Assign(models.UserDocument{UserID: userID, DocumentName: documentName}).
			FirstOrCreate(&doc).Error; err != nil {
			return err
		}

		documentUpdated := false

		// Process collections (similar to application logic)
		for _, coll := range collections {
			var collection models.UserCollection

			if err := tx.Where("collection_name = ?", coll.Collection).
				FirstOrCreate(&collection, models.UserCollection{CollectionName: coll.Collection}).Error; err != nil {
				return err
			}

			var count int64
			tx.Table("user_documents_collections").
				Where("document_id = ? AND collection_id = ?", doc.DocumentID, collection.CollectionID).
				Count(&count)

			if count == 0 {
				if err := tx.Exec("INSERT INTO user_documents_collections (document_id, collection_id) VALUES (?, ?)",
					doc.DocumentID, collection.CollectionID).Error; err != nil {
					return err
				}
				documentUpdated = true
			}

			for propName, propValue := range coll.Properties {
				jsonValue, err := json.Marshal(propValue)
				if err != nil {
					return err
				}

				var property models.UserProperty

				err = tx.Table("user_properties p").
					Select("p.*").
					Joins("JOIN user_collections_properties cp ON p.property_id = cp.property_id").
					Where("cp.collection_id = ? AND p.property_name = ?", collection.CollectionID, propName).
					First(&property).Error

				if err == gorm.ErrRecordNotFound {
					property = models.UserProperty{
						PropertyName:  propName,
						PropertyValue: jsonValue,
					}
					if err := tx.Create(&property).Error; err != nil {
						return err
					}

					if err := tx.Exec("INSERT INTO user_collections_properties (collection_id, property_id) VALUES (?, ?)",
						collection.CollectionID, property.PropertyID).Error; err != nil {
						return err
					}
					documentUpdated = true
				} else if err != nil {
					return err
				} else {
					if string(property.PropertyValue) != string(jsonValue) {
						if err := tx.Model(&property).Update("property_value", jsonValue).Error; err != nil {
							return err
						}
						documentUpdated = true
					}
				}
			}
		}

		if documentUpdated {
			newVersion = doc.DocumentVersion + 1
			result := tx.Model(&doc).Where("document_version = ?", doc.DocumentVersion).
				Update("document_version", newVersion)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("E_VERSION - Failed to update document due to concurrent modification")
			}
			affectedRows = result.RowsAffected
		} else {
			newVersion = doc.DocumentVersion
		}

		return nil
	})

	return newVersion, affectedRows, err
}
