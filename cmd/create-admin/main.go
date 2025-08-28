// Package main provides a command to create admin users securely
// reading credentials from environment variables
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// Command line flags
var (
	environment string
	envFile     string
	forceUpdate bool
)

// Environment file constants
const (
	LocalEnvFile = ".env.development"
	AivenEnvFile = ".env.aiven"
)

func main() {
	// Setup command line flags
	flag.StringVar(&environment, "env", "local", "Environment to use (local, aiven)")
	flag.StringVar(&envFile, "envfile", "", "Custom environment file path (overrides -env)")
	flag.BoolVar(&forceUpdate, "force", false, "Force update existing admin user if exists")
	flag.Parse()

	// Determine which env file to use
	if envFile == "" {
		switch strings.ToLower(environment) {
		case constants.EnvLocal:
			envFile = LocalEnvFile
		case constants.EnvAiven:
			envFile = AivenEnvFile
		default:
			log.Fatalf("Invalid environment '%s'. Must be 'local' or 'aiven'", environment)
		}
	}

	// Load environment variables
	log.Printf("Loading environment from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Error loading %s file: %v", envFile, err)
	}

	// Get admin credentials from environment
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminUsername := os.Getenv("ADMIN_USERNAME")

	// Validate required environment variables
	if adminEmail == "" {
		log.Fatal("ADMIN_EMAIL environment variable is required. Set it to the admin's email address.")
	}
	if adminPassword == "" {
		log.Fatal("ADMIN_PASSWORD environment variable is required. Must be at least 8 characters with uppercase, lowercase, and number.")
	}

	// If username not provided, use email as username
	if adminUsername == "" {
		adminUsername = adminEmail
		log.Printf("ADMIN_USERNAME not provided, using email as username: %s", adminUsername)
	} else {
		log.Printf("Using custom username: %s (email: %s)", adminUsername, adminEmail)
	}

	// Validate password strength
	if len(adminPassword) < 8 {
		log.Fatal("ADMIN_PASSWORD must be at least 8 characters long")
	}

	// Connect to database
	dbInstance, err := connectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB: %v", err)
	}
	defer sqlDB.Close()

	// Create repositories
	userRepo := db.NewUserRepository(dbInstance)
	memberRepo := db.NewMemberRepository(dbInstance)
	tokenRepo := db.NewVerificationTokenRepository(dbInstance)

	// Create logger
	appLogger, err := logger.InitLogger(logger.Config{
		Level:       logger.InfoLevel,
		Development: true,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Create email service (stub for this command)
	emailService := &stubEmailService{}

	// Create user service
	userService := services.NewUserService(
		userRepo,
		memberRepo,
		tokenRepo,
		emailService,
		appLogger,
		os.Getenv("BASE_URL"),
	)

	ctx := context.Background()

	// Check if user already exists
	existingUser, _ := userRepo.FindByUsername(ctx, adminUsername)

	if existingUser != nil {
		if !forceUpdate {
			log.Printf("Admin user '%s' already exists. Use -force flag to update password.", adminUsername)
			os.Exit(0)
		}

		log.Printf("Updating existing admin user '%s'...", adminUsername)

		// Update existing user
		updates := map[string]interface{}{
			"password": adminPassword,
			"email":    adminEmail,
			"role":     models.RoleAdmin,
			"isActive": true,
		}

		_, err = userService.UpdateUser(ctx, existingUser.ID, updates)
		if err != nil {
			log.Fatalf("Failed to update admin user: %v", err)
		}

		// Reset email verification status directly in database
		existingUser.EmailVerified = false
		existingUser.EmailVerifiedAt = nil
		if err := userRepo.Update(ctx, existingUser); err != nil {
			log.Fatalf("Failed to reset email verification: %v", err)
		}

		log.Println("================================")
		log.Println("Admin user updated successfully!")
		log.Printf("Username: %s", adminUsername)
		log.Printf("Email: %s", adminEmail)
		log.Println("Email verified: false")
		log.Println("The user must verify their email on first login")
		log.Println("================================")
	} else {
		log.Printf("Creating new admin user '%s'...", adminUsername)

		// Create new admin user
		user, err := userService.CreateUser(
			ctx,
			adminUsername,
			adminEmail,
			adminPassword,
			models.RoleAdmin,
			nil, // No member association for admin
		)
		if err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		log.Println("================================")
		log.Println("Admin user created successfully!")
		log.Printf("ID: %d", user.ID)
		log.Printf("Username: %s", user.Username)
		log.Printf("Email: %s", user.Email)
		log.Printf("Role: %s", user.Role)
		log.Printf("Email verified: %v", user.EmailVerified)
		log.Println("The user must verify their email on first login")
		log.Println("================================")
	}

	// Security reminder
	log.Println("\nSECURITY REMINDERS:")
	log.Println("1. Never commit .env files with real credentials")
	log.Println("2. Use strong passwords in production")
	log.Println("3. Enable 2FA when possible")
	log.Println("4. Regularly rotate admin credentials")
}

// connectDatabase establishes a connection to the database
func connectDatabase() (*gorm.DB, error) {
	// Build DSN from environment variables
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"),
	)

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// stubEmailService is a no-op email service for this command
type stubEmailService struct{}

func (s *stubEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	log.Printf("Email service stub: Would send email to %s with subject: %s", to, subject)
	return nil
}

func (s *stubEmailService) SendHTMLEmail(ctx context.Context, to, subject, htmlBody string) error {
	log.Printf("Email service stub: Would send HTML email to %s with subject: %s", to, subject)
	return nil
}

func (s *stubEmailService) SendVerificationEmail(ctx context.Context, to, username, verificationURL string) error {
	log.Printf("Email service stub: Would send verification email to %s", to)
	return nil
}

func (s *stubEmailService) SendPasswordResetEmail(ctx context.Context, to, username, resetURL string) error {
	return nil
}

func (s *stubEmailService) SendPasswordChangedEmail(ctx context.Context, to, username string) error {
	return nil
}
