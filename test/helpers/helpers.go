package helpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// SetupTestDB creates and returns a test database connection
// It initializes a PostgreSQL database for testing and migrates the necessary tables
func SetupTestDB(t *testing.T) *gorm.DB {
	// Connection string for test database
	// In a real implementation, these values would come from environment variables or test config
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Madrid",
		"localhost", "postgres", "postgres", "asam_test", "5432",
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

	return db
}
