package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// This script automatically creates test users without user interaction
// It's used by start-docker.ps1 to initialize the system

func main() {
	fmt.Println("=== Auto-creating Test Users ===")

	// Load environment
	envFile := ".env.development"
	if _, err := os.Stat(envFile); err != nil {
		envFile = ".env"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Could not load %s: %v\n", envFile, err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Connect to database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Test users to create
	testUsers := []struct {
		username string
		password string
		role     models.Role
	}{
		{
			username: "admin@asam.org",
			password: "admin123",
			role:     models.RoleAdmin,
		},
		{
			username: "user@asam.org",
			password: "admin123",
			role:     models.RoleUser,
		},
	}

	// Create or update each user
	for _, userData := range testUsers {
		if err := createOrUpdateUser(database, userData.username, userData.password, userData.role); err != nil {
			log.Printf("Error processing user %s: %v", userData.username, err)
		} else {
			fmt.Printf("✓ User %s ready\n", userData.username)
		}
	}

	fmt.Println("\n✅ Test users created successfully!")
	fmt.Println("You can login with:")
	fmt.Println("  Admin: admin@asam.org / admin123")
	fmt.Println("  User:  user@asam.org / admin123")
}

func createOrUpdateUser(db *gorm.DB, username, password string, role models.Role) error {
	var user models.User

	// Check if user exists
	err := db.Where("username = ?", username).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Username: username,
			Role:     role,
			IsActive: true,
		}

		// Set password using the model method
		if err := user.SetPassword(password); err != nil {
			return fmt.Errorf("failed to set password: %w", err)
		}

		// Create user
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("Created new user: %s\n", username)
	} else if err == nil {
		// Update existing user's password
		if err := user.SetPassword(password); err != nil {
			return fmt.Errorf("failed to set password: %w", err)
		}

		// Ensure user is active and has correct role
		user.IsActive = true
		user.Role = role

		// Save changes
		if err := db.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		fmt.Printf("Updated existing user: %s\n", username)
	} else {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}
