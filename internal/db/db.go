package db

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gmonitor/internal/models"
	"gorm.io/gorm"
	"log"
	"time"
)

// Config holds database connection configuration
type Config struct {
	DSN          string
	MaxIdleConns int
	MaxOpenConns int
	Timeout      time.Duration
}

// NewConfig returns a new database configuration with default values
func NewConfig(dsn string) Config {
	return Config{
		DSN:          dsn,
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		Timeout:      10 * time.Second,
	}
}

// Connect initializes the database connection with optimized settings
func Connect(config Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(config.DSN), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database handle: %v", err)
	}
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.Timeout)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Database connected successfully")
	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database handle is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// Migrate applies database migrations with optimized settings
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Repository{},
		&models.Commit{},
	); err != nil {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}
	log.Println("Database migrations applied successfully")
	return nil
}
