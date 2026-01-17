// data_delete.go
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

package services

import (
	"fmt"

	"github.com/localnerve/jam-build-propsdb/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// DeleteApplicationCollection deletes a collection from an application document
func DeleteApplicationCollection(db *gorm.DB, documentName string, version uint64, collectionName string) (uint64, int64, error) {
	var newVersion uint64
	var affectedRows int64

	err := db.Transaction(func(tx *gorm.DB) error {
		// Lock and check version
		var doc models.ApplicationDocument
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		// Find collection associated with this document
		var collection models.ApplicationCollection
		err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Model(&doc).Where("collection_name = ?", collectionName).Association("Collections").Find(&collection)

		if err != nil {
			return fmt.Errorf("collection not found: %w", err) // GORM association find returns error? Usually nil if empty?
		}
		// Check if it was actually found (GORM Find might not error on empty result for associations depending on usage,
		// but Model Association Find usually fills struct or slice. If ID is 0, it wasn't validly found).
		if collection.CollectionID == 0 {
			// Try to find it explicitly to confirm error or just return not found
			return fmt.Errorf("collection not found")
		}

		// Remove association between document and collection using GORM
		if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
			return err
		}

		// Check if collection is orphaned (not associated with any other documents)
		var count int64
		// We need to count associations for this collection using the join table
		// GORM standard way:
		if err := tx.Table("application_documents_collections").
			Where("collection_id = ?", collection.CollectionID).
			Count(&count).Error; err != nil {
			return err
		}

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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("document_name = ?", documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		documentUpdated := false

		for _, coll := range collections {
			// Find collection for this document
			var collection models.ApplicationCollection
			err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
				Model(&doc).Where("collection_name = ?", coll.Collection).Association("Collections").Find(&collection)

			if err != nil || collection.CollectionID == 0 {
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
					// Find property in this collection
					err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
						Model(&collection).Where("property_name = ?", propName).Association("Properties").Find(&property)

					if err == nil && property.PropertyID != 0 {
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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND document_name = ?", userID, documentName).
			First(&doc).Error; err != nil {
			return err
		}

		if doc.DocumentVersion != version {
			return fmt.Errorf("E_VERSION")
		}

		var collection models.UserCollection
		err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Model(&doc).Where("collection_name = ?", collectionName).Association("Collections").Find(&collection)

		if err != nil {
			return fmt.Errorf("collection not found: %w", err)
		}
		if collection.CollectionID == 0 {
			return fmt.Errorf("collection not found")
		}

		if err := tx.Model(&doc).Association("Collections").Delete(&collection); err != nil {
			return err
		}

		// Check if collection is orphaned
		var count int64
		if err := tx.Table("user_documents_collections").
			Where("collection_id = ?", collection.CollectionID).
			Count(&count).Error; err != nil {
			return err
		}

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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
		if err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
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
			err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
				Model(&doc).Where("collection_name = ?", coll.Collection).Association("Collections").Find(&collection)

			if err != nil || collection.CollectionID == 0 {
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
					err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
						Model(&collection).Where("property_name = ?", propName).Association("Properties").Find(&property)

					if err == nil && property.PropertyID != 0 {
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
	// Simple approach: Use NOT IN subqueries but with mapped table names if possible.
	// However, GORM doesn't make subqueries on many2many join tables very elegant without raw SQL.
	// But we can use the table names from the models if we really want to be safe,
	// OR we can rely on standard "DELETE FROM [collection_table] WHERE id NOT IN (SELECT ...)"

	// The problem described by the user was specific column names: `application_document_document_id`.
	// The GORM default M2M table columns might be different.
	// e.g. `application_document_id` vs `application_document_document_id`.

	// Since we defined the join table in simpler terms in `data_service.go` logic before?
	// The User error message showed: "Unknown column 'dc.application_document_document_id' in 'WHERE'"

	// We should inspect `models/application.go` again to see the `gorm:"many2many:..."` tag.
	// It was: `gorm:"many2many:application_documents_collections;joinForeignKey:document_id;joinReferences:collection_id"`
	// So the columns are likely `document_id` and `collection_id` in the join table `application_documents_collections`.
	// The OLD code was expecting `application_document_document_id`. This confirms the schema change hypothesis.

	// So for orphan cleanup, we need to use the CORRECT column names.
	// `application_documents_collections` table likely has `collection_id` (ref to collection) and `document_id` (ref to doc).

	// Delete collections not associated with any document
	if err := tx.Exec(`DELETE FROM application_collections 
		WHERE collection_id NOT IN (SELECT collection_id FROM application_documents_collections)`).Error; err != nil {
		return err
	}

	// Delete collection-property associations for non-existent collections
	// (This table `application_collections_properties` connects collection <-> property)
	// Tag in model: `gorm:"many2many:application_collections_properties;joinForeignKey:collection_id;joinReferences:property_id"`
	// So columns are `collection_id` and `property_id`.
	// Wait, if we delete the collection, the M2M association rows might remain if not cascaded?

	if err := tx.Exec(`DELETE FROM application_collections_properties 
		WHERE collection_id NOT IN (SELECT collection_id FROM application_collections)`).Error; err != nil {
		return err
	}

	// Delete properties not associated with any collection
	if err := tx.Exec(`DELETE FROM application_properties 
		WHERE property_id NOT IN (SELECT property_id FROM application_collections_properties)`).Error; err != nil {
		return err
	}

	return nil
}

// cleanupUserOrphans removes orphaned user collections and properties
func cleanupUserOrphans(tx *gorm.DB) error {
	// UserDocument -> Collections: `user_documents_collections` (`document_id`, `collection_id`)
	// UserCollection -> Properties: `user_collections_properties` (`collection_id`, `property_id`)

	if err := tx.Exec(`DELETE FROM user_collections 
		WHERE collection_id NOT IN (SELECT collection_id FROM user_documents_collections)`).Error; err != nil {
		return err
	}

	if err := tx.Exec(`DELETE FROM user_collections_properties 
		WHERE collection_id NOT IN (SELECT collection_id FROM user_collections)`).Error; err != nil {
		return err
	}

	if err := tx.Exec(`DELETE FROM user_properties 
		WHERE property_id NOT IN (SELECT property_id FROM user_collections_properties)`).Error; err != nil {
		return err
	}

	return nil
}
