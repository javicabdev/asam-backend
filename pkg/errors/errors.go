package errors

import "fmt"

// ErrorCode representa los códigos de error del sistema
type ErrorCode string

const (
	// Errores de Validación
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	//ErrInvalidFormat    ErrorCode = "INVALID_FORMAT"
	//ErrInvalidDate      ErrorCode = "INVALID_DATE"
	//ErrInvalidAmount    ErrorCode = "INVALID_AMOUNT"
	//ErrInvalidStatus    ErrorCode = "INVALID_STATUS"

	// Errores de Negocio
	//ErrDuplicateEntry    ErrorCode = "DUPLICATE_ENTRY"
	ErrNotFound ErrorCode = "NOT_FOUND"
	//ErrInvalidOperation  ErrorCode = "INVALID_OPERATION"
	//ErrInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"

	// Errores de Sistema
	ErrDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrInternalError ErrorCode = "INTERNAL_ERROR"
	//ErrNetworkError  ErrorCode = "NETWORK_ERROR"

	// Errores de Autenticación
	//ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	//ErrForbidden    ErrorCode = "FORBIDDEN"
	//ErrInvalidToken ErrorCode = "INVALID_TOKEN"
)

// AppError representa un error de la aplicación
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
	Fields  map[string]string
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewValidationError crea un error de validación
func NewValidationError(message string, fields map[string]string) *AppError {
	return &AppError{
		Code:    ErrValidationFailed,
		Message: message,
		Fields:  fields,
	}
}

// NewNotFoundError crea un error de recurso no encontrado
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewBusinessError crea un error de negocio
func NewBusinessError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewDatabaseError crea un error de base de datos
func NewDatabaseError(message string, err error) *AppError {
	return &AppError{
		Code:    ErrDatabaseError,
		Message: message,
		Cause:   err,
	}
}

func (e *AppError) Unwrap() error {
	return e.Cause
}
