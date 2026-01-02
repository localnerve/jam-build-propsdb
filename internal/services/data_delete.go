package services

import (
	"fmt"

	"github.com/localnerve/propsdb/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DeleteApplicationCollection deletes a collection from an application document
func DeleteApplicationCollection(db *gorm.DB, documentName string, version uint64, collectionName string) (uint64, int64, error) {
	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.ApplicationDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		// Find collection
		var collection models.ApplicationCollection
		err := tx.Table("application_collections c").
			Select("c.*").
			Joins("JOIN application_documents_collections dc ON c.collection_id = dc.application_collection_collection_id").
			Where("dc.application_document_document_id = ? AND c.collection_name = ?", doc.DocumentID, collectionName).
			First(&collection).Error

		if err != nil {
			return fmt.Errorf("collection not found: %w", err)
		}

		// Remove association between document and collection using GORM
		if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
			return err
		}

		// Check if collection is orphaned (not associated with any other documents)
		var count int64
		tx.Model(&models.ApplicationDocument{}).
			Joins("JOIN application_documents_collections ON application_documents.document_id = application_documents_collections.application_document_document_id").
			Where("application_documents_collections.application_collection_collection_id = ?", collection.CollectionID).
			Count(&count)

		// If the collection is orphaned, delete its properties first, then the collection
		if count == 0 {
			// Delete all property associations for this collection
			if err := tx.Model(&collection).Association("Properties").Clear(); err != nil {
				return err
			}

			// Now delete the collection itself
			if err := tx.Delete(&collection).Error; err != nil {
				return err
			}
		}

		// Cleanup any remaining orphaned properties
		if err := cleanupApplicationOrphans(tx); err != nil {
			return err
		}

		// Update version
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

		return nil
	})

	return newVersion, affectedRows, err
}

// DeleteApplicationDocument deletes an entire application document
func DeleteApplicationDocument(db *gorm.DB, documentName string, version uint64) (uint64, int64, error) {
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.ApplicationDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		// Delete document (CASCADE will handle associations)
		result := tx.Delete(&doc)
		if result.Error != nil {
			return result.Error
		}
		affectedRows = result.RowsAffected

		// Cleanup orphaned collections and properties
		if err := cleanupApplicationOrphans(tx); err != nil {
			return err
		}

		return nil
	})

	return 0, affectedRows, err
}

// DeleteApplicationProperties deletes properties or collections from an application document
func DeleteApplicationProperties(db *gorm.DB, documentName string, version uint64, collections []DeleteCollectionInput, deleteDocument bool) (uint64, int64, error) {
	if deleteDocument {
		return DeleteApplicationDocument(db, documentName, version)
	}

	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.ApplicationDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		documentUpdated := false

		for _, coll := range collections {
			// Find collection
			var collection models.ApplicationCollection
			err := tx.Table("application_collections c").
				Select("c.*").
				Joins("JOIN application_documents_collections dc ON c.collection_id = dc.application_collection_collection_id").
				Where("dc.application_document_document_id = ? AND c.collection_name = ?", doc.DocumentID, coll.Collection).
				First(&collection).Error

			if err != nil {
				continue // Collection not found, skip
			}

			// If no properties specified, delete entire collection
			if len(coll.Properties) == 0 {
				if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
					return err
				}
				documentUpdated = true
			} else {
				// Delete specific properties
				for _, propName := range coll.Properties {
					var property models.ApplicationProperty
					err := tx.Table("application_properties p").
						Select("p.*").
						Joins("JOIN application_collections_properties cp ON p.property_id = cp.property_id").
						Where("cp.collection_id = ? AND p.property_name = ?", collection.CollectionID, propName).
						First(&property).Error

					if err == nil {
						if err := tx.Model(&collection).Association("Properties").Delete(&property); err != nil {
							return err
						}
						documentUpdated = true
					}
				}
			}
		}

		// Cleanup orphaned collections and properties
		if err := cleanupApplicationOrphans(tx); err != nil {
			return err
		}

		// Update version if changes were made
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

// DeleteUserCollection deletes a collection from a user document
func DeleteUserCollection(db *gorm.DB, userID, documentName string, version uint64, collectionName string) (uint64, int64, error) {
	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		var doc models.UserDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND document_name = ?", userID, documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		var collection models.UserCollection
		err := tx.Table("user_collections c").
			Select("c.*").
			Joins("JOIN user_documents_collections dc ON c.collection_id = dc.user_collection_collection_id").
			Where("dc.user_document_document_id = ? AND c.collection_name = ?", doc.DocumentID, collectionName).
			First(&collection).Error

		if err != nil {
			return fmt.Errorf("collection not found: %w", err)
		}

		if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
			return err
		}

		// Check if collection is orphaned
		var count int64
		tx.Model(&models.UserDocument{}).
			Joins("JOIN user_documents_collections ON user_documents.document_id = user_documents_collections.user_document_document_id").
			Where("user_documents_collections.user_collection_collection_id = ?", collection.CollectionID).
			Count(&count)

		// If orphaned, delete properties first, then collection
		if count == 0 {
			if err := tx.Model(&collection).Association("Properties").Clear(); err != nil {
				return err
			}
			if err := tx.Delete(&collection).Error; err != nil {
				return err
			}
		}

		if err := cleanupUserOrphans(tx); err != nil {
			return err
		}

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

		return nil
	})

	return newVersion, affectedRows, err
}

// DeleteUserDocument deletes an entire user document
func DeleteUserDocument(db *gorm.DB, userID, documentName string, version uint64) (uint64, int64, error) {
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		var doc models.UserDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND document_name = ?", userID, documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		result := tx.Delete(&doc)
		if result.Error != nil {
			return result.Error
		}
		affectedRows = result.RowsAffected

		if err := cleanupUserOrphans(tx); err != nil {
			return err
		}

		return nil
	})

	return 0, affectedRows, err
}

// DeleteUserProperties deletes properties or collections from a user document
func DeleteUserProperties(db *gorm.DB, userID, documentName string, version uint64, collections []DeleteCollectionInput, deleteDocument bool) (uint64, int64, error) {
	if deleteDocument {
		return DeleteUserDocument(db, userID, documentName, version)
	}

	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		var doc models.UserDocument
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND document_name = ?", userID, documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		documentUpdated := false

		for _, coll := range collections {
			var collection models.UserCollection
			err := tx.Table("user_collections c").
				Select("c.*").
				Joins("JOIN user_documents_collections dc ON c.collection_id = dc.user_collection_collection_id").
				Where("dc.user_document_document_id = ? AND c.collection_name = ?", doc.DocumentID, coll.Collection).
				First(&collection).Error

			if err != nil {
				continue
			}

			if len(coll.Properties) == 0 {
				if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
					return err
				}
				documentUpdated = true
			} else {
				for _, propName := range coll.Properties {
					var property models.UserProperty
					err := tx.Table("user_properties p").
						Select("p.*").
						Joins("JOIN user_collections_properties cp ON p.property_id = cp.property_id").
						Where("cp.collection_id = ? AND p.property_name = ?", collection.CollectionID, propName).
						First(&property).Error

					if err == nil {
						if err := tx.Model(&collection).Association("Properties").Delete(&property); err != nil {
							return err
						}
						documentUpdated = true
					}
				}
			}
		}

		if err := cleanupUserOrphans(tx); err != nil {
			return err
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

// cleanupApplicationOrphans removes orphaned collections and properties
func cleanupApplicationOrphans(tx *gorm.DB) error {
	// Delete collections not associated with any document
	if err := tx.Exec(`DELETE FROM application_collections 
		WHERE collection_id NOT IN (SELECT application_collection_collection_id FROM application_documents_collections)`).Error; err != nil {
		return err
	}

	// Delete collection-property associations for non-existent collections
	if err := tx.Exec(`DELETE FROM application_collections_properties 
		WHERE application_collection_collection_id NOT IN (SELECT application_collection_collection_id FROM application_documents_collections)`).Error; err != nil {
		return err
	}

	// Delete properties not associated with any collection
	if err := tx.Exec(`DELETE FROM application_properties 
		WHERE property_id NOT IN (SELECT application_property_property_id FROM application_collections_properties)`).Error; err != nil {
		return err
	}

	return nil
}

// cleanupUserOrphans removes orphaned user collections and properties
func cleanupUserOrphans(tx *gorm.DB) error {
	if err := tx.Exec(`DELETE FROM user_collections 
		WHERE collection_id NOT IN (SELECT user_collection_collection_id FROM user_documents_collections)`).Error; err != nil {
		return err
	}

	if err := tx.Exec(`DELETE FROM user_collections_properties 
		WHERE user_collection_collection_id NOT IN (SELECT user_collection_collection_id FROM user_documents_collections)`).Error; err != nil {
		return err
	}

	if err := tx.Exec(`DELETE FROM user_properties 
		WHERE property_id NOT IN (SELECT user_property_property_id FROM user_collections_properties)`).Error; err != nil {
		return err
	}

	return nil
}
