// Package main implementa un servicio para ejecutar migraciones de base de datos
// y proporciona comandos para gestionar la evolución del esquema de la base de datos
// en diferentes entornos (local, desarrollo, producción, etc).
package main

import (
	"errors" // Added standard errors package
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // PostgreSQL driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       // File source driver
	"github.com/joho/godotenv"
)

// Command line options
var (
	environment string
	command     string
)

// Environment files
const (
	LocalEnvFile = ".env.development"
	AivenEnvFile = ".env.aiven"
)

// envVarsToClear is a list of environment variables to clear between runs for different .env files.
var envVarsToClear = []string{
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DATABASE_URL", "DB_MAX_IDLE_CONNS", "DB_MAX_OPEN_CONNS", "DB_CONN_MAX_LIFETIME",
}

func init() {
	// Setup command line flags
	flag.StringVar(&environment, "env", "local", "Environment to use (local, aiven, all)")
	flag.StringVar(&command, "cmd", "up", "Migration command (up, down, force, version, goto, drop)")
}

func main() {
	flag.Parse()

	// Validate environment flag
	environment = strings.ToLower(environment)
	if environment != "local" && environment != "aiven" && environment != "all" {
		log.Fatalf("Invalid environment '%s'. Must be 'local', 'aiven', or 'all'", environment)
	}

	// Get remaining arguments (e.g., number for 'up'/'down', version for 'force'/'goto')
	args := flag.Args()

	// Execute migrations for the specified environment(s)
	if environment == "local" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running migrations on LOCAL database")
		log.Println("==================================================")
		clearEnvironmentVariables()
		if err := executeMigrationsForEnv(LocalEnvFile, command, args); err != nil {
			log.Printf("Error migrating local database: %v", err)
			// If only running for 'local' and it fails, exit with error for scripting purposes.
			if environment == "local" {
				os.Exit(1)
			}
		}
	}

	if environment == "aiven" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running migrations on AIVEN database")
		log.Println("==================================================")
		clearEnvironmentVariables()
		if err := executeMigrationsForEnv(AivenEnvFile, command, args); err != nil {
			log.Printf("Error migrating Aiven database: %v", err)
			os.Exit(1) // Exit with error if Aiven migration fails.
		}
	}

	log.Println("All specified migrations completed.")
}

// clearEnvironmentVariables unsets database-related environment variables.
func clearEnvironmentVariables() {
	log.Println("Clearing database-related environment variables...")
	for _, envVar := range envVarsToClear {
		if err := os.Unsetenv(envVar); err != nil {
			// Log the error but don't necessarily fail the whole process for an unset error.
			log.Printf("Warning: could not unset environment variable %s: %v", envVar, err)
		}
	}
}

// executeMigrationsForEnv loads a specific .env file and runs migrations.
func executeMigrationsForEnv(envFile string, cmd string, args []string) error {
	log.Printf("Loading environment variables from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	// Construct database connection URL
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE") // Default to "disable" if not set, or handle as needed

	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
		return fmt.Errorf("one or more database connection parameters (DB_HOST, DB_PORT, DB_USER, DB_NAME) not found in environment file %s", envFile)
	}
	if sslMode == "" {
		sslMode = "disable" // Default SSL mode if not specified
		log.Printf("DB_SSL_MODE not set, defaulting to '%s'", sslMode)
	}

	log.Printf("Database configuration: Host=%s, Port=%s, User=%s, DB=%s, SSLMode=%s",
		dbHost, dbPort, dbUser, dbName, sslMode)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, sslMode)

	// Determine migrations path
	projectDir, err := getProjectRootDirectory()
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}
	migrationsPath := filepath.Join(projectDir, "migrations")
	migrationsURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))
	log.Printf("Using migrations from: %s", migrationsURL)

	// Create migrate instance
	m, err := migrate.New(migrationsURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w. Check DB connection and migrations path", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("Error closing migration source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("Error closing migration database connection: %v", dbErr)
		}
	}()

	// Execute the specified migration command
	log.Printf("Executing command: '%s' with arguments: %v", cmd, args)
	// Assign the result of runMigrationCommand to err to check it later
	migrationErr := runMigrationCommand(m, cmd, args)
	if migrationErr != nil {
		return migrationErr // Error is already contextualized by runMigrationCommand
	}

	log.Printf("Migrations command '%s' completed successfully for %s", cmd, envFile)
	return nil
}

// runMigrationCommand dispatches to specific command handlers.
func runMigrationCommand(m *migrate.Migrate, cmd string, args []string) error {
	switch cmd {
	case "up":
		return handleMigrationUp(m, args)
	case "down":
		return handleMigrationDown(m, args)
	case "version":
		return handleMigrationVersion(m)
	case "force":
		return handleMigrationForce(m, args)
	case "goto":
		return handleMigrationGoto(m, args)
	case "drop":
		return handleMigrationDrop(m)
	default:
		return fmt.Errorf("unknown command: '%s'. Supported commands are: up, down, version, force, goto, drop", cmd)
	}
}

// handleMigrationUp applies migrations.
func handleMigrationUp(m *migrate.Migrate, args []string) error {
	var err error // Declare err here to check for migrate.ErrNoChange at the end
	if len(args) > 0 {
		n, convErr := strconv.Atoi(args[0])
		if convErr != nil || n < 0 { // Also check for negative numbers
			return fmt.Errorf("invalid number of migrations for 'up' command: '%s'. Must be a non-negative integer: %w", args[0], convErr)
		}
		log.Printf("Applying next %d migration(s)...", n)
		err = m.Steps(n)
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply %d migrations: %w", n, err)
		}
	} else {
		log.Println("Applying all available 'up' migrations...")
		err = m.Up()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply all 'up' migrations: %w", err)
		}
	}
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No new 'up' migrations to apply.")
	}
	return nil
}

// handleMigrationDown reverts migrations.
func handleMigrationDown(m *migrate.Migrate, args []string) error {
	var err error // Declare err to be accessible for the final ErrNoChange check
	if len(args) > 0 {
		n, convErr := strconv.Atoi(args[0])
		if convErr != nil || n < 0 { // Also check for negative numbers
			return fmt.Errorf("invalid number of migrations for 'down' command: '%s'. Must be a non-negative integer: %w", args[0], convErr)
		}
		log.Printf("Reverting last %d migration(s)...", n)
		if n == 0 { // Reverting 0 steps is a no-op.
			log.Println("Reverting 0 migrations means no change.")
			return nil // Or set err to migrate.ErrNoChange if preferred
		}
		err = m.Steps(-n) // Pass negative n to revert
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to revert %d migrations: %w", n, err)
		}
	} else {
		log.Println("Reverting last migration (or all if 'all' specified)...")
		// m.Down() typically reverts one migration.
		// If "revert all" is truly desired, m.Goto(0) is more explicit.
		// This matches the original behavior of reverting one step if no arg.
		err = m.Down()
		if err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to revert last migration: %w", err)
		}
	}
	if errors.Is(err, migrate.ErrNoChange) {
		log.Println("No 'down' migrations to revert or already at the initial state.")
	}
	return nil
}

// handleMigrationVersion shows the current migration version.
func handleMigrationVersion(m *migrate.Migrate) error {
	version, dirty, err := m.Version()
	if err != nil {
		// Special case: if ErrNilVersion, it means no migrations have been applied yet.
		if errors.Is(err, migrate.ErrNilVersion) {
			log.Println("No migrations have been applied yet. Version is considered 0 (not dirty).")
			return nil
		}
		return fmt.Errorf("failed to get migration version: %w", err)
	}
	log.Printf("Current migration version: %d (dirty: %t)", version, dirty)
	return nil
}

// handleMigrationForce sets the database migration version.
func handleMigrationForce(m *migrate.Migrate, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("force command requires a version number argument")
	}
	vStr := args[0]
	// Use ParseInt as m.Force takes an int. MaxInt check was removed in previous version,
	// but it's good practice if there's a specific range. For now, rely on int conversion.
	v, err := strconv.ParseInt(vStr, 10, 0) // bitSize 0 means int
	if err != nil {
		return fmt.Errorf("invalid version number '%s' for force command: %w", vStr, err)
	}
	if v < 0 { // Migration versions are typically non-negative.
		return fmt.Errorf("version number for force command must be non-negative, got: %d", v)
	}
	log.Printf("Forcing migration version to: %d", v)
	if err := m.Force(int(v)); err != nil {
		return fmt.Errorf("failed to force version %d: %w", v, err)
	}
	return nil
}

// handleMigrationGoto migrates to a specific version.
func handleMigrationGoto(m *migrate.Migrate, args []string) error {
	var err error // Declare err here to check for migrate.ErrNoChange at the end
	if len(args) == 0 {
		return fmt.Errorf("goto command requires a version number argument")
	}
	vStr := args[0]
	v, convErr := strconv.ParseUint(vStr, 10, 0) // m.Migrate takes uint, bitSize 0 means uint
	if convErr != nil {
		return fmt.Errorf("invalid version number '%s' for goto command: %w", vStr, convErr)
	}
	log.Printf("Migrating to version: %d...", v)
	err = m.Migrate(uint(v))
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to migrate to version %d: %w", v, err)
	}
	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Already at version %d or no migrations needed to reach it.", v)
	}
	return nil
}

// handleMigrationDrop removes all tables (drops all migrations).
func handleMigrationDrop(m *migrate.Migrate) error {
	log.Println("Dropping all tables (reverting all migrations)...")
	if err := m.Drop(); err != nil {
		return fmt.Errorf("failed to drop all tables: %w", err)
	}
	return nil
}

// getProjectRootDirectory determines the project's root directory by looking for a 'migrations' folder.
// It navigates up from the current working directory.
func getProjectRootDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check up to a certain number of parent directories to avoid infinite loops on strange filesystems.
	for i := 0; i < 10; i++ { // Limit search depth
		migrationsDir := filepath.Join(currentDir, "migrations")
		if _, err := os.Stat(migrationsDir); err == nil {
			// Found migrations directory, this is the project root.
			return currentDir, nil
		}
		// Move to parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached the root of the filesystem without finding 'migrations'
			break
		}
		currentDir = parentDir
	}
	return "", fmt.Errorf("could not find project root directory containing 'migrations' folder (searched upwards from initial CWD: %s)", currentDir) // Added initial CWD for clarity
}
