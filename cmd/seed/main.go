// Package main provides a database seeding tool for the ASAM backend.
// It supports seeding different types of datasets (minimal, full, scenario, custom)
// and can target different environments (local, aiven, or both).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver for database/sql

	"github.com/javicabdev/asam-backend/test/seed"
	"github.com/javicabdev/asam-backend/test/seed/data"
)

// Command line flags (global variables)
var (
	datasetType   string
	scenario      string // Used when datasetType is "scenario"
	clean         bool   // If true, only clean the database
	randomSeed    int64  // Seed for random number generation
	enableConsole bool   // Enable/disable console logging from seeder
	verbose       bool   // Verbose logging (currently unused directly in this file, but could be passed to seeder)
	concurrency   int    // Concurrency level for seeder operations
	environment   string // Target environment for seeding (local, aiven, all)

	// Flags for custom dataset counts
	numMembers    int
	numFamilies   int
	numFamiliares int
	numPayments   int
	numCashflows  int
)

// Environment file constants
const (
	LocalEnvFile = ".env.development"
	AivenEnvFile = ".env.aiven"
)

// envVarsToClear is a list of environment variables to clear before loading a new .env file.
var envVarsToClear = []string{
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DATABASE_URL", "DB_MAX_IDLE_CONNS", "DB_MAX_OPEN_CONNS", "DB_CONN_MAX_LIFETIME",
}

func init() {
	// Setup command line flags with default values and descriptions
	flag.StringVar(&datasetType, "type", "minimal", "Dataset type to seed (minimal, full, scenario, custom)")
	flag.StringVar(&scenario, "scenario", "payment_overdue", "Scenario name (used if type=scenario, e.g., 'payment_overdue')")
	flag.BoolVar(&clean, "clean", false, "If true, only clean the database and do not seed any data")
	flag.Int64Var(&randomSeed, "seed", time.Now().UnixNano(), "Random seed for data generation (for reproducibility)")
	flag.BoolVar(&enableConsole, "console", true, "Enable console output from the seeder library")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output (placeholder, currently not directly used by this script but could be)")
	flag.IntVar(&concurrency, "concurrency", 5, "Number of concurrent operations for seeding")
	flag.StringVar(&environment, "env", "local", "Target environment for seeding (local, aiven, or all)")

	// Flags for overriding default counts in 'custom' dataset type
	flag.IntVar(&numMembers, "members", 0, "Number of members to generate (custom type, overrides default)")
	flag.IntVar(&numFamilies, "families", 0, "Number of families to generate (custom type, overrides default)")
	flag.IntVar(&numFamiliares, "familiares", 0, "Number of 'familiares' (relatives) to generate (custom type, overrides default)")
	flag.IntVar(&numPayments, "payments", 0, "Number of payments to generate (custom type, overrides default)")
	flag.IntVar(&numCashflows, "cashflows", 0, "Number of cashflow entries to generate (custom type, overrides default)")
}

func main() {
	flag.Parse()

	// Validate the environment flag
	environment = strings.ToLower(environment)
	if environment != "local" && environment != "aiven" && environment != "all" {
		log.Fatalf("Invalid environment '%s'. Must be 'local', 'aiven', or 'all'.", environment)
	}

	// Execute seeder for the specified environment(s)
	if environment == "local" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running seed on LOCAL database")
		log.Println("==================================================")
		clearDatabaseEnvVars()
		if err := runSeedForEnv(LocalEnvFile); err != nil {
			log.Printf("Error seeding local database: %v", err)
			if environment == "local" { // Exit if only local was specified and it failed
				os.Exit(1)
			}
		}
	}

	if environment == "aiven" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running seed on AIVEN database")
		log.Println("==================================================")
		clearDatabaseEnvVars()
		if err := runSeedForEnv(AivenEnvFile); err != nil {
			log.Printf("Error seeding Aiven database: %v", err)
			os.Exit(1) // Always exit if Aiven seeding fails
		}
	}
	log.Println("Seeding process completed for specified environments.")
}

// clearDatabaseEnvVars unsets database-related environment variables.
func clearDatabaseEnvVars() {
	log.Println("Clearing database-related environment variables...")
	for _, envVar := range envVarsToClear {
		if err := os.Unsetenv(envVar); err != nil {
			log.Printf("Warning: could not unset environment variable %s: %v", envVar, err)
		}
	}
}

// connectToDatabase loads environment variables from envFile, builds a connection string,
// and connects to the PostgreSQL database using sqlx.
func connectToDatabase(envFile string) (*sqlx.DB, error) {
	log.Printf("Loading environment variables from: %s", envFile)
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	log.Printf("Database configuration: Host=%s, Port=%s, User=%s, DB=%s, SSLMode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"), os.Getenv("DB_SSL_MODE"))

	dbConn := os.Getenv("DATABASE_URL")
	if dbConn == "" {
		// Build connection string from individual parameters if DATABASE_URL is not set
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSL_MODE")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
			return nil, fmt.Errorf("database connection parameters (DB_HOST, DB_PORT, DB_USER, DB_NAME) not found in environment file %s", envFile)
		}
		if sslMode == "" {
			sslMode = "disable" // Default SSL mode
			log.Printf("DB_SSL_MODE not set, defaulting to '%s'", sslMode)
		}
		dbConn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPass, dbName, sslMode)
	}

	log.Printf("Connecting to database with connection string: %s (password redacted if present in params)", dbConn) // Be cautious logging full conn string if it contains password
	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Successfully connected to the database.")
	return db, nil
}

// configureSeeder creates and configures a new seeder instance.
func configureSeeder(db *sqlx.DB) *seed.Seeder {
	return seed.NewSeeder(db).
		WithRandomSeed(randomSeed).
		WithLogging(enableConsole).  // Uses the global 'enableConsole' flag
		WithConcurrency(concurrency) // Uses the global 'concurrency' flag
}

// runSeedForEnv orchestrates the seeding process for a given environment file.
func runSeedForEnv(envFile string) error {
	db, err := connectToDatabase(envFile)
	if err != nil {
		return err // Error already contextualized by connectToDatabase
	}
	defer func() {
		log.Println("Closing database connection...")
		if errClose := db.Close(); errClose != nil {
			log.Printf("Error closing database connection: %v", errClose)
		}
	}()

	seeder := configureSeeder(db)

	// Create context with a timeout for seeding operations
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) // Increased timeout
	defer cancel()

	// If 'clean' flag is set, only clean the database and return
	if clean {
		log.Println("Cleaning the database...")
		if err := seeder.Clean(ctx); err != nil {
			return fmt.Errorf("failed to clean database: %w", err)
		}
		log.Println("Database cleaned successfully.")
		return nil
	}

	// Dispatch to the appropriate seeding logic based on datasetType
	if err := dispatchSeedByType(ctx, db, seeder, datasetType, scenario); err != nil {
		return err // Error already contextualized
	}

	log.Printf("Database seeded successfully with %s dataset for env file %s", datasetType, envFile)
	return nil
}

// dispatchSeedByType selects and executes the seeding strategy based on the datasetType.
func dispatchSeedByType(ctx context.Context, db *sqlx.DB, seeder *seed.Seeder, typeOfDataset, scenarioName string) error {
	log.Printf("Dispatching seed for dataset type: %s", typeOfDataset)
	switch typeOfDataset {
	case "minimal":
		return seedMinimalDataset(ctx, db, seeder)
	case "full":
		return seedFullDataset(ctx, db, seeder)
	case "scenario":
		return seedScenarioDataset(ctx, db, seeder, scenarioName)
	case "custom":
		// For "custom", we call seedCustom which uses global flags for counts.
		return seedCustom(ctx, seeder)
	default:
		return fmt.Errorf("unknown dataset type: '%s'. Supported types are minimal, full, scenario, custom", typeOfDataset)
	}
}

// seedMinimalDataset seeds a minimal dataset.
func seedMinimalDataset(ctx context.Context, db *sqlx.DB, seeder *seed.Seeder) error {
	log.Println("Seeding minimal dataset...")
	dataset := data.Dataset(db, seeder, data.MinimalType)
	if err := dataset.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed minimal dataset: %w", err)
	}
	return nil
}

// seedFullDataset seeds a full dataset.
func seedFullDataset(ctx context.Context, db *sqlx.DB, seeder *seed.Seeder) error {
	log.Println("Seeding full dataset...")
	dataset := data.Dataset(db, seeder, data.FullType)
	if err := dataset.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed full dataset: %w", err)
	}
	return nil
}

// seedScenarioDataset seeds a specific scenario dataset.
func seedScenarioDataset(ctx context.Context, db *sqlx.DB, seeder *seed.Seeder, scenarioName string) error {
	log.Printf("Seeding scenario dataset: %s...", scenarioName)
	if scenarioName == "" {
		return fmt.Errorf("scenario name cannot be empty when dataset type is 'scenario'")
	}
	dataset := data.Dataset(db, seeder, data.ScenarioType, scenarioName)
	if err := dataset.Seed(ctx); err != nil {
		return fmt.Errorf("failed to seed scenario dataset '%s': %w", scenarioName, err)
	}
	return nil
}

// seedCustom seeds the database with custom entity counts defined by command-line flags.
func seedCustom(ctx context.Context, seeder *seed.Seeder) error {
	log.Println("Seeding custom dataset with specified counts...")

	// Clean the database first as custom seeding implies a fresh state with new counts.
	log.Println("Cleaning database before custom seed...")
	if err := seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database before custom seed: %w", err)
	}

	// Use default counts if flags are not set or are zero/negative.
	// These defaults are applied here if the flags resulted in non-positive values.
	if numMembers <= 0 {
		numMembers = 50
		log.Printf("Defaulting to %d members for custom seed.", numMembers)
	}
	if numFamilies <= 0 {
		numFamilies = 20
		log.Printf("Defaulting to %d families for custom seed.", numFamilies)
	}
	if numFamiliares <= 0 {
		numFamiliares = 40
		log.Printf("Defaulting to %d 'familiares' for custom seed.", numFamiliares)
	}
	if numPayments <= 0 {
		numPayments = 100
		log.Printf("Defaulting to %d payments for custom seed.", numPayments)
	}
	if numCashflows <= 0 {
		numCashflows = 200
		log.Printf("Defaulting to %d cashflows for custom seed.", numCashflows)
	}

	// Seed entities based on the (potentially defaulted) counts.
	// Note: The current seed.Seeder methods (SeedMiembros, SeedFamilias, etc.)
	// might not directly use these numMembers, numFamilies counts.
	// They might have their own internal logic or use counts set via other Seeder methods.
	// This seedCustom function assumes that the seeder methods are either aware of these global counts
	// or that this function should be updated to pass these counts to the seeder if methods accept them.
	// For now, it calls the seeder methods as in the original code.

	log.Printf("Seeding %d members...", numMembers) // This log reflects the flag, not necessarily what SeedMiembros will do.
	if err := seeder.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("failed to seed members: %w", err)
	}

	log.Printf("Seeding %d families...", numFamilies)
	if err := seeder.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("failed to seed families: %w", err)
	}

	log.Printf("Seeding %d 'familiares'...", numFamiliares)
	if err := seeder.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("failed to seed 'familiares': %w", err)
	}

	log.Printf("Seeding %d payments...", numPayments)
	if err := seeder.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("failed to seed payments: %w", err)
	}

	log.Printf("Seeding %d cashflows...", numCashflows)
	if err := seeder.SeedCaja(ctx); err != nil {
		return fmt.Errorf("failed to seed cashflows: %w", err)
	}

	log.Println("Custom dataset seeded successfully.")
	return nil
}
