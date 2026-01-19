// user.go
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

package models

import (
	"time"
)

// UserDocument represents a document in the user scope
type UserDocument struct {
	DocumentID      uint64 `gorm:"primaryKey;autoIncrement"`
	UserID          string `gorm:"type:char(36);not null;index:idx_user_document,unique"`
	DocumentName    string `gorm:"size:255;not null;index:idx_user_document,unique"`
	DocumentVersion uint64 `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Collections     []UserCollection `gorm:"many2many:user_documents_collections;joinForeignKey:document_id;joinReferences:collection_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// UserCollection represents a collection of properties for users
type UserCollection struct {
	CollectionID   uint64 `gorm:"primaryKey;autoIncrement"`
	CollectionName string `gorm:"size:255;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Properties     []UserProperty `gorm:"many2many:user_collections_properties;joinForeignKey:collection_id;joinReferences:property_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// UserProperty represents a single property with a JSON value for users
type UserProperty struct {
	PropertyID    uint64 `gorm:"primaryKey;autoIncrement"`
	PropertyName  string `gorm:"size:255;not null"`
	PropertyValue JSON
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName overrides the table name for UserDocument
func (UserDocument) TableName() string {
	return "user_documents"
}

// TableName overrides the table name for UserCollection
func (UserCollection) TableName() string {
	return "user_collections"
}

// TableName overrides the table name for UserProperty
func (UserProperty) TableName() string {
	return "user_properties"
}
