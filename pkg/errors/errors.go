package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode representa los códigos de error del sistema
type ErrorCode string

const (
	// ErrValidationFailed error de validación general
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrInvalidFormat    ErrorCode = "INVALID_FORMAT"
	ErrInvalidDate      ErrorCode = "INVALID_DATE"
	ErrInvalidAmount    ErrorCode = "INVALID_AMOUNT"
	ErrInvalidStatus    ErrorCode = "INVALID_STATUS"

	// ErrDuplicateEntry error de entrada duplicada
	ErrDuplicateEntry    ErrorCode = "DUPLICATE_ENTRY"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrInvalidOperation  ErrorCode = "INVALID_OPERATION"
	ErrInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"

	// ErrDatabaseError error de base de datos
	ErrDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrInternalError ErrorCode = "INTERNAL_ERROR"
	ErrNetworkError  ErrorCode = "NETWORK_ERROR"

	// ErrUnauthorized error de autenticación
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrForbidden    ErrorCode = "FORBIDDEN"
	ErrInvalidToken ErrorCode = "INVALID_TOKEN"
)

// AppError representa un error de la aplicación
type AppError struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Cause   error             `json:"-"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// Error implementa la interfaz error
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implementa la interfaz unwrapper para errors.Is y errors.As
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithCause añade o reemplaza la causa subyacente del error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithField añade o reemplaza un campo de validación
func (e *AppError) WithField(key, value string) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[key] = value
	return e
}

// IsAppError verifica si un error es de tipo AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError convierte un error a AppError si es posible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Is verifica si un error es de un código específico
func Is(err error, code ErrorCode) bool {
	appErr, ok := AsAppError(err)
	return ok && appErr.Code == code
}

// IsValidationError verifica si es un error de validación
func IsValidationError(err error) bool {
	return Is(err, ErrValidationFailed)
}

// IsNotFoundError verifica si es un error de recurso no encontrado
func IsNotFoundError(err error) bool {
	return Is(err, ErrNotFound)
}

// IsAuthError verifica si es un error de autenticación
func IsAuthError(err error) bool {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code == ErrUnauthorized ||
			appErr.Code == ErrForbidden ||
			appErr.Code == ErrInvalidToken
	}
	return false
}

// IsDatabaseError verifica si es un error de base de datos
func IsDatabaseError(err error) bool {
	return Is(err, ErrDatabaseError)
}

// GetFields obtiene los campos de un error de validación, si existen
func GetFields(err error) map[string]string {
	appErr, ok := AsAppError(err)
	if !ok {
		return nil
	}
	return appErr.Fields
}

// GetCode obtiene el código de error, si existe
func GetCode(err error) (ErrorCode, bool) {
	appErr, ok := AsAppError(err)
	if !ok {
		return "", false
	}
	return appErr.Code, true
}

// GetMessage obtiene el mensaje de error, si existe
func GetMessage(err error) (string, bool) {
	appErr, ok := AsAppError(err)
	if !ok {
		return "", false
	}
	return appErr.Message, true
}

// GetCause obtiene la causa subyacente del error, si existe
func GetCause(err error) (bool, error) {
	appErr, ok := AsAppError(err)
	if !ok || appErr.Cause == nil {
		return false, nil
	}
	return true, appErr.Cause
}

// New crea un nuevo AppError con el código y mensaje dados
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Fields:  make(map[string]string),
	}
}

// Wrap envuelve un error existente en un AppError
func Wrap(err error, code ErrorCode, message string) *AppError {
	// Si el error es nulo, creamos uno nuevo sin causa
	if err == nil {
		return New(code, message)
	}

	// Si ya es un AppError, solo actualizamos código y mensaje si se proporcionan
	if appErr, ok := AsAppError(err); ok {
		if code != "" {
			appErr.Code = code
		}
		if message != "" {
			appErr.Message = message
		}
		return appErr
	}

	// Crear nuevo AppError
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Fields:  make(map[string]string),
	}
}

// AddField añade un campo a un error existente o crea un nuevo error de validación
func AddField(err error, fieldName, fieldValue string) error {
	if err == nil {
		return nil
	}

	if appErr, ok := AsAppError(err); ok {
		if appErr.Fields == nil {
			appErr.Fields = make(map[string]string)
		}
		appErr.Fields[fieldName] = fieldValue
		return appErr
	}

	return Validation(err.Error(), fieldName, fieldValue)
}

// FromError convierte un error estándar a un AppError
func FromError(err error) error {
	if err == nil {
		return nil
	}

	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	return InternalError("internal error", err)
}

// extractFieldsFromAppError extrae los campos de un AppError si existen
func extractFieldsFromAppError(appErr *AppError, fields map[string]string) {
	if appErr.Fields != nil {
		for k, v := range appErr.Fields {
			fields[k] = v
		}
	}
}

// findBaseError encuentra el primer error no nulo para usar como base
func findBaseError(errs []error) (*AppError, map[string]string) {
	fields := make(map[string]string)

	for _, err := range errs {
		if err == nil {
			continue
		}

		if appErr, ok := AsAppError(err); ok {
			extractFieldsFromAppError(appErr, fields)
			return appErr, fields
		}

		// Si no es AppError, crear uno nuevo
		return New(ErrValidationFailed, err.Error()), fields
	}

	return nil, fields
}

// collectAdditionalFields recolecta campos de los errores adicionales
func collectAdditionalFields(errs []error, fields map[string]string) {
	// Saltamos el primer error que ya se procesó como base
	firstErrorProcessed := false

	for _, err := range errs {
		if err == nil {
			continue
		}

		if !firstErrorProcessed {
			firstErrorProcessed = true
			continue
		}

		if appErr, ok := AsAppError(err); ok && appErr.Fields != nil {
			extractFieldsFromAppError(appErr, fields)
		}
	}
}

// Combine une varios errores en uno solo, combinando sus campos si son de validación
func Combine(errs ...error) error {
	baseError, fields := findBaseError(errs)

	if baseError == nil {
		return nil
	}

	collectAdditionalFields(errs, fields)

	// Asignamos los campos combinados
	baseError.Fields = fields
	return baseError
}

// FromHTTPStatus crea un AppError a partir de un código HTTP
func FromHTTPStatus(status int, message string) error {
	switch {
	case status >= 400 && status < 500:
		switch status {
		case http.StatusUnauthorized:
			return NewUnauthorizedError()
		case http.StatusForbidden:
			return NewBusinessError(ErrForbidden, message)
		case http.StatusNotFound:
			return NewNotFoundError(message)
		case http.StatusBadRequest:
			return NewValidationError(message, nil)
		default:
			return NewBusinessError(ErrInvalidOperation, message)
		}
	case status >= 500:
		return InternalError(message, nil)
	default:
		return nil
	}
}

// NewValidationError crea un error de validación
func NewValidationError(message string, fields map[string]string) error {
	return &AppError{
		Code:    ErrValidationFailed,
		Message: message,
		Fields:  fields,
	}
}

// NewNotFoundError crea un error de recurso no encontrado
func NewNotFoundError(resource string) error {
	return &AppError{
		Code:    ErrNotFound,
		Message: fmt.Sprintf("%s no encontrada", resource),
	}
}

// NewBusinessError crea un error de negocio
func NewBusinessError(code ErrorCode, message string) error {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewDatabaseError crea un error de base de datos
func NewDatabaseError(message string, err error) error {
	return &AppError{
		Code:    ErrDatabaseError,
		Message: message,
		Cause:   err,
	}
}

// NewUnauthorizedError crea un error de autorización
func NewUnauthorizedError() error {
	return &AppError{
		Code:    ErrUnauthorized,
		Message: "Unauthorized access to the resource",
	}
}

// DB wraps database-specific errors into structured AppErrors
func DB(err error, message string) error {
	if err == nil {
		return nil
	}

	// Already an AppError
	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	return NewDatabaseError(message, err)
}

// NotFound returns a standardized not found error
func NotFound(resource string, err error) error {
	notFoundErr := NewNotFoundError(resource)
	if err != nil {
		if appErr, ok := notFoundErr.(*AppError); ok {
			appErr.Cause = err
		}
	}
	return notFoundErr
}

// Validation returns a standardized validation error
func Validation(message string, fieldName, fieldError string) error {
	fields := make(map[string]string)
	if fieldName != "" {
		fields[fieldName] = fieldError
	}
	return NewValidationError(message, fields)
}

// AuthError returns a standardized authentication error
func AuthError(code ErrorCode, message string, err error) error {
	authErr := &AppError{
		Code:    code,
		Message: message,
	}
	if err != nil {
		authErr.Cause = err
	}
	return authErr
}

// NetworkError returns a standardized network error
func NetworkError(message string, err error) error {
	return &AppError{
		Code:    ErrNetworkError,
		Message: message,
		Cause:   err,
	}
}

// IsNetworkError verifica si es un error de red
func IsNetworkError(err error) bool {
	return Is(err, ErrNetworkError)
}

// TokenError returns a standardized token error
func TokenError(message string, err error) error {
	return &AppError{
		Code:    ErrInvalidToken,
		Message: message,
		Cause:   err,
	}
}

// InternalError returns a standardized internal error
func InternalError(message string, err error) error {
	return &AppError{
		Code:    ErrInternalError,
		Message: message,
		Cause:   err,
	}
}

// Business returns a standardized business error
func Business(code ErrorCode, message string, err error) error {
	businessErr := &AppError{
		Code:    code,
		Message: message,
	}
	if err != nil {
		businessErr.Cause = err
	}
	return businessErr
}
