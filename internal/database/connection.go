// connection.go
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

package database

import (
	"fmt"
	"log"

	"github.com/localnerve/jam-build-propsdb/internal/config"
	"github.com/localnerve/jam-build-propsdb/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a database connection based on the configured DB_TYPE
func Connect(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "mysql", "mariadb":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBAppUser,
			cfg.DBAppPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBAppDatabase,
		)
		dialector = mysql.Open(dsn)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost,
			cfg.DBAppUser,
			cfg.DBAppPassword,
			cfg.DBAppDatabase,
			cfg.DBPort,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		// For SQLite, DBAppDatabase is the file path
		dialector = sqlite.Open(cfg.DBAppDatabase)

	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.DBAppUser,
			cfg.DBAppPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBAppDatabase,
		)
		dialector = sqlserver.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.DBAppConnectionLimit)
	sqlDB.SetMaxIdleConns(cfg.DBAppConnectionLimit / 2)

	log.Printf("Connected to %s database: %s", cfg.DBType, cfg.DBAppDatabase)

	return db, nil
}

// ConnectUser establishes a user database connection (with different credentials)
func ConnectUser(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch cfg.DBType {
	case "mysql", "mariadb":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBAppDatabase,
		)
		dialector = mysql.Open(dsn)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBAppDatabase,
			cfg.DBPort,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		// For SQLite, use the same connection (no separate user credentials)
		dialector = sqlite.Open(cfg.DBAppDatabase)

	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBAppDatabase,
		)
		dialector = sqlserver.Open(dsn)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user database: %w", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.DBConnectionLimit)
	sqlDB.SetMaxIdleConns(cfg.DBConnectionLimit / 2)

	log.Printf("Connected to %s user database: %s", cfg.DBType, cfg.DBAppDatabase)

	return db, nil
}

// AutoMigrate runs automatic migrations for all models
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.ApplicationDocument{},
		&models.ApplicationCollection{},
		&models.ApplicationProperty{},
		&models.UserDocument{},
		&models.UserCollection{},
		&models.UserProperty{},
	)
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
