package output

import "context"

type DBHandler interface {
	Connect(ctx context.Context) error
	Close() error
	// Aquí añadiremos métodos para operaciones de base de datos
}

type DBConfig interface {
	GetMaxIdleConns() int
	GetMaxOpenConns() int
	GetConnMaxLifetime() int
	GetMaxRetries() int
	GetRetryInterval() int
}
