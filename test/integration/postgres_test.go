// test/integration/postgres_test.go
package integration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"gorm.io/gorm"
)

// TestInitDB verifica la conexión a la base de datos con diferentes escenarios.
func TestInitDB(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "successful connection",
			envVars: map[string]string{
				"DB_HOST":              "localhost",
				"DB_PORT":              "5432",
				"DB_USER":              "postgres",
				"DB_PASSWORD":          "123456",
				"DB_NAME":              "asam_db",
				"DB_SSL_MODE":          "disable",
				"DB_MAX_IDLE_CONNS":    "5",
				"DB_MAX_OPEN_CONNS":    "15",
				"DB_CONN_MAX_LIFETIME": "2m",
			},
			wantErr: false,
		},
		{
			name: "invalid credentials",
			envVars: map[string]string{
				"DB_HOST":              "localhost",
				"DB_PORT":              "5432",
				"DB_USER":              "no_such_user", // Rol inexistente para forzar error
				"DB_PASSWORD":          "whatever",
				"DB_NAME":              "asam_db",
				"DB_SSL_MODE":          "disable",
				"DB_MAX_IDLE_CONNS":    "3",
				"DB_MAX_OPEN_CONNS":    "3",
				"DB_CONN_MAX_LIFETIME": "1m",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt // Captura el valor actual de tt
		t.Run(tt.name, func(t *testing.T) {
			// No usar t.Parallel() para evitar conflictos en variables de entorno

			// 1. Set environment variables para el test actual
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			// 2. Cargar la configuración
			cfg, err := config.LoadConfig()
			if err != nil && !tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 3. Inicializar la DB si la configuración se cargó correctamente
			if err == nil {
				gdb, errDB := db.InitDB(cfg)
				if (errDB != nil) != tt.wantErr {
					t.Errorf("InitDB() error = %v, wantErr %v", errDB, tt.wantErr)
				}

				// 4. Para los casos sin error, verificar que podemos hacer ping
				if !tt.wantErr && gdb != nil {
					sqlDB, err := gdb.DB()
					if err != nil {
						t.Errorf("Failed to get *sql.DB: %v", err)
					} else {
						if errPing := sqlDB.Ping(); errPing != nil {
							t.Errorf("Failed to ping database: %v", errPing)
						}
					}
				}
			}

			// 5. Clean up: unsetear variables de entorno
			for k := range tt.envVars {
				_ = os.Unsetenv(k)
			}
		})
	}
}

// TestInitDB_ConnectionRetry prueba la reconexión automática al corregir las credenciales en caliente.
func TestInitDB_ConnectionRetry(t *testing.T) {
	t.Run("Should succeed after credentials are corrected", func(t *testing.T) {
		// 1. Setear inicialmente credenciales inválidas
		_ = os.Setenv("DB_HOST", "localhost")
		_ = os.Setenv("DB_PORT", "5432")
		_ = os.Setenv("DB_USER", "no_such_user") // Rol inexistente
		_ = os.Setenv("DB_PASSWORD", "wrongpass")
		_ = os.Setenv("DB_NAME", "asam_db")
		_ = os.Setenv("DB_SSL_MODE", "disable")
		_ = os.Setenv("DB_MAX_IDLE_CONNS", "3")
		_ = os.Setenv("DB_MAX_OPEN_CONNS", "3")
		_ = os.Setenv("DB_CONN_MAX_LIFETIME", "1m")

		// 2. Goroutine para corregir las credenciales después de 2 segundos
		go func() {
			time.Sleep(2 * time.Second)
			_ = os.Setenv("DB_USER", "postgres")
			_ = os.Setenv("DB_PASSWORD", "123456")
			fmt.Println("[DEBUG] Credenciales corregidas en el test")
		}()

		// 3. Implementar reconexión manual en el test
		maxRetries := 5
		retryInterval := 1 * time.Second
		var dbConn *gorm.DB
		var err error

		for i := 0; i < maxRetries; i++ {
			// Cargar la configuración actualizada
			cfg, cfgErr := config.LoadConfig()
			if cfgErr != nil {
				err = cfgErr
				t.Logf("Attempt %d: LoadConfig() error = %v", i+1, cfgErr)
			} else {
				// Intentar inicializar la DB
				dbConn, err = db.InitDB(cfg)
				if err == nil {
					t.Logf("Attempt %d: Successfully connected to the database", i+1)
					break
				}
				t.Logf("Attempt %d: InitDB() error = %v", i+1, err)
			}
			time.Sleep(retryInterval)
		}

		// 4. Verificar que la conexión finalmente se estableció
		if err != nil {
			t.Fatalf("Failed to connect to the database after %d retries: %v", maxRetries, err)
		}

		// 5. Verificar que podemos hacer ping
		if dbConn != nil {
			sqlDB, err := dbConn.DB()
			if err != nil {
				t.Fatalf("Failed to get *sql.DB: %v", err)
			}
			if errPing := sqlDB.Ping(); errPing != nil {
				t.Fatalf("Ping() error: %v", errPing)
			}
		}

		// 6. Clean up: unsetear variables de entorno
		_ = os.Unsetenv("DB_HOST")
		_ = os.Unsetenv("DB_PORT")
		_ = os.Unsetenv("DB_USER")
		_ = os.Unsetenv("DB_PASSWORD")
		_ = os.Unsetenv("DB_NAME")
		_ = os.Unsetenv("DB_SSL_MODE")
		_ = os.Unsetenv("DB_MAX_IDLE_CONNS")
		_ = os.Unsetenv("DB_MAX_OPEN_CONNS")
		_ = os.Unsetenv("DB_CONN_MAX_LIFETIME")
	})
}
