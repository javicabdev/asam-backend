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

// ErrorHandlerKey es la clave tipada para el contexto
type ErrorHandlerKey struct{}

// ErrorMiddleware maneja los errores de la aplicación
type ErrorMiddleware struct {
	logger logger.Logger
	next   http.Handler
}

// NewErrorMiddleware crea un nuevo middleware de error
func NewErrorMiddleware(logger logger.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

// Handler establece el siguiente handler en la cadena
func (m *ErrorMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

// ServeHTTP implementa http.Handler
func (m *ErrorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Crear un nuevo contexto con error handler
	ctx := context.WithValue(r.Context(), ErrorHandlerKey{}, m)

	// Ejecutar el siguiente handler con el nuevo contexto
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// FromContext obtiene el ErrorMiddleware desde el contexto
func FromContext(ctx context.Context) *ErrorMiddleware {
	if handler, ok := ctx.Value(ErrorHandlerKey{}).(*ErrorMiddleware); ok {
		return handler
	}
	// Default handler si no se encuentra en el contexto
	return &ErrorMiddleware{
		logger: nil, // Esto deberíamos evitarlo en producción
	}
}

// HandleError procesa un error y lo transforma para GraphQL
func (m *ErrorMiddleware) HandleError(ctx context.Context, err error) *gqlerror.Error {
	// Si ya es un error GraphQL, retornarlo
	if gqlErr, ok := err.(*gqlerror.Error); ok {
		m.logGraphQLError(gqlErr)
		return gqlErr
	}

	// Convertir a AppError si es posible
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		return m.handleAppError(ctx, appErr)
	}

	// Manejar errores comunes
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return m.handleNotFound(ctx, err)
	}

	// Para otros errores, crear un error genérico interno
	return m.handleUnknownError(ctx, err)
}

// handleAppError maneja errores de tipo AppError
func (m *ErrorMiddleware) handleAppError(ctx context.Context, err *appErrors.AppError) *gqlerror.Error {
	// Log según el tipo de error
	switch err.Code {
	case appErrors.ErrValidationFailed, appErrors.ErrInvalidFormat:
		m.logger.Debug("Validation error",
			zap.String("code", string(err.Code)),
			zap.Any("fields", err.Fields),
			zap.Error(err))

	case appErrors.ErrNotFound:
		m.logger.Debug("Resource not found",
			zap.String("message", err.Message),
			zap.Error(err))

	case appErrors.ErrDatabaseError:
		m.logger.Error("Database error",
			zap.String("message", err.Message),
			zap.Error(err))

	case appErrors.ErrInternalError:
		m.logger.Error("Internal server error",
			zap.String("message", err.Message),
			zap.Error(err))

	case appErrors.ErrUnauthorized, appErrors.ErrForbidden:
		m.logger.Warn("Authorization error",
			zap.String("code", string(err.Code)),
			zap.String("message", err.Message),
			zap.Error(err))

	default:
		m.logger.Warn("Unhandled error code",
			zap.String("code", string(err.Code)),
			zap.String("message", err.Message),
			zap.Error(err))
	}

	// Crear error GraphQL
	path := graphql.GetPath(ctx)

	return &gqlerror.Error{
		Path:    path,
		Message: err.Message,
		Extensions: map[string]interface{}{
			"code":   err.Code,
			"fields": err.Fields,
		},
	}
}

// handleNotFound maneja errores de tipo "not found"
func (m *ErrorMiddleware) handleNotFound(ctx context.Context, err error) *gqlerror.Error {
	if m.logger != nil {
		m.logger.Debug("Resource not found", zap.Error(err))
	}

	path := graphql.GetPath(ctx)

	return &gqlerror.Error{
		Path:    path,
		Message: "Resource not found",
		Extensions: map[string]interface{}{
			"code": appErrors.ErrNotFound,
		},
	}
}

// handleUnknownError maneja errores desconocidos
func (m *ErrorMiddleware) handleUnknownError(ctx context.Context, err error) *gqlerror.Error {
	if m.logger != nil {
		m.logger.Error("Unhandled error", zap.Error(err))
	}

	path := graphql.GetPath(ctx)

	return &gqlerror.Error{
		Path:    path,
		Message: "Internal server error",
		Extensions: map[string]interface{}{
			"code": appErrors.ErrInternalError,
		},
	}
}

// logGraphQLError registra errores GraphQL
func (m *ErrorMiddleware) logGraphQLError(err *gqlerror.Error) {
	if m.logger == nil {
		return
	}

	code := "UNKNOWN"
	if codeExt, exists := err.Extensions["code"]; exists {
		if codeStr, ok := codeExt.(string); ok {
			code = codeStr
		}
	}

	switch code {
	case string(appErrors.ErrValidationFailed), string(appErrors.ErrInvalidFormat):
		m.logger.Debug("GraphQL validation error",
			zap.String("code", code),
			zap.Any("path", err.Path),
			zap.Any("extensions", err.Extensions))

	case string(appErrors.ErrNotFound):
		m.logger.Debug("GraphQL resource not found",
			zap.String("message", err.Message),
			zap.Any("path", err.Path))

	case string(appErrors.ErrDatabaseError), string(appErrors.ErrInternalError):
		m.logger.Error("GraphQL server error",
			zap.String("code", code),
			zap.String("message", err.Message),
			zap.Any("path", err.Path))

	default:
		m.logger.Warn("GraphQL unhandled error",
			zap.String("code", code),
			zap.String("message", err.Message),
			zap.Any("path", err.Path))
	}
}

// PresentError es un wrapper para integración con gqlgen
func PresentError(ctx context.Context, err error) *gqlerror.Error {
	handler := FromContext(ctx)
	return handler.HandleError(ctx, err)
}

// ConfigErrorPresenter configura el presentador de errores en gqlgen
func ConfigErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, err error) *gqlerror.Error {
		return PresentError(ctx, err)
	}
}

// LogPanic es una función para manejar pánico en los resolvers
func LogPanic() graphql.RecoverFunc {
	return func(ctx context.Context, err interface{}) error {
		handler := FromContext(ctx)

		// Log de pánico
		if handler.logger != nil {
			handler.logger.Error("GraphQL panic recovered",
				zap.Any("error", err),
				zap.String("stack", fmt.Sprintf("%+v", err)),
				zap.Any("path", graphql.GetPath(ctx)))
		}

		// Crear un error GraphQL para el pánico
		return &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Internal server error",
			Extensions: map[string]interface{}{
				"code": "INTERNAL_SERVER_ERROR",
			},
		}
	}
}
