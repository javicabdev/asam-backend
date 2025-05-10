package output

import "context"

// DBHandler define las operaciones básicas de la base de datos
type DBHandler interface {
	Connect(ctx context.Context) error
	Close() error
	// Aquí añadiremos métodos para operaciones de base de datos
}

// DBConfig define la configuración de la base de datos
type DBConfig interface {
	GetMaxIdleConns() int
	GetMaxOpenConns() int
	GetConnMaxLifetime() int
	GetMaxRetries() int
	GetRetryInterval() int
}
