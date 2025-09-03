package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresConnection creates a new PostgreSQL connection with GORM
func NewPostgresConnection(cfg DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// GORM configuration
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), 
	}

	// Open connection with retry logic
	var db *gorm.DB
	var err error

	maxRetries := 5
	for i := range maxRetries {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			break
		}

		log.Printf("Failed to connect to PostgreSQL (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries - 1 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDBConnectionFailed, err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("Successfully connected to PostgreSQL database: %s", cfg.DBName)
	return db, nil
}

// HealthCheck verifies the database connection is still alive
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// CloseDatabaseConnection gracefully closes the database connection
func CloseDatabaseConnection(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get PostgreSQL db: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Failed to close database connection: %v", err)
	} else {
		log.Println("Database connection closed gracefully")
	}
}

// AutoMigrate handles database migrations for given models
func AutoMigrate(db *gorm.DB, models []any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	log.Printf("Database migrations completed for %d models", len(models))
	return nil
}
