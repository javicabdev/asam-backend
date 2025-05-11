package integration

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	// Import para errors.As de la biblioteca estándar
	stdErrors "errors"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	customErrors "github.com/javicabdev/asam-backend/pkg/errors" // Alias para tu paquete de errores
)

// TestMain asegura que las variables de entorno necesarias estén disponibles
func TestMain(m *testing.M) {
	// Cargar archivo .env si existe (para desarrollo local)
	// En CI, las variables se establecen directamente en el pipeline.
	if os.Getenv("CI") != "true" {
		err := godotenv.Load("../../.env") // Asume que los tests están en test/integration/
		if err != nil {
			log.Println("No .env file found at ../../.env or error loading, continuing with existing environment variables...")
		}
	} else {
		log.Println("CI environment detected. Skipping .env load, relying on pipeline environment variables.")
	}

	_ = os.Setenv("APP_ENV", "test")

	if os.Getenv("CI") == "true" {
		if os.Getenv("JWT_ACCESS_SECRET") == "" {
			log.Println("Warning: JWT_ACCESS_SECRET is not set in CI environment. config.LoadConfig() might fail if it's required.")
		}
		if os.Getenv("JWT_REFRESH_SECRET") == "" {
			log.Println("Warning: JWT_REFRESH_SECRET is not set in CI environment. config.LoadConfig() might fail if it's required.")
		}
	}

	os.Exit(m.Run())
}

// TestInitDB verifica la conexión a la base de datos con diferentes escenarios.
func TestInitDB(t *testing.T) {
	isCI := os.Getenv("CI") == "true"

	if isCI {
		log.Printf("CI Mode in TestInitDB: Expecting DB_HOST=%s, DB_USER=%s, DB_NAME=%s, DB_PORT=%s from pipeline env",
			os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	}

	tests := []struct {
		name          string
		envVarsToSet  map[string]string
		wantErr       bool
		wantErrorCode customErrors.ErrorCode
	}{
		{
			name: "successful connection",
			envVarsToSet: map[string]string{
				"DB_SSL_MODE":          "disable",
				"DB_MAX_IDLE_CONNS":    "5",
				"DB_MAX_OPEN_CONNS":    "15",
				"DB_CONN_MAX_LIFETIME": "2m",
			},
			wantErr:       false,
			wantErrorCode: "",
		},
		{
			name: "invalid credentials",
			envVarsToSet: map[string]string{
				"DB_USER":     "user_que_no_existe_en_ci",
				"DB_PASSWORD": "password_incorrecta_en_ci",
				"DB_SSL_MODE": "disable",
			},
			wantErr:       true,
			wantErrorCode: customErrors.ErrDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalEnvValues := make(map[string]string)
			for k, vToSet := range tt.envVarsToSet {
				originalEnvValues[k] = os.Getenv(k)
				_ = os.Setenv(k, vToSet)
			}
			defer func() {
				for k, originalV := range originalEnvValues {
					if originalV == "" {
						_ = os.Unsetenv(k)
					} else {
						_ = os.Setenv(k, originalV)
					}
				}
			}()

			cfg, err := config.LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() failed: %v. Ensure all required env vars (like JWT secrets) are set.", err)
			}

			t.Logf("Test '%s': Using DB Config: Host=%s, Port=%s, User=%s, DBName=%s, SSLMode=%s",
				tt.name, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBName, cfg.DBSSLMode)

			gdb, errDB := db.InitDB(cfg)

			if tt.wantErr {
				if errDB == nil {
					t.Errorf("InitDB() expected error but got nil")
				} else {
					t.Logf("InitDB() got expected error: %v", errDB)
					if tt.wantErrorCode != "" {
						// Usar errors.As para verificar si errDB es o envuelve un *customErrors.AppError
						var appError *customErrors.AppError // Declarar como puntero
						if stdErrors.As(errDB, &appError) { // Pasar la dirección del puntero
							// Si As tiene éxito, appError es ahora un puntero no nulo a la instancia de AppError.
							// Accedemos al campo Code directamente.
							if appError.Code != tt.wantErrorCode {
								t.Errorf("InitDB() error code = %s, wantErrorCode %s", appError.Code, tt.wantErrorCode)
							}
						} else {
							// Si no es un AppError o no lo envuelve, pero esperábamos un código específico.
							t.Errorf("InitDB() error type is not *customErrors.AppError or does not wrap one (got %T), but expected code %s. Full error: %v", errDB, tt.wantErrorCode, errDB)
						}
					}
				}
			} else {
				if errDB != nil {
					t.Errorf("InitDB() unexpected error = %v", errDB)
				} else {
					if gdb == nil {
						t.Fatal("InitDB() returned nil gdb without error")
					}
					sqlDB, errSQL := gdb.DB()
					if errSQL != nil {
						t.Errorf("gdb.DB() failed: %v", errSQL)
					} else if errPing := sqlDB.Ping(); errPing != nil {
						t.Errorf("sqlDB.Ping() failed: %v", errPing)
					} else {
						t.Logf("Successfully connected and pinged database for test '%s'", tt.name)
					}
				}
			}
		})
	}
}

// TestInitDB_ConnectionRetry prueba la reconexión automática al corregir las credenciales en caliente.
func TestInitDB_ConnectionRetry(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping TestInitDB_ConnectionRetry in CI environment due to potential flakiness with os.Setenv and goroutines timing.")
	}

	t.Run("Should succeed after credentials are corrected (local only)", func(t *testing.T) {
		correctUser := os.Getenv("DB_USER")
		correctPass := os.Getenv("DB_PASSWORD")
		if correctUser == "" || correctPass == "" {
			t.Fatal("DB_USER or DB_PASSWORD not set in local environment for retry test base state.")
		}

		_ = os.Setenv("DB_USER", "no_such_user_for_retry")
		_ = os.Setenv("DB_PASSWORD", "incorrect_password_for_retry")

		defer func() {
			_ = os.Setenv("DB_USER", correctUser)
			_ = os.Setenv("DB_PASSWORD", correctPass)
		}()

		go func() {
			time.Sleep(1 * time.Second)
			log.Println("[DEBUG] TestInitDB_ConnectionRetry: Correcting credentials in goroutine...")
			_ = os.Setenv("DB_USER", correctUser)
			_ = os.Setenv("DB_PASSWORD", correctPass)
		}()

		maxRetries := 5
		retryInterval := 500 * time.Millisecond
		var dbConn *gorm.DB
		var lastErr error

		for i := 0; i < maxRetries; i++ {
			cfg, cfgErr := config.LoadConfig()
			if cfgErr != nil {
				lastErr = fmt.Errorf("LoadConfig() failed on attempt %d: %w", i+1, cfgErr)
				t.Log(lastErr)
				time.Sleep(retryInterval)
				continue
			}

			t.Logf("Attempt %d: Trying to connect with DB_USER=%s", i+1, cfg.DBUser)
			dbConn, lastErr = db.InitDB(cfg)
			if lastErr == nil {
				t.Logf("Attempt %d: Successfully connected to the database.", i+1)
				break
			}
			t.Logf("Attempt %d: db.InitDB error: %v", i+1, lastErr)
			time.Sleep(retryInterval)
		}

		if lastErr != nil {
			t.Fatalf("Failed to connect to the database after %d retries. Last error: %v", maxRetries, lastErr)
		}

		if dbConn == nil {
			t.Fatal("dbConn is nil even though no error was reported from the retry loop.")
		}

		sqlDB, errSQL := dbConn.DB()
		if errSQL != nil {
			t.Fatalf("Failed to get *sql.DB from successfully connected dbConn: %v", errSQL)
		}
		if errPing := sqlDB.Ping(); errPing != nil {
			t.Fatalf("Ping() failed after successful connection through retry: %v", errPing)
		}
		t.Log("Successfully pinged database after retry logic.")
	})
}
