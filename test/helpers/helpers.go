// Package helpers provides test helper functions
package helpers

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// getEnvWithDefault returns the value of the environment variable or the default value if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// SetupTestDB creates and returns a test database connection
// It initializes a PostgreSQL database for testing and migrates the necessary tables
func SetupTestDB(t *testing.T) *gorm.DB {
	// Get database configuration from environment variables with fallbacks
	host := getEnvWithDefault("DB_HOST", "localhost")
	user := getEnvWithDefault("DB_USER", "postgres")
	password := getEnvWithDefault("DB_PASSWORD", "postgres")
	dbname := getEnvWithDefault("DB_NAME", "asam_test_db")
	port := getEnvWithDefault("DB_PORT", "5432")

	// Connection string for test database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Madrid",
		host, user, password, dbname, port,
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent mode for tests
	})
	require.NoError(t, err)

	// Clean up existing data
	db.Exec("TRUNCATE TABLE refresh_tokens CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")

	// Migrate the necessary tables
	err = db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		// Add other models as needed
	)
	require.NoError(t, err)

	// Ensure the unique constraint on refresh_tokens.uuid has the expected name
	// This is needed because GORM's AutoMigrate might create a constraint with a different name
	db.Exec("ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS uni_refresh_tokens_uuid")
	db.Exec("ALTER TABLE refresh_tokens ADD CONSTRAINT uni_refresh_tokens_uuid UNIQUE (uuid)")

	return db
}
