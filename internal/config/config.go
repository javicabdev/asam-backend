// Package config proporciona funcionalidades para cargar y gestionar la configuración de la aplicación
// desde variables de entorno o archivos .env.
package config

import (
	"context"
	"time"

	"github.com/joho/godotenv"
	envconfigpkg "github.com/sethvargo/go-envconfig"

	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Config define las variables de entorno que tu aplicación necesita.
type Config struct {
	// Configuración general
	Environment string `env:"ENVIRONMENT,default=development"`
	Port        string `env:"PORT,default=8080"`

	// Configuración de la base de datos
	DBHost     string `env:"DB_HOST,default=localhost"`
	DBPort     string `env:"DB_PORT,default=5432"`
	DBUser     string `env:"DB_USER,default=postgres"`
	DBPassword string `env:"DB_PASSWORD,default=postgres"`
	DBName     string `env:"DB_NAME,default=asam_db"`
	DBSSLMode  string `env:"DB_SSL_MODE,default=disable"`

	// Pool de conexiones
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS,default=10"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS,default=100"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME,default=1h"`

	// Configuración JWT
	JWTAccessSecret  string        `env:"JWT_ACCESS_SECRET,required"`
	JWTRefreshSecret string        `env:"JWT_REFRESH_SECRET,required"`
	JWTAccessTTL     time.Duration `env:"JWT_ACCESS_TTL,default=15m"`
	JWTRefreshTTL    time.Duration `env:"JWT_REFRESH_TTL,default=168h"`

	// Límites de tasa
	RateLimitRPS     float64       `env:"RATE_LIMIT_RPS,default=10"`
	RateLimitBurst   int           `env:"RATE_LIMIT_BURST,default=20"`
	RateLimitCleanup time.Duration `env:"RATE_LIMIT_CLEANUP,default=1h"`

	// Login Rate Limiting
	LoginMaxAttempts     int           `env:"LOGIN_MAX_ATTEMPTS,default=5"`
	LoginLockoutDuration time.Duration `env:"LOGIN_LOCKOUT_DURATION,default=15m"`
	LoginWindowDuration  time.Duration `env:"LOGIN_WINDOW_DURATION,default=5m"`

	// Configuraciones de notificación por email - SMTP (DEPRECATED - usar MailerSend)
	SMTPServer    string `env:"SMTP_SERVER,default=localhost"`
	SMTPPort      int    `env:"SMTP_PORT,default=587"`
	SMTPUser      string `env:"SMTP_USER"`
	SMTPPassword  string `env:"SMTP_PASSWORD"`
	SMTPUseTLS    bool   `env:"SMTP_USE_TLS,default=true"`
	SMTPFromEmail string `env:"SMTP_FROM_EMAIL,default=noreply@asam.org"`

	// Configuraciones de MailerSend (nuevo servicio de email)
	MailerSendAPIKey    string `env:"MAILERSEND_API_KEY,required"`
	MailerSendFromEmail string `env:"MAILERSEND_FROM_EMAIL,default=noreply@asam.org"`
	MailerSendFromName  string `env:"MAILERSEND_FROM_NAME,default=ASAM"`

	// Configuraciones de monitoreo y rendimiento
	EnableProfiling       bool          `env:"ENABLE_PROFILING,default=false"`
	ProfilingPort         string        `env:"PROFILING_PORT,default=6060"`
	LogSlowQueries        bool          `env:"LOG_SLOW_QUERIES,default=true"`
	SlowQueryThreshold    time.Duration `env:"SLOW_QUERY_THRESHOLD,default=100ms"`
	LogSlowResolvers      bool          `env:"LOG_SLOW_RESOLVERS,default=true"`
	SlowResolverThreshold time.Duration `env:"SLOW_RESOLVER_THRESHOLD,default=100ms"`
	MemProfileDir         string        `env:"MEM_PROFILE_DIR,default=logs/memory-profiles"`
	MemAlertThreshold     uint64        `env:"MEM_ALERT_THRESHOLD,default=200"`    // MB
	MemCriticalThreshold  uint64        `env:"MEM_CRITICAL_THRESHOLD,default=500"` // MB

	// Configuración de seguridad admin
	AdminUser     string `env:"ADMIN_USER"`
	AdminPassword string `env:"ADMIN_PASSWORD"`

	// Configuración de GraphQL
	GQLComplexityLimit     int `env:"GQL_COMPLEXITY_LIMIT,default=1000"`
	GQLConcurrentResolvers int `env:"GQL_CONCURRENT_RESOLVERS,default=10"`

	// Configuración de tokens
	MaxTokensPerUser     int           `env:"MAX_TOKENS_PER_USER,default=5"`
	TokenCleanupEnabled  bool          `env:"TOKEN_CLEANUP_ENABLED,default=true"`
	TokenCleanupInterval time.Duration `env:"TOKEN_CLEANUP_INTERVAL,default=24h"`

	// URL base de la aplicación
	BaseURL string `env:"BASE_URL,default=http://localhost:5173"`
}

// LoadConfig carga las variables de entorno y las mapea a la estructura Config.
func LoadConfig() (*Config, error) {
	// Cargar .env.production en entorno de producción
	if env := getenv("ENVIRONMENT"); env == "production" {
		_ = godotenv.Load(".env.production")
	} else {
		// Intentar cargar .env.development, luego .env.local, luego .env
		_ = godotenv.Load(".env.development")
		_ = godotenv.Load(".env.local")
		_ = godotenv.Load()
	}

	ctx := context.Background()
	var c Config
	if err := envconfigpkg.Process(ctx, &c); err != nil {
		return nil, errors.InternalError("failed to parse env config", err)
	}

	// Validar la configuración para producción
	if c.Environment == "production" {
		if c.JWTAccessSecret == "your-access-secret" || c.JWTRefreshSecret == "your-refresh-secret" {
			return nil, errors.InternalError("JWT secrets must be changed for production", nil)
		}

		if c.MailerSendAPIKey == "" {
			return nil, errors.InternalError("MailerSend API key must be set for production", nil)
		}

		if c.AdminUser == "" || c.AdminPassword == "" {
			return nil, errors.InternalError("admin credentials must be set for production", nil)
		}
	}

	return &c, nil
}

// getenv returns the environment variable or fallback if not set
func getenv(key string) string {
	value, exists := context.Background().Value(key).(string)
	if exists {
		return value
	}
	return ""
}
