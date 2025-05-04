package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

// Lista de variables de entorno que necesitamos limpiar entre ejecuciones
var envVars = []string{
	"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE",
	"DATABASE_URL", "DB_MAX_IDLE_CONNS", "DB_MAX_OPEN_CONNS", "DB_CONN_MAX_LIFETIME",
}

func init() {
	// Setup command line flags
	flag.StringVar(&environment, "env", "local", "Environment to use (local, aiven, all)")
	flag.StringVar(&command, "cmd", "up", "Migration command (up, down, force, version, etc.)")
}

func main() {
	flag.Parse()

	// Validate environment
	environment = strings.ToLower(environment)
	if environment != "local" && environment != "aiven" && environment != "all" {
		log.Fatalf("Invalid environment. Must be 'local', 'aiven', or 'all'")
	}

	// Get remaining arguments for potential migration numbers
	args := flag.Args()

	// Run migrations for specified environment(s)
	if environment == "local" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running migrations on LOCAL database")
		log.Println("==================================================")

		// Limpiar las variables de entorno antes de cargar el archivo local
		clearEnvVars()

		if err := runMigrations(LocalEnvFile, command, args); err != nil {
			log.Printf("Error migrating local database: %v", err)
			// Si solo estamos ejecutando migraciones para un entorno y falla,
			// salir con error para que el script PowerShell reciba un código de error
			if environment == "local" {
				os.Exit(1)
			}
		}
	}

	if environment == "aiven" || environment == "all" {
		log.Println("==================================================")
		log.Println("Running migrations on AIVEN database")
		log.Println("==================================================")

		// Limpiar las variables de entorno antes de cargar el archivo Aiven
		clearEnvVars()

		if err := runMigrations(AivenEnvFile, command, args); err != nil {
			log.Printf("Error migrating Aiven database: %v", err)
			os.Exit(1)
		}
	}
}

// clearEnvVars limpia las variables de entorno relacionadas con la base de datos
func clearEnvVars() {
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

// runMigrations ejecuta las migraciones con el archivo .env especificado
func runMigrations(envFile string, cmd string, args []string) error {
	// Cargar las variables de entorno del archivo específico
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading %s file: %w", envFile, err)
	}

	// Construir la URL de conexión a la base de datos
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE")

	// Verificar que todas las variables necesarias estén definidas
	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
		return fmt.Errorf("database connection parameters not found in environment file %s", envFile)
	}

	// Imprimir la configuración de la base de datos para depuración
	log.Printf("Database configuration: Host=%s, Port=%s, User=%s, DB=%s, SSL=%s",
		dbHost, dbPort, dbUser, dbName, sslMode)

	// Construir el connection string en formato URL para la biblioteca migrate
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, sslMode)

	// Obtener la ruta del proyecto
	projectDir, err := getProjectDir()
	if err != nil {
		return fmt.Errorf("failed to get project directory: %w", err)
	}

	// Construir la ruta absoluta al directorio de migraciones
	migrationsPath := filepath.Join(projectDir, "migrations")
	migrationsURL := fmt.Sprintf("file://%s", filepath.ToSlash(migrationsPath))

	// Crear una nueva instancia de migrate
	m, err := migrate.New(migrationsURL, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Ejecutar el comando especificado
	switch cmd {
	case "up":
		// Obtener el número de migraciones a aplicar si se especificó
		if len(args) > 0 {
			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number of migrations: %w", err)
			}
			if err := m.Steps(n); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to apply %d migrations: %w", n, err)
			}
		} else {
			// Aplicar todas las migraciones
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to apply all migrations: %w", err)
			}
		}

	case "down":
		// Obtener el número de migraciones a revertir si se especificó
		if len(args) > 0 {
			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid number of migrations: %w", err)
			}
			if err := m.Steps(-n); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to revert %d migrations: %w", n, err)
			}
		} else {
			// Revertir todas las migraciones
			if err := m.Down(); err != nil && err != migrate.ErrNoChange {
				return fmt.Errorf("failed to revert all migrations: %w", err)
			}
		}

	case "version":
		// Mostrar la versión actual
		version, dirty, err := m.Version()
		if err != nil {
			return fmt.Errorf("failed to get migration version: %w", err)
		}
		log.Printf("Current migration version: %d (dirty: %t)", version, dirty)

	case "force":
		// Forzar la versión de la base de datos
		if len(args) == 0 {
			return fmt.Errorf("force command requires a version number")
		}
		v, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		if err := m.Force(int(v)); err != nil {
			return fmt.Errorf("failed to force version %d: %w", v, err)
		}

	case "goto":
		// Migrar a una versión específica
		if len(args) == 0 {
			return fmt.Errorf("goto command requires a version number")
		}
		v, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		if err := m.Migrate(uint(v)); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("failed to migrate to version %d: %w", v, err)
		}

	case "drop":
		// Eliminar todas las tablas
		if err := m.Drop(); err != nil {
			return fmt.Errorf("failed to drop all tables: %w", err)
		}

	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	log.Printf("Migrations completed successfully")
	return nil
}

// getProjectDir obtiene la ruta al directorio raíz del proyecto
func getProjectDir() (string, error) {
	// Obtener la ruta actual
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Comprobar si estamos ya en el directorio raíz del proyecto
	if _, err := os.Stat(filepath.Join(currentDir, "migrations")); err == nil {
		return currentDir, nil
	}

	// Si estamos en cmd/migrate, subir dos niveles
	if strings.HasSuffix(currentDir, filepath.Join("cmd", "migrate")) {
		return filepath.Dir(filepath.Dir(currentDir)), nil
	}

	// Si estamos en cmd, subir un nivel
	if strings.HasSuffix(currentDir, "cmd") {
		return filepath.Dir(currentDir), nil
	}

	// Comprobar si el directorio padre es el raíz del proyecto
	parentDir := filepath.Dir(currentDir)
	if _, err := os.Stat(filepath.Join(parentDir, "migrations")); err == nil {
		return parentDir, nil
	}

	return "", fmt.Errorf("could not find project root directory")
}
