package database

import (
	"fmt"
	"log"

	"github.com/localnerve/propsdb/internal/config"
	"github.com/localnerve/propsdb/internal/models"
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
			cfg.DBDatabase,
		)
		dialector = mysql.Open(dsn)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost,
			cfg.DBAppUser,
			cfg.DBAppPassword,
			cfg.DBDatabase,
			cfg.DBPort,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		// For SQLite, DBDatabase is the file path
		dialector = sqlite.Open(cfg.DBDatabase)

	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.DBAppUser,
			cfg.DBAppPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBDatabase,
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

	log.Printf("Connected to %s database: %s", cfg.DBType, cfg.DBDatabase)

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
			cfg.DBDatabase,
		)
		dialector = mysql.Open(dsn)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost,
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBDatabase,
			cfg.DBPort,
		)
		dialector = postgres.Open(dsn)

	case "sqlite":
		// For SQLite, use the same connection (no separate user credentials)
		dialector = sqlite.Open(cfg.DBDatabase)

	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBDatabase,
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

	log.Printf("Connected to %s user database: %s", cfg.DBType, cfg.DBDatabase)

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
