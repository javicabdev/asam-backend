package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/pkg/errors"
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
		return nil, errors.DB(err, "error connecting to database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.InternalError("error getting database instance", err)
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
	if err != nil {
		return nil, errors.DB(err, "error reopening database with prepared statements")
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, errors.DB(err, "error pinging database")
	}

	return db, nil
}
