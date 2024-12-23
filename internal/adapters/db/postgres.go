package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/javicabdev/asam-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the database connection.
// Recibe una instancia de config.Config con los valores de entorno ya cargados.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound
			Colorful:                  true,        // Enable color
		},
	)

	// Construct DSN (Data Source Name) utilizando los valores en cfg
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
	fmt.Println("[DEBUG] DSN = ", dsn)

	// Open connection with GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("error getting database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)       // Mantener conexiones inactivas
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)       // Máximo de conexiones simultáneas
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime) // Tiempo de vida de las conexiones

	// Configurar statement cache
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Habilitar cache de prepared statements
		NowFunc: func() time.Time {
			return time.Now().UTC() // Use UTC for all timestamps
		},
		Logger: logger.Default.LogMode(logger.Silent), // Reducir logging en producción
	})

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}
