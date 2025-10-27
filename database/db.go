package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes a connection to the PostgreSQL database
func Connect() error {
	dbConnectString := os.Getenv("DB_CONNECT_STRING")
	if dbConnectString == "" {
		return fmt.Errorf("DB_CONNECT_STRING environment variable is required")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dbConnectString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("✓ Database connection established")
	return nil
}

// AutoMigrate runs automatic migrations for all models
func AutoMigrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&User{},
		&Source{},
		&EmailSent{},
		&EmailArticle{},
		&UserEmail{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("✓ Database migrations completed")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Close closes the database connection
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
