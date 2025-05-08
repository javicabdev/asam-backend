package db

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Helper functions para detectar tipos específicos de errores de base de datos

// IsDuplicateKeyError detecta errores de clave duplicada
// Específico para PostgreSQL
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return (len(errMsg) >= 9 && errMsg[:9] == "ERROR: 23") ||
		(len(errMsg) >= 19 && errMsg[:19] == "duplicate key value")
}

// IsConstraintViolationError detecta errores de violación de restricciones (foreign keys, etc)
// Específico para PostgreSQL
func IsConstraintViolationError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return (len(errMsg) >= 9 && errMsg[:9] == "ERROR: 23") ||
		(len(errMsg) >= 14 && errMsg[:14] == "foreign key violation")
}

// IsNotFoundError verifica si un error es ErrRecordNotFound de GORM
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Usamos errors.Is para verificar si el error es ErrRecordNotFound
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// IsConnectionError detecta errores de conexión a la base de datos
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return contains(errMsg, "connection refused") ||
		contains(errMsg, "failed to connect") ||
		contains(errMsg, "connection reset by peer")
}

// Contains verifica si una cadena contiene otra
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
