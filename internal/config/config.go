package config

import (
	"context"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

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
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=60m"` // ← Cambiar a time.Duration
}

// LoadConfig usa godotenv y go-envconfig para cargar la configuración.
func LoadConfig() (*Config, error) {
	// Carga .env (silenciosamente, en local, sin error si no existe)
	_ = godotenv.Load()

	ctx := context.Background()
	var c Config
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}
	return &c, nil
}
