package main

import (
	"fmt"
	"log"

	"github.com/localnerve/propsdb/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Auto-migrate to see what GORM creates
	err = db.AutoMigrate(
		&models.ApplicationDocument{},
		&models.ApplicationCollection{},
		&models.ApplicationProperty{},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Get the schema
	var tables []string
	db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables)

	for _, table := range tables {
		fmt.Printf("\n=== Table: %s ===\n", table)
		var schema string
		db.Raw(fmt.Sprintf("SELECT sql FROM sqlite_master WHERE name='%s'", table)).Scan(&schema)
		fmt.Println(schema)
	}
}
