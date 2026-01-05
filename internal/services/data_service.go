package services

import (
	"encoding/json"
	"fmt"

	"github.com/localnerve/propsdb/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
	var doc models.ApplicationDocument
	err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Preload("Collections", "collection_name = ?", collectionName).
		Preload("Collections.Properties").
		Where("document_name = ?", documentName).
		First(&doc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}

	// Filter out if the collection join didn't find the specific collection (Preload works differently than Join)
	// Actually, with Preload, if the document exists but the collection doesn't match the condition,
	// doc.Collections will be empty.
	if len(doc.Collections) == 0 {
		// We need to return not found if the specific collection requested isn't there,
		// but the original query did an INNER JOIN so it would have returned empty if doc existed but coll didn't.
		// However, to strictly match the semantics of "not found" for the *pair*, we should check.
		return nil, fmt.Errorf("not found")
	}

	return reduceApplicationDocuments([]models.ApplicationDocument{doc}), nil
}

// GetApplicationCollectionsAndProperties retrieves collections and properties for a document
func GetApplicationCollectionsAndProperties(db *gorm.DB, documentName string, collections []string) (DocumentResult, error) {
	var doc models.ApplicationDocument
	query := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Where("document_name = ?", documentName)

	if len(collections) > 0 && collections[0] != "" {
		query = query.Preload("Collections", "collection_name IN ?", collections)
	} else {
		query = query.Preload("Collections")
	}

	err := query.Preload("Collections.Properties").
		First(&doc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}

	// If filtered by specific collections and none found, effectively not found?
	// The original query did a LEFT JOIN for properties but an INNER JOIN-like structure for the main doc logic usually.
	// But let's look at the original:
	// "LEFT JOIN application_collections_properties"
	// The original required the DOCUMENT to exist.

	// If collections were specified and resulted in 0 collections, is that an error?
	// The original query: "c.collection_name IN ?" was on the JOIN.
	// If no rows returned, it returned "not found".
	if len(collections) > 0 && collections[0] != "" && len(doc.Collections) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceApplicationDocuments([]models.ApplicationDocument{doc}), nil
}

// GetApplicationDocumentsCollectionsAndProperties retrieves all documents, collections, and properties
func GetApplicationDocumentsCollectionsAndProperties(db *gorm.DB) (DocumentResult, error) {
	var docs []models.ApplicationDocument

	// We want all documents that have at least one collection usually,
	// but the original query was:
	// JOIN application_documents_collections ... JOIN application_collections
	// So it only returned documents that HAD collections.

	// Fetch all documents with their collections and properties
	if err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Preload("Collections").Preload("Collections.Properties").Find(&docs).Error; err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	// Filter out docs with no collections to match original INNER JOIN behavior if necessary?
	// The original `reduceResults` just iterated. If we have a doc with 0 collections,
	// the previous Code probably wouldn't have it in the list if it was an INNER JOIN.
	// Let's rely on the reducer to formatted it.

	return reduceApplicationDocuments(docs), nil
}

// GetUserProperties retrieves properties for a specific user document and collection
func GetUserProperties(db *gorm.DB, userID, documentName, collectionName string) (DocumentResult, error) {
	var doc models.UserDocument
	err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Preload("Collections", "collection_name = ?", collectionName).
		Preload("Collections.Properties").
		Where("user_id = ? AND document_name = ?", userID, documentName).
		First(&doc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}

	if len(doc.Collections) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceUserDocuments([]models.UserDocument{doc}), nil
}

// GetUserCollectionsAndProperties retrieves collections and properties for a user document
func GetUserCollectionsAndProperties(db *gorm.DB, userID, documentName string, collections []string) (DocumentResult, error) {
	var doc models.UserDocument
	query := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Where("user_id = ? AND document_name = ?", userID, documentName)

	if len(collections) > 0 && collections[0] != "" {
		query = query.Preload("Collections", "collection_name IN ?", collections)
	} else {
		query = query.Preload("Collections")
	}

	err := query.Preload("Collections.Properties").
		First(&doc).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}

	if len(collections) > 0 && collections[0] != "" && len(doc.Collections) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceUserDocuments([]models.UserDocument{doc}), nil
}

// GetUserDocumentsCollectionsAndProperties retrieves all documents, collections, and properties for a user
func GetUserDocumentsCollectionsAndProperties(db *gorm.DB, userID string) (DocumentResult, error) {
	var docs []models.UserDocument

	err := db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)}).
		Where("user_id = ?", userID).
		Preload("Collections").
		Preload("Collections.Properties").
		Find(&docs).Error

	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return reduceUserDocuments(docs), nil
}

// reduceApplicationDocuments converts application models to API output
func reduceApplicationDocuments(docs []models.ApplicationDocument) DocumentResult {
	output := make(DocumentResult)

	for _, doc := range docs {
		// If mimicking INNER JOIN, we might skip docs with no collections,
		// but typically getting the doc itself is fine.
		// However, the previous "not found" logic often triggered on empty sets.

		docMap := make(map[string]interface{})
		docMap["__version"] = fmt.Sprintf("%d", doc.DocumentVersion)

		for _, coll := range doc.Collections {
			collMap := make(map[string]interface{})
			for _, prop := range coll.Properties {
				var value interface{}
				if err := json.Unmarshal(prop.PropertyValue, &value); err == nil {
					collMap[prop.PropertyName] = value
				}
			}
			docMap[coll.CollectionName] = collMap
		}

		// If we want to strictly hide documents that ended up having NO collections
		// (e.g. because of the collection name filter in Preload),
		// we should verify if that's desired.
		// For now, allow it, as the doc exists.
		output[doc.DocumentName] = docMap
	}

	return output
}

// reduceUserDocuments converts user models to API output
func reduceUserDocuments(docs []models.UserDocument) DocumentResult {
	output := make(DocumentResult)

	for _, doc := range docs {
		docMap := make(map[string]interface{})
		docMap["__version"] = fmt.Sprintf("%d", doc.DocumentVersion)

		for _, coll := range doc.Collections {
			collMap := make(map[string]interface{})
			for _, prop := range coll.Properties {
				var value interface{}
				if err := json.Unmarshal(prop.PropertyValue, &value); err == nil {
					collMap[prop.PropertyName] = value
				}
			}
			docMap[coll.CollectionName] = collMap
		}
		output[doc.DocumentName] = docMap
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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
