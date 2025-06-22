package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Try to load .env.production file
	if err := godotenv.Load(".env.production"); err != nil {
		log.Println("No .env.production file found, using environment variables")
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "asam"),
		getEnv("DB_SSL_MODE", "require"),
	)

	fmt.Println("Testing database connection...")
	fmt.Printf("Host: %s\n", getEnv("DB_HOST", "localhost"))
	fmt.Printf("Port: %s\n", getEnv("DB_PORT", "5432"))
	fmt.Printf("Database: %s\n", getEnv("DB_NAME", "asam"))
	fmt.Printf("SSL Mode: %s\n", getEnv("DB_SSL_MODE", "require"))

	// Try to connect
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Ping the database
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Successfully connected to database!")

	// Check if migrations have been run
	var tableCount int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount)
	fmt.Printf("📊 Found %d tables in the database\n", tableCount)

	// List tables
	var tables []string
	db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name").Scan(&tables)
	if len(tables) > 0 {
		fmt.Println("📋 Tables found:")
		for _, table := range tables {
			fmt.Printf("   - %s\n", table)
		}
	}

	// Check for specific required tables
	requiredTables := []string{"users", "members", "families", "payments", "refresh_tokens", "verification_tokens"}
	missingTables := []string{}

	for _, table := range requiredTables {
		var exists bool
		db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", table).Scan(&exists)
		if !exists {
			missingTables = append(missingTables, table)
		}
	}

	if len(missingTables) > 0 {
		fmt.Println("⚠️  Missing required tables:")
		for _, table := range missingTables {
			fmt.Printf("   - %s\n", table)
		}
		fmt.Println("\n💡 Run migrations with: go run cmd/migrate/main.go -cmd up")
	} else {
		fmt.Println("✅ All required tables exist!")
	}

	// Close connection
	sqlDB.Close()
	fmt.Println("\n🎉 Database connection test completed successfully!")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
