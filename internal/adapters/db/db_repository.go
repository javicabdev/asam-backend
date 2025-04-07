package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection.
// It takes a config.Config instance with loaded environment values.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
			Colorful:                  true,        // Enable color
		},
	)

	// Construct DSN (Data Source Name) using values in cfg
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	// Log the DSN with sensitive info masked for debugging
	maskedDSN := fmt.Sprintf("host=%s user=%s password=***** dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	log.Printf("Connecting to database: %s", maskedDSN)

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
	})
	if err != nil {
		return nil, errors.DB(err, "Failed to connect to database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "Failed to get database instance")
	}

	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)       // Keep idle connections
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)       // Maximum simultaneous connections
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime) // Connection lifetime

	// Configure statement cache
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Enable prepared statement cache
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
		Logger: logger.Default.LogMode(logger.Silent), // Reduce logging in production
	})
	if err != nil {
		return nil, errors.DB(err, "Failed to reopen database with prepared statements")
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to ping database server")
	}

	log.Printf("Successfully connected to database %s at %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort)
	return db, nil
}

// Additional helper functions for database operations

// IsConnected checks if the database connection is still active
func IsConnected(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Failed to get underlying database connection")
	}

	if err := sqlDB.Ping(); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "Database connection lost")
	}

	return nil
}

// GetDBStats returns statistics about the database connection pool
func GetDBStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to get underlying database connection")
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration,
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}
