// application.go
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

// ApplicationDocument represents a document in the application scope
type ApplicationDocument struct {
	DocumentID      uint64 `gorm:"primaryKey;autoIncrement"`
	DocumentName    string `gorm:"uniqueIndex;size:255;not null"`
	DocumentVersion uint64 `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Collections     []ApplicationCollection `gorm:"many2many:application_documents_collections;joinForeignKey:document_id;joinReferences:collection_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// ApplicationCollection represents a collection of properties
type ApplicationCollection struct {
	CollectionID   uint64 `gorm:"primaryKey;autoIncrement"`
	CollectionName string `gorm:"size:255;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Properties     []ApplicationProperty `gorm:"many2many:application_collections_properties;joinForeignKey:collection_id;joinReferences:property_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// ApplicationProperty represents a single property with a JSON value
type ApplicationProperty struct {
	PropertyID    uint64 `gorm:"primaryKey;autoIncrement"`
	PropertyName  string `gorm:"size:255;not null"`
	PropertyValue JSON
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName overrides the table name for ApplicationDocument
func (ApplicationDocument) TableName() string {
	return "application_documents"
}

// TableName overrides the table name for ApplicationCollection
func (ApplicationCollection) TableName() string {
	return "application_collections"
}

// TableName overrides the table name for ApplicationProperty
func (ApplicationProperty) TableName() string {
	return "application_properties"
}
