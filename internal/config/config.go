package config

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"time"
)

// Config define las variables de entorno que tu aplicación necesita.
type Config struct {
	Environment string `env:"ENVIRONMENT,default=development"`
	DBHost      string `env:"DB_HOST,default=localhost"`
	DBPort      string `env:"DB_PORT,default=5432"`
	DBUser      string `env:"DB_USER,default=postgres"`
	DBPassword  string `env:"DB_PASSWORD,default=postgres"`
	DBName      string `env:"DB_NAME,default=asam_db"`
	DBSSLMode   string `env:"DB_SSL_MODE,default=disable"`

	// Pool de conexiones
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS,default=10"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS,default=100"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=1h"`

	// JWT Configuration
	JWTAccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
	JWTRefreshSecret string        `env:"JWT_REFRESH_SECRET,required"`
	JWTAccessTTL     time.Duration `env:"JWT_ACCESS_TTL,default=15m"`
	JWTRefreshTTL    time.Duration `env:"JWT_REFRESH_TTL,default=7d"`
}

// LoadConfig carga las variables de entorno y las mapea a la estructura Config.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	ctx := context.Background()
	var c Config
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, fmt.Errorf("failed to parse env config: %w", err)
	}
	return &c, nil
}
