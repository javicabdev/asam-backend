package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"gorm.io/gorm"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

type ErrorHandlerKey struct{}
type HttpRequestKey struct{}

// ErrorMiddleware handles application errors
type ErrorMiddleware struct {
	logger logger.Logger
	next   http.Handler
}

// NewErrorMiddleware creates a new error middleware
func NewErrorMiddleware(logger logger.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

// Handler sets the next handler in the chain
func (m *ErrorMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

// ServeHTTP implements http.Handler.
func (m *ErrorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORRECCIÓN: Usar el tipo exportado
	ctx := context.WithValue(r.Context(), ErrorHandlerKey{}, m)
	// CORRECCIÓN: Usar el tipo exportado
	ctx = context.WithValue(ctx, HttpRequestKey{}, r)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// HandleError processes an error and transforms it for GraphQL
func (m *ErrorMiddleware) HandleError(ctx context.Context, err error) *gqlerror.Error {
	// Si ya es un gqlerror.Error, intentar extraer el AppError subyacente
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		// Intentar extraer el AppError del error original de gqlErr
		if gqlErr.Err != nil {
			var appErr *appErrors.AppError
			if errors.As(gqlErr.Err, &appErr) {
				// Encontramos el AppError original, procesarlo
				return m.handleAppError(ctx, appErr)
			}
		}
		// Si no hay AppError subyacente, loggear y retornar el gqlErr tal cual
		if m.logger != nil {
			m.logError(ctx, "GraphQL error", gqlErr.Message, "graphql_error", gqlErr.Extensions["code"], gqlErr.Path)
		}
		return gqlErr
	}

	// Intentar como AppError directamente
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		return m.handleAppError(ctx, appErr)
	}

	// Error de GORM (registro no encontrado)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m.createNotFoundError(ctx)
	}

	// Error genérico/desconocido
	return m.createInternalError(ctx, err)
}

// Log levels
const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

// getErrorLevel maps AppError codes to logging levels
func (m *ErrorMiddleware) getErrorLevel(code appErrors.ErrorCode) string {
	switch code {
	case appErrors.ErrValidationFailed, appErrors.ErrNotFound:
		return levelDebug
	case appErrors.ErrUnauthorized, appErrors.ErrForbidden, appErrors.ErrDuplicateEntry, appErrors.ErrInvalidToken:
		return levelWarn
	default:
		return levelError
	}
}

// handleAppError handles errors of type AppError
func (m *ErrorMiddleware) handleAppError(ctx context.Context, err *appErrors.AppError) *gqlerror.Error {
	// Create extensions with the error code
	extensions := map[string]any{"code": string(err.Code)}

	// Add fields if they exist
	if len(err.Fields) > 0 {
		extensions["fields"] = err.Fields
	}

	// Detailed logging for debugging
	if m.logger != nil {
		level := m.getErrorLevel(err.Code)
		logFields := []zap.Field{
			zap.String("error_code", string(err.Code)),
			zap.String("message", err.Message),
			zap.Any("path", graphql.GetPath(ctx)),
			zap.Int("fields_count", len(err.Fields)),
		}

		// Log fields if they exist
		if len(err.Fields) > 0 {
			logFields = append(logFields, zap.Any("error_fields", err.Fields))
		}

		switch level {
		case levelDebug:
			m.logger.Debug("GraphQL validation error with fields", logFields...)
		case levelWarn:
			m.logger.Warn("GraphQL error with fields", logFields...)
		default:
			m.logger.Error("GraphQL error with fields", logFields...)
		}
	}

	// Create the GraphQL error
	gqlError := &gqlerror.Error{
		Message:    err.Message,
		Path:       graphql.GetPath(ctx),
		Extensions: extensions,
	}

	// Log the complete serialized error for verification
	if m.logger != nil && len(err.Fields) > 0 {
		m.logger.Debug("Created GraphQL error",
			zap.String("message", gqlError.Message),
			zap.Any("extensions", gqlError.Extensions),
		)
	}

	return gqlError
}

// logError logs errors in a standardized format
func (m *ErrorMiddleware) logError(ctx context.Context, errType, message, level string, code any, path any) {
	if m.logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("error_type", errType),
		zap.String("message", message),
	}
	if code != nil {
		fields = append(fields, zap.Any("code", code))
	}
	if path != nil {
		fields = append(fields, zap.Any("path", path))
	}

	// CORRECCIÓN: Usar el tipo exportado para extraer el request del contexto
	if httpReq, ok := ctx.Value(HttpRequestKey{}).(*http.Request); ok {
		if operationName, err := getOperationName(httpReq); err == nil && operationName != "" {
			fields = append(fields, zap.String("operation", operationName))
		}
	}

	switch level {
	case levelDebug:
		m.logger.Debug("GraphQL error", fields...)
	case levelInfo:
		m.logger.Info("GraphQL error", fields...)
	case levelWarn:
		m.logger.Warn("GraphQL error", fields...)
	case levelError:
		m.logger.Error("GraphQL error", fields...)
	default:
		m.logger.Warn("GraphQL error (unknown level)", fields...)
	}
}

// createNotFoundError creates a standardized "not found" error
func (m *ErrorMiddleware) createNotFoundError(ctx context.Context) *gqlerror.Error {
	path := graphql.GetPath(ctx)
	message := "Resource not found"
	if m.logger != nil {
		m.logError(ctx, "Not found", message, levelDebug, appErrors.ErrNotFound, path)
	}
	return &gqlerror.Error{
		Path:    path,
		Message: message,
		Extensions: map[string]any{
			"code": appErrors.ErrNotFound,
		},
	}
}

// createInternalError creates a standardized internal error
func (m *ErrorMiddleware) createInternalError(ctx context.Context, err error) *gqlerror.Error {
	path := graphql.GetPath(ctx)
	message := "Internal server error"
	if m.logger != nil {
		m.logger.Error("Unhandled error",
			zap.Error(err),
			zap.Any("path", path),
		)
	}
	return &gqlerror.Error{
		Path:    path,
		Message: message,
		Extensions: map[string]any{
			"code": appErrors.ErrInternalError,
		},
	}
}
