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
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Parse command line arguments
	if err := parseFlags(); err != nil {
		return err
	}

	// Load environment
	if err := loadEnvironment(); err != nil {
		return err
	}

	// Get and validate credentials
	creds, err := getCredentials()
	if err != nil {
		return err
	}

	// Create or update admin user
	return createOrUpdateAdmin(creds)
}

func parseFlags() error {
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
			return fmt.Errorf("invalid environment '%s'. Must be 'local' or 'aiven'", environment)
		}
	}
	return nil
}

func loadEnvironment() error {
	log.Printf("Loading environment from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading %s file: %v", envFile, err)
	}
	return nil
}

type adminCredentials struct {
	Email    string
	Password string
	Username string
}

func getCredentials() (*adminCredentials, error) {
	creds := &adminCredentials{
		Email:    os.Getenv("ADMIN_EMAIL"),
		Password: os.Getenv("ADMIN_PASSWORD"),
		Username: os.Getenv("ADMIN_USERNAME"),
	}

	// Validate required fields
	if creds.Email == "" {
		return nil, fmt.Errorf("ADMIN_EMAIL environment variable is required. Set it to the admin's email address")
	}
	if creds.Password == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD environment variable is required. Must be at least 8 characters with uppercase, lowercase, and number")
	}

	// If username not provided, use email as username
	if creds.Username == "" {
		creds.Username = creds.Email
		log.Printf("ADMIN_USERNAME not provided, using email as username: %s", creds.Username)
	} else {
		log.Printf("Using custom username: %s (email: %s)", creds.Username, creds.Email)
	}

	// Validate password strength
	if len(creds.Password) < 8 {
		return nil, fmt.Errorf("ADMIN_PASSWORD must be at least 8 characters long")
	}

	return creds, nil
}

func createOrUpdateAdmin(creds *adminCredentials) error {
	// Connect to database
	dbInstance, err := connectDatabase()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	sqlDB, err := dbInstance.DB()
	if err != nil {
		return fmt.Errorf("failed to get SQL DB: %v", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Warning: error closing database connection: %v", err)
		}
	}()

	// Setup services
	services, err := setupServices(dbInstance)
	if err != nil {
		return fmt.Errorf("failed to setup services: %v", err)
	}

	ctx := context.Background()

	// Check if user already exists
	existingUser, _ := services.userRepo.FindByUsername(ctx, creds.Username)

	if existingUser != nil {
		return handleExistingUser(ctx, services, existingUser, creds)
	}

	return createNewAdmin(ctx, services, creds)
}

type appServices struct {
	userRepo    output.UserRepository
	userService input.UserService
}

func setupServices(dbInstance *gorm.DB) (*appServices, error) {
	userRepo := db.NewUserRepository(dbInstance)
	memberRepo := db.NewMemberRepository(dbInstance)
	tokenRepo := db.NewVerificationTokenRepository(dbInstance)

	// Create logger
	appLogger, err := logger.InitLogger(logger.Config{
		Level:       logger.InfoLevel,
		Development: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
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

	return &appServices{
		userRepo:    userRepo,
		userService: userService,
	}, nil
}

func handleExistingUser(ctx context.Context, services *appServices, existingUser *models.User, creds *adminCredentials) error {
	if !forceUpdate {
		log.Printf("Admin user '%s' already exists. Use -force flag to update password.", creds.Username)
		return nil
	}

	log.Printf("Updating existing admin user '%s'...", creds.Username)

	// Update existing user
	updates := map[string]interface{}{
		"password": creds.Password,
		"email":    creds.Email,
		"role":     models.RoleAdmin,
		"isActive": true,
	}

	_, err := services.userService.UpdateUser(ctx, existingUser.ID, updates)
	if err != nil {
		return fmt.Errorf("failed to update admin user: %v", err)
	}

	// Reset email verification status directly in database
	existingUser.EmailVerified = false
	existingUser.EmailVerifiedAt = nil
	if err := services.userRepo.Update(ctx, existingUser); err != nil {
		return fmt.Errorf("failed to reset email verification: %v", err)
	}

	printSuccess(creds.Username, creds.Email, 0, true)
	return nil
}

func createNewAdmin(ctx context.Context, services *appServices, creds *adminCredentials) error {
	log.Printf("Creating new admin user '%s'...", creds.Username)

	// Create new admin user
	user, err := services.userService.CreateUser(
		ctx,
		creds.Username,
		creds.Email,
		creds.Password,
		models.RoleAdmin,
		nil, // No member association for admin
	)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	printSuccess(user.Username, user.Email, user.ID, false)
	printSecurityReminders()
	return nil
}

func printSuccess(username, email string, userID uint, isUpdate bool) {
	log.Println("================================")
	if isUpdate {
		log.Println("Admin user updated successfully!")
	} else {
		log.Println("Admin user created successfully!")
		log.Printf("ID: %d", userID)
	}
	log.Printf("Username: %s", username)
	log.Printf("Email: %s", email)
	log.Printf("Role: %s", models.RoleAdmin)
	log.Println("Email verified: false")
	log.Println("The user must verify their email on first login")
	log.Println("================================")
}

func printSecurityReminders() {
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
