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
	environment   string
	numMembers    int
	numFamilies   int
	numFamiliares int
	numPayments   int
	numCashflows  int
)

// Environment files
const (
	LocalEnvFile = ".env.development"
	AivenEnvFile = ".env.aiven"
)

// Lista de variables de entorno que necesitamos limpiar entre ejecuciones
var envVars = []string{
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DATABASE_URL", "DB_MAX_IDLE_CONNS", "DB_MAX_OPEN_CONNS", "DB_CONN_MAX_LIFETIME",
}

func init() {
	// Setup command line flags
	flag.StringVar(&datasetType, "type", "minimal", "Dataset type (minimal, full, scenario, custom)")
	flag.StringVar(&scenario, "scenario", "payment_overdue", "Scenario name when type=scenario")
	flag.BoolVar(&clean, "clean", false, "Only clean the database without seeding")
	flag.Int64Var(&randomSeed, "seed", time.Now().UnixNano(), "Random seed for reproducible generation")
	flag.BoolVar(&enableConsole, "console", true, "Enable console output")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.IntVar(&concurrency, "concurrency", 5, "Number of concurrent operations")
	flag.StringVar(&environment, "env", "local", "Environment to use (local, aiven, all)")
	flag.IntVar(&numMembers, "members", 0, "Number of members to generate (override default)")
	flag.IntVar(&numFamilies, "families", 0, "Number of families to generate (override default)")
	flag.IntVar(&numFamiliares, "familiares", 0, "Number of familiares to generate (override default)")
	flag.IntVar(&numPayments, "payments", 0, "Number of payments to generate (override default)")
	flag.IntVar(&numCashflows, "cashflows", 0, "Number of cashflows to generate (override default)")
}

func main() {
	flag.Parse()

	// Validate environment
	environment = strings.ToLower(environment)
	if environment != "local" && environment != "aiven" && environment != "all" {
		log.Fatalf("Invalid environment. Must be 'local', 'aiven', or 'all'")
	}

	// Run seeder for specified environment(s)
	if environment == "local" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running seed on LOCAL database")
		log.Println("==================================================")

		// Limpiar las variables de entorno antes de cargar el archivo local
		clearEnvVars()

		if err := runSeed(LocalEnvFile); err != nil {
			log.Printf("Error seeding local database: %v", err)
		}
	}

	if environment == "aiven" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running seed on AIVEN database")
		log.Println("==================================================")

		// Limpiar las variables de entorno antes de cargar el archivo Aiven
		clearEnvVars()

		if err := runSeed(AivenEnvFile); err != nil {
			log.Printf("Error seeding Aiven database: %v", err)
		}
	}
}

// clearEnvVars limpia las variables de entorno relacionadas con la base de datos
func clearEnvVars() {
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

// runSeed executes the seeder with the specified env file
func runSeed(envFile string) error {
	// Cargar las variables de entorno del archivo específico
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	// Imprimir la configuración de la base de datos para depuración
	log.Printf("Database configuration: Host=%s, Port=%s, User=%s, DB=%s, SSL=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSL_MODE"))

	// Get database connection string from environment
	dbConn := os.Getenv("DATABASE_URL")
	if dbConn == "" {
		// Build connection string from individual parameters
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")
		sslMode := os.Getenv("DB_SSL_MODE")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
			return fmt.Errorf("database connection parameters not found in environment file %s", envFile)
		}

		dbConn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			dbHost, dbPort, dbUser, dbPass, dbName, sslMode,
		)
	}

	// Connect to database
	db, err := sqlx.Connect("postgres", dbConn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
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
			return fmt.Errorf("failed to clean database: %w", err)
		}
		log.Println("Database cleaned successfully")
		return nil
	}

	// Seed the database according to the specified type
	switch datasetType {
	case "minimal":
		dataset := data.Dataset(db, seeder, data.MinimalType)
		if err := dataset.Seed(ctx); err != nil {
			return fmt.Errorf("failed to seed minimal dataset: %w", err)
		}
	case "full":
		dataset := data.Dataset(db, seeder, data.FullType)
		if err := dataset.Seed(ctx); err != nil {
			return fmt.Errorf("failed to seed full dataset: %w", err)
		}
	case "scenario":
		dataset := data.Dataset(db, seeder, data.ScenarioType, scenario)
		if err := dataset.Seed(ctx); err != nil {
			return fmt.Errorf("failed to seed scenario dataset: %w", err)
		}
	case "custom":
		// Custom seeding with specific counts
		if err := seedCustom(ctx, seeder); err != nil {
			return fmt.Errorf("failed to seed custom dataset: %w", err)
		}
	default:
		return fmt.Errorf("unknown dataset type: %s", datasetType)
	}

	log.Printf("Database seeded successfully with %s dataset", datasetType)
	return nil
}

// seedCustom seeds the database with custom entity counts
func seedCustom(ctx context.Context, seeder *seed.Seeder) error {
	log.Println("Seeding custom dataset")

	// Clean the database first
	if err := seeder.Clean(ctx); err != nil {
		return fmt.Errorf("failed to clean database: %w", err)
	}

	// Set default values if not specified
	if numMembers <= 0 {
		numMembers = 50
	}
	if numFamilies <= 0 {
		numFamilies = 20
	}
	if numFamiliares <= 0 {
		numFamiliares = 40
	}
	if numPayments <= 0 {
		numPayments = 100
	}
	if numCashflows <= 0 {
		numCashflows = 200
	}

	// Seed members
	log.Printf("Seeding %d members", numMembers)
	if err := seeder.SeedMiembros(ctx); err != nil {
		return fmt.Errorf("failed to seed members: %w", err)
	}

	// Seed families
	log.Printf("Seeding %d families", numFamilies)
	if err := seeder.SeedFamilias(ctx); err != nil {
		return fmt.Errorf("failed to seed families: %w", err)
	}

	// Seed familiares
	log.Printf("Seeding %d familiares", numFamiliares)
	if err := seeder.SeedFamiliares(ctx); err != nil {
		return fmt.Errorf("failed to seed familiares: %w", err)
	}

	// Seed payments
	log.Printf("Seeding %d payments", numPayments)
	if err := seeder.SeedCuotasMembresia(ctx); err != nil {
		return fmt.Errorf("failed to seed payments: %w", err)
	}

	// Seed cashflows
	log.Printf("Seeding %d cashflows", numCashflows)
	if err := seeder.SeedCaja(ctx); err != nil {
		return fmt.Errorf("failed to seed cashflows: %w", err)
	}

	return nil
}
