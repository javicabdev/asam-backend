package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/pkg/errors"
	appLogger "github.com/javicabdev/asam-backend/pkg/logger"
)

// dsnValueEscaper escapes a value for a single-quoted libpq keyword/value pair.
var dsnValueEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

// buildDSN assembles a libpq keyword/value connection string. Every value is
// single-quoted and escaped, so credentials containing spaces, quotes,
// backslashes or '=' reach the server verbatim instead of corrupting the DSN.
func buildDSN(host, user, password, dbName, port, sslMode string) string {
	pairs := [...]struct{ key, value string }{
		{"host", host},
		{"user", user},
		{"password", password},
		{"dbname", dbName},
		{"port", port},
		{"sslmode", sslMode},
	}
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, p.key+"='"+dsnValueEscaper.Replace(p.value)+"'")
	}
	return strings.Join(parts, " ")
}

// InitDB initializes the database connection with optional query monitoring.
// It takes a config.Config instance with loaded environment values and an optional logger.
func InitDB(cfg *config.Config, logs ...appLogger.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	var gormLogger logger.Interface

	// Check if a custom logger was provided for query monitoring
	if len(logs) > 0 && logs[0] != nil && cfg.LogSlowQueries {
		// Use our custom query monitor that integrates with the app logger
		gormLogger = NewQueryMonitor(logs[0], cfg.SlowQueryThreshold)
	} else {
		// Use default GORM logger
		gormLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Info, // Log level
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
				Colorful:                  true,        // Enable color
			},
		)
	}

	// Construct DSN (Data Source Name) using values in cfg
	dsn := buildDSN(cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	// Log the DSN with sensitive info masked for debugging
	maskedDSN := fmt.Sprintf("host=%s user=%s password=***** dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	if len(logs) > 0 && logs[0] != nil {
		logs[0].Info(fmt.Sprintf("Connecting to database: %s", maskedDSN))
	} else {
		log.Printf("Connecting to database: %s", maskedDSN)
	}

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
		PrepareStmt: true, // Enable prepared statement cache for better performance
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

	// Reconfigure with optimized settings
	sqlDB, err = db.DB()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "Failed to get database instance")
	}

	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)       // Keep idle connections
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)       // Maximum simultaneous connections
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime) // Connection lifetime

	// Test connection with timeout for Cloud Run
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to ping database server")
	}

	if len(logs) > 0 && logs[0] != nil {
		logs[0].Info(fmt.Sprintf("Successfully connected to database %s at %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort))
	} else {
		log.Printf("Successfully connected to database %s at %s:%s", cfg.DBName, cfg.DBHost, cfg.DBPort)
	}
	return db, nil
}
