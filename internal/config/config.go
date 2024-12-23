package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Config define las variables de entorno que tu aplicación necesita.
type Config struct {
	DBHost     string `env:"DB_HOST,default=localhost"`
	DBPort     string `env:"DB_PORT,default=5432"`
	DBUser     string `env:"DB_USER,default=postgres"`
	DBPassword string `env:"DB_PASSWORD,default=postgres"`
	DBName     string `env:"DB_NAME,default=asam_db"`
	DBSSLMode  string `env:"DB_SSL_MODE,default=disable"`

	// Pool de conexiones
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS,default=10"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS,default=100"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=60m"`
}

// LoadConfig carga las variables de entorno y las mapea a la estructura Config.
func LoadConfig() (*Config, error) {
	// Determinar el entorno actual
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Valor por defecto
	}

	// Cargar el archivo .env correspondiente
	envFile := fmt.Sprintf(".env.%s", env)
	err := godotenv.Load(envFile)
	if err != nil {
		fmt.Printf("[WARNING] No se pudo cargar el archivo %s: %v\n", envFile, err)
		// No retornamos error si el archivo no existe; asumimos que las variables están establecidas
	}

	ctx := context.Background()
	var c Config
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}
	return &c, nil
}
