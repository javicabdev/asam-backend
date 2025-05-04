package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gorm.io/gorm"
	"net/http"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
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

// FromContext gets the ErrorMiddleware from context
func FromContext(ctx context.Context) *ErrorMiddleware {
	if handler, ok := ctx.Value(ErrorHandlerKey{}).(*ErrorMiddleware); ok {
		return handler
	}

	// Default handler if not found in context
	return &ErrorMiddleware{
		logger: nil, // We should avoid this in production
	}
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
	case appErrors.ErrValidationFailed, appErrors.ErrInvalidFormat, appErrors.ErrNotFound:
		return levelDebug
	case appErrors.ErrUnauthorized, appErrors.ErrForbidden, appErrors.ErrDuplicateEntry:
		return levelWarn
	case appErrors.ErrDatabaseError, appErrors.ErrInternalError:
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
	if err.Fields != nil && len(err.Fields) > 0 {
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

// PresentError is a wrapper for integration with gqlgen
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	handler := FromContext(ctx)
	return handler.HandleError(ctx, err)
}

// ConfigErrorPresenter configures the error presenter in gqlgen
func ConfigErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		return PresentError(ctx, err)
	}
}

// LogPanic is a function to handle panic in resolvers
func LogPanic() graphql.RecoverFunc {
	return func(ctx context.Context, err any) error {
		handler := FromContext(ctx)

		// Log panic
		if handler.logger != nil {
			handler.logger.Error("GraphQL panic recovered",
				zap.Any("error", err),
				zap.String("stack", fmt.Sprintf("%+v", err)),
				zap.Any("path", graphql.GetPath(ctx)))
		}

		// Create a GraphQL error for the panic
		return &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Internal server error",
			Extensions: map[string]any{
				"code": appErrors.ErrInternalError,
			},
		}
	}
}
