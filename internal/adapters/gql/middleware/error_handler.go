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

// ErrorHandlerKey is a typed key for context
type ErrorHandlerKey struct{}

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

// ServeHTTP implements http.Handler
func (m *ErrorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create a new context with error handler
	ctx := context.WithValue(r.Context(), ErrorHandlerKey{}, m)

	// Execute next handler with the new context
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// HandleError processes an error and transforms it for GraphQL
func (m *ErrorMiddleware) HandleError(ctx context.Context, err error) *gqlerror.Error {
	// If already a GraphQL error, return it
	var gqlErr *gqlerror.Error
	if errors.As(err, &gqlErr) {
		if m.logger != nil {
			m.logError(ctx, "GraphQL error", gqlErr.Message, "graphql_error", gqlErr.Extensions["code"], gqlErr.Path)
		}
		return gqlErr
	}

	// Convert to AppError if possible
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		return m.handleAppError(ctx, appErr)
	}

	// Handle common errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m.createNotFoundError(ctx)
	}

	// For other errors, create a generic internal error
	return m.createInternalError(ctx, err)
}

// Log levels for different error types
const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

// Map AppError codes to logging levels
func (m *ErrorMiddleware) getErrorLevel(code appErrors.ErrorCode) string {
	switch code {
	// Debug level - Client errors that are expected
	case appErrors.ErrValidationFailed,
		appErrors.ErrInvalidFormat,
		appErrors.ErrInvalidDate,
		appErrors.ErrInvalidAmount,
		appErrors.ErrInvalidStatus,
		appErrors.ErrNotFound:
		return levelDebug

	// Warn level - Important but expected errors
	case appErrors.ErrUnauthorized,
		appErrors.ErrForbidden,
		appErrors.ErrDuplicateEntry,
		appErrors.ErrInvalidOperation,
		appErrors.ErrInsufficientFunds,
		appErrors.ErrInvalidToken,
		appErrors.ErrInvalidRequest,
		appErrors.ErrRateLimitExceeded:
		return levelWarn

	// Error level - System errors that need attention
	case appErrors.ErrDatabaseError,
		appErrors.ErrInternalError,
		appErrors.ErrNetworkError:
		return levelError

	default:
		return levelWarn
	}
}

// handleAppError handles errors of type AppError
func (m *ErrorMiddleware) handleAppError(ctx context.Context, err *appErrors.AppError) *gqlerror.Error {
	if m.logger != nil {
		level := m.getErrorLevel(err.Code)
		m.logError(ctx, string(err.Code), err.Message, level, err.Code, graphql.GetPath(ctx))
	}

	// Create GraphQL error
	path := graphql.GetPath(ctx)
	extensions := map[string]any{
		"code": err.Code,
	}

	// Add validation fields if they exist
	if len(err.Fields) > 0 {
		extensions["fields"] = err.Fields
	}

	return &gqlerror.Error{
		Path:       path,
		Message:    err.Message,
		Extensions: extensions,
	}
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

	// Add operation name if available
	if op := graphql.GetOperationContext(ctx); op != nil {
		fields = append(fields, zap.String("operation", op.OperationName))
	}

	// Log at the appropriate level
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
