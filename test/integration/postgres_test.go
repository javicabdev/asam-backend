package integration

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// TestMain asegura que las variables de entorno necesarias estén disponibles
func TestMain(m *testing.M) {
	// Cargar archivo .env
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found, continuing...")
	}

	// Configurar variables adicionales para tests
	_ = os.Setenv("APP_ENV", "test")
	defer func() {
		_ = os.Unsetenv("APP_ENV")
	}()

	// Ejecutar tests
	os.Exit(m.Run())
}

// TestInitDB verifica la conexión a la base de datos con diferentes escenarios.
func TestInitDB(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		wantCode errors.ErrorCode // Añadimos el código de error esperado
	}{
		{
			name: "successful connection",
			envVars: map[string]string{
				"DB_HOST":              "postgres-test",
				"DB_PORT":              "5432",
				"DB_USER":              "postgres",
				"DB_PASSWORD":          "123456",
				"DB_NAME":              "asam_test_db",
				"DB_SSL_MODE":          "disable",
				"DB_MAX_IDLE_CONNS":    "5",
				"DB_MAX_OPEN_CONNS":    "15",
				"DB_CONN_MAX_LIFETIME": "2m",
			},
			wantErr:  false,
			wantCode: "", // No se espera error
		},
		{
			name: "invalid credentials",
			envVars: map[string]string{
				"DB_HOST":              "postgres-test",
				"DB_PORT":              "5432",
				"DB_USER":              "no_such_user",
				"DB_PASSWORD":          "whatever",
				"DB_NAME":              "asam_test_db",
				"DB_SSL_MODE":          "disable",
				"DB_MAX_IDLE_CONNS":    "3",
				"DB_MAX_OPEN_CONNS":    "3",
				"DB_CONN_MAX_LIFETIME": "1m",
			},
			wantErr:  true,
			wantCode: errors.ErrDatabaseError, // Esperamos un error de base de datos
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configurar variables de entorno para el test actual
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			// Cargar configuración
			cfg, err := config.LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() unexpected error = %v", err)
			}

			// Probar conexión a la base de datos
			{
				gdb, errDB := db.InitDB(cfg)
				if tt.wantErr {
					if errDB == nil {
						t.Errorf("InitDB() expected error but got nil")
					} else if tt.wantCode != "" && !errors.Is(errDB, tt.wantCode) {
						t.Errorf("InitDB() error = %v, wantCode %v", errDB, tt.wantCode)
					}
				} else if errDB != nil {
					t.Errorf("InitDB() unexpected error = %v", errDB)
				} else {
					// Verificar conexión
					sqlDB, err := gdb.DB()
					if err != nil {
						t.Errorf("Failed to get *sql.DB: %v", err)
					} else if errPing := sqlDB.Ping(); errPing != nil {
						t.Errorf("Failed to ping database: %v", errPing)
					}
				}
			}

			// Limpiar variables de entorno
			for k := range tt.envVars {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// TestInitDB_ConnectionRetry prueba la reconexión automática al corregir las credenciales en caliente.
func TestInitDB_ConnectionRetry(t *testing.T) {
	t.Run("Should succeed after credentials are corrected", func(t *testing.T) {
		// Configurar credenciales iniciales incorrectas
		_ = os.Setenv("DB_HOST", "postgres-test")
		_ = os.Setenv("DB_PORT", "5432")
		_ = os.Setenv("DB_USER", "no_such_user")
		_ = os.Setenv("DB_PASSWORD", "wrongpass")
		_ = os.Setenv("DB_NAME", "asam_test_db")
		_ = os.Setenv("DB_SSL_MODE", "disable")

		// Corregir credenciales después de 2 segundos
		go func() {
			time.Sleep(2 * time.Second)
			_ = os.Setenv("DB_USER", "postgres")
			_ = os.Setenv("DB_PASSWORD", "123456")
			fmt.Println("[DEBUG] Credenciales corregidas en el test")
		}()

		// Intentar reconectar
		maxRetries := 5
		retryInterval := 1 * time.Second
		var dbConn *gorm.DB
		var err error

		for i := 0; i < maxRetries; i++ {
			cfg, cfgErr := config.LoadConfig()
			if cfgErr != nil {
				err = cfgErr
				t.Logf("Attempt %d: LoadConfig() error = %v", i+1, cfgErr)
			} else {
				dbConn, err = db.InitDB(cfg)
				if err == nil {
					t.Logf("Attempt %d: Successfully connected to the database", i+1)
					break
				}
				if errors.IsDatabaseError(err) {
					t.Logf("Attempt %d: Database error = %v", i+1, err)
				} else {
					t.Logf("Attempt %d: Unexpected error = %v", i+1, err)
				}
			}
			time.Sleep(retryInterval)
		}

		if err != nil {
			t.Fatalf("Failed to connect to the database after %d retries: %v", maxRetries, err)
		}

		if dbConn != nil {
			sqlDB, err := dbConn.DB()
			if err != nil {
				t.Fatalf("Failed to get *sql.DB: %v", err)
			}
			if errPing := sqlDB.Ping(); errPing != nil {
				t.Fatalf("Ping() error: %v", errPing)
			}
		}

		// Limpiar variables de entorno
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("DB_PORT")
		_ = os.Unsetenv("DB_USER")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_NAME")
		_ = os.Unsetenv("DB_SSL_MODE")
	})
}
