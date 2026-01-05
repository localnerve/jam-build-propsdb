package models

import (
	"time"

	"gorm.io/datatypes"
)

// UserDocument represents a document in the user scope
type UserDocument struct {
	DocumentID      uint64 `gorm:"primaryKey;autoIncrement"`
	UserID          string `gorm:"type:char(36);not null;index:idx_user_document,unique"`
	DocumentName    string `gorm:"size:255;not null;index:idx_user_document,unique"`
	DocumentVersion uint64 `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Collections     []UserCollection `gorm:"many2many:user_documents_collections;joinForeignKey:document_id;joinReferences:collection_id"`
}

// UserCollection represents a collection of properties for users
type UserCollection struct {
	CollectionID   uint64 `gorm:"primaryKey;autoIncrement"`
	CollectionName string `gorm:"size:255;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Properties     []UserProperty `gorm:"many2many:user_collections_properties;joinForeignKey:collection_id;joinReferences:property_id"`
}

// UserProperty represents a single property with a JSON value for users
type UserProperty struct {
	PropertyID    uint64         `gorm:"primaryKey;autoIncrement"`
	PropertyName  string         `gorm:"size:255;not null"`
	PropertyValue datatypes.JSON `gorm:"type:json"`
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
