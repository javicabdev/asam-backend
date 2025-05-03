package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/javicabdev/asam-backend/test/seed"
	"github.com/javicabdev/asam-backend/test/seed/data"
)

// Command line options
var (
	datasetType   string
	scenario      string
	clean         bool
	randomSeed    int64
	enableConsole bool
	verbose       bool
	concurrency   int
	envFile       string
	numMembers    int
	numFamilies   int
	numFamiliares int
	numPayments   int
	numCashflows  int
)

func init() {
	// Setup command line flags
	flag.StringVar(&datasetType, "type", "minimal", "Dataset type (minimal, full, scenario)")
	flag.StringVar(&scenario, "scenario", "payment_overdue", "Scenario name when type=scenario")
	flag.BoolVar(&clean, "clean", false, "Only clean the database without seeding")
	flag.Int64Var(&randomSeed, "seed", time.Now().UnixNano(), "Random seed for reproducible generation")
	flag.BoolVar(&enableConsole, "console", true, "Enable console output")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.IntVar(&concurrency, "concurrency", 5, "Number of concurrent operations")
	flag.StringVar(&envFile, "env", ".env", "Path to environment file")
	flag.IntVar(&numMembers, "members", 0, "Number of members to generate (override default)")
	flag.IntVar(&numFamilies, "families", 0, "Number of families to generate (override default)")
	flag.IntVar(&numFamiliares, "familiares", 0, "Number of familiares to generate (override default)")
	flag.IntVar(&numPayments, "payments", 0, "Number of payments to generate (override default)")
	flag.IntVar(&numCashflows, "cashflows", 0, "Number of cashflows to generate (override default)")
}

func main() {
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database connection string from environment
	dbConn := os.Getenv("DATABASE_URL")
	if dbConn == "" {
		// Build connection string from individual parameters
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
			log.Fatalf("Database connection parameters not found in environment")
		}

		dbConn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, dbName,
		)
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure seeder
	seeder := seed.NewSeeder(db).
		WithRandomSeed(randomSeed).
		WithLogging(enableConsole).
		WithConcurrency(concurrency)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute command
	if clean {
		// Only clean the database
		if err := seeder.Clean(ctx); err != nil {
			log.Fatalf("Failed to clean database: %v", err)
		}
		log.Println("Database cleaned successfully")
		return
	}

	// Seed the database according to the specified type
	var dataset data.Seedable

	switch datasetType {
	case "minimal":
		dataset = data.Dataset(db, seeder, data.MinimalType)
	case "full":
		dataset = data.Dataset(db, seeder, data.FullType)
	case "scenario":
		dataset = data.Dataset(db, seeder, data.ScenarioType, scenario)
	case "custom":
		// Custom seeding with specific counts
		if err := seedCustom(ctx, seeder); err != nil {
			log.Fatalf("Failed to seed custom dataset: %v", err)
		}
		return
	default:
		log.Fatalf("Unknown dataset type: %s", datasetType)
	}

	// Seed the dataset
	if err := dataset.Seed(ctx); err != nil {
		log.Fatalf("Failed to seed dataset: %v", err)
	}

	log.Printf("Database seeded successfully with %s dataset", datasetType)
}

// seedCustom seeds the database with custom entity counts
func seedCustom(ctx context.Context, seeder *seed.Seeder) error {
	log.Println("Seeding custom dataset")

	// Clean the database first
	if err := seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Seed members
	if numMembers > 0 {
		log.Printf("Seeding %d members", numMembers)
		if err := seeder.SeedMiembros(ctx); err != nil {
			return fmt.Errorf("failed to seed members: %w", err)
		}
	}

	// Seed families
	if numFamilies > 0 {
		log.Printf("Seeding %d families", numFamilies)
		if err := seeder.SeedFamilias(ctx); err != nil {
			return fmt.Errorf("failed to seed families: %w", err)
		}
	}

	// Seed familiares
	if numFamiliares > 0 {
		log.Printf("Seeding %d familiares", numFamiliares)
		if err := seeder.SeedFamiliares(ctx); err != nil {
			return fmt.Errorf("failed to seed familiares: %w", err)
		}
	}

	// Seed payments
	if numPayments > 0 {
		log.Printf("Seeding %d payments", numPayments)
		if err := seeder.SeedCuotasMembresia(ctx); err != nil {
			return fmt.Errorf("failed to seed payments: %w", err)
		}
	}

	// Seed cashflows
	if numCashflows > 0 {
		log.Printf("Seeding %d cashflows", numCashflows)
		if err := seeder.SeedCaja(ctx); err != nil {
			return fmt.Errorf("failed to seed cashflows: %w", err)
		}
	}

	return nil
}
