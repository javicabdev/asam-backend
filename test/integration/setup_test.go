package integration

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/output"
)

// testMutex ensures integration tests run sequentially to avoid database conflicts
var testMutex sync.Mutex

// setupTestDB crea una conexión a la base de datos de prueba
// y retorna una función de limpieza para ejecutar después del test
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Adquirir mutex para asegurar que solo un test corra a la vez
	testMutex.Lock()

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
		// Usar TRUNCATE que hace hard delete y es más rápido
		// DELETE respeta soft deletes de GORM (deleted_at), dejando datos fantasma
		// TRUNCATE hace hard delete físico que ignora soft deletes completamente
		tables := []string{
			"cash_flows",
			"payments",
			"familiares",
			"families",
			"members",
			"membership_fees",
		}

		// Ejecutar TRUNCATE para cada tabla (hard delete, ignora soft deletes)
		for _, table := range tables {
			// TRUNCATE hace hard delete y resetea secuencias automáticamente con RESTART IDENTITY
			result := database.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
			if result.Error != nil {
				// Solo loguear si la tabla existe (ignorar error de tabla no existente)
				if !isTableNotExistError(result.Error) {
					t.Logf("Warning: Failed to truncate table %s: %v", table, result.Error)
				}
			}
		}

		// Verificar que cash_flows está realmente vacío (incluyendo soft-deleted)
		var count int64
		database.Raw("SELECT COUNT(*) FROM cash_flows").Scan(&count)
		if count > 0 {
			t.Logf("WARNING: cash_flows table still has %d rows after cleanup!", count)
		}
	}

	// Limpiar todas las tablas ANTES del test para garantizar estado limpio
	t.Logf("Setting up test database, acquiring lock...")
	cleanTables()
	t.Logf("Tables cleaned, ready for test")

	// Función de limpieza
	cleanup := func() {
		// Liberar mutex para permitir que el siguiente test corra
		defer testMutex.Unlock()

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

// setupPaymentRepository crea una instancia del repositorio de Payment
func setupPaymentRepository(database *gorm.DB) output.PaymentRepository {
	return db.NewPaymentRepository(database)
}

// setupMemberRepository crea una instancia del repositorio de Member
func setupMemberRepository(database *gorm.DB) output.MemberRepository {
	return db.NewMemberRepository(database)
}

// setupMembershipFeeRepository crea una instancia del repositorio de MembershipFee
func setupMembershipFeeRepository(database *gorm.DB) output.MembershipFeeRepository {
	return db.NewMembershipFeeRepository(database)
}

// getEnvOrDefault obtiene una variable de entorno o retorna un valor por defecto
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isTableNotExistError verifica si un error es porque la tabla no existe
func isTableNotExistError(err error) bool {
	if err == nil {
		return false
	}
	// PostgreSQL error code 42P01 = undefined_table
	return err.Error() == "ERROR: relation \"familiares\" does not exist (SQLSTATE 42P01)" ||
		err.Error() == "ERROR: relation \"families\" does not exist (SQLSTATE 42P01)"
}
