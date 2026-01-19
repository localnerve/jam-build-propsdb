package models

import (
	"database/sql/driver"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// JSON is a wrapper around gorm.io/datatypes.JSON to allow for custom data type mapping
type JSON struct {
	datatypes.JSON
}

// Value promotes the embedded JSON's Value method
func (j JSON) Value() (driver.Value, error) {
	return j.JSON.Value()
}

// Scan promotes the embedded JSON's Scan method
func (j *JSON) Scan(value interface{}) error {
	return j.JSON.Scan(value)
}

// GormDBDataType ensures the correct data type is used for each database driver.
// This resolves the issue where MSSQL does not support the 'json' data type.
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "JSON"
	case "postgres":
		return "JSONB"
	case "sqlserver", "mssql":
		return "NVARCHAR(MAX)"
	case "sqlite":
		return "JSON"
	}
	return "TEXT"
}
