package db

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
