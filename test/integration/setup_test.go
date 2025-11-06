package integration

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/output"
)

// setupTestDB crea una conexión a la base de datos de prueba
// y retorna una función de limpieza para ejecutar después del test
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Leer configuración desde variables de entorno
	// Por defecto usa las credenciales del CI (asam_test)
	// En desarrollo local, pasar DB_NAME=asam_db si es necesario
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "asam_test")
	sslMode := getEnvOrDefault("DB_SSL_MODE", "disable")

	// Construir DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslMode,
	)

	// Conectar a la base de datos
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silenciar logs en tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// No ejecutar AutoMigrate si las tablas ya existen
	// La base de datos de desarrollo ya tiene las migraciones aplicadas
	// Solo verificar que podemos acceder a la BD
	var count int64
	if err := database.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'cash_flows'").Scan(&count).Error; err != nil {
		t.Fatalf("Failed to check database: %v", err)
	}
	if count == 0 {
		t.Fatal("cash_flows table does not exist. Please run migrations first.")
	}

	// Función de limpieza de tablas
	cleanTables := func() {
		// Orden importante: primero las tablas con foreign keys
		database.Exec("TRUNCATE TABLE cash_flows RESTART IDENTITY CASCADE")
		database.Exec("TRUNCATE TABLE payments RESTART IDENTITY CASCADE")
		database.Exec("TRUNCATE TABLE familiares RESTART IDENTITY CASCADE")
		database.Exec("TRUNCATE TABLE families RESTART IDENTITY CASCADE")
		database.Exec("TRUNCATE TABLE members RESTART IDENTITY CASCADE")
		database.Exec("TRUNCATE TABLE membership_fees RESTART IDENTITY CASCADE")
	}

	// Limpiar todas las tablas ANTES del test para garantizar estado limpio
	cleanTables()

	// Función de limpieza
	cleanup := func() {
		// Limpiar todas las tablas después del test también
		cleanTables()

		sqlDB, _ := database.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	return database, cleanup
}

// setupCashFlowRepository crea una instancia del repositorio de CashFlow
func setupCashFlowRepository(database *gorm.DB) output.CashFlowRepository {
	return db.NewCashFlowRepository(database)
}

// setupCashFlowService crea una instancia del servicio de CashFlow
func setupCashFlowService(repo output.CashFlowRepository) *services.CashFlowService {
	return services.NewCashFlowService(repo)
}

// getEnvOrDefault obtiene una variable de entorno o retorna un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
