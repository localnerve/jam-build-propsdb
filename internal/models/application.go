package models

import (
	"time"

	"gorm.io/datatypes"
)

// ApplicationDocument represents a document in the application scope
type ApplicationDocument struct {
	DocumentID      uint64 `gorm:"primaryKey;autoIncrement"`
	DocumentName    string `gorm:"uniqueIndex;size:255;not null"`
	DocumentVersion uint64 `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Collections     []ApplicationCollection `gorm:"many2many:application_documents_collections;"`
}

// ApplicationCollection represents a collection of properties
type ApplicationCollection struct {
	CollectionID   uint64 `gorm:"primaryKey;autoIncrement"`
	CollectionName string `gorm:"size:255;not null"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Properties     []ApplicationProperty `gorm:"many2many:application_collections_properties;"`
}

// ApplicationProperty represents a single property with a JSON value
type ApplicationProperty struct {
	PropertyID    uint64         `gorm:"primaryKey;autoIncrement"`
	PropertyName  string         `gorm:"size:255;not null"`
	PropertyValue datatypes.JSON `gorm:"type:json"`
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
