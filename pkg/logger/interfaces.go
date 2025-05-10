// Package logger proporciona interfaces y utilidades para el registro de eventos
// y errores en la aplicación de manera uniforme.
package logger

import "go.uber.org/zap"

// Logger es la interfaz genérica de logging
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Sync() error
}
