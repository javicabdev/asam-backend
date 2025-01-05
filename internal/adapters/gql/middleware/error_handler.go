// internal/adapters/gql/middleware/error_handler.go

package middleware

import (
	"context"
	"errors"
	domainErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

type ErrorMiddleware struct {
	logger logger.Logger
	next   http.Handler
}

func NewErrorMiddleware(logger logger.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

func (m *ErrorMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

func (m *ErrorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Crear un nuevo contexto con error handler
	ctx := context.WithValue(r.Context(), "error_handler", m)

	// Ejecutar el siguiente handler con el nuevo contexto
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *ErrorMiddleware) HandleError(err error) error {
	var e *domainErrors.AppError
	switch {
	case errors.As(err, &e):
		// Loguear según el tipo de error
		switch e.Code {
		case domainErrors.ErrValidationFailed, domainErrors.ErrNotFound:
			m.logger.Debug("[DEBUG] Validation or NotFound error", zap.Error(err))
		case domainErrors.ErrDatabaseError, domainErrors.ErrInternalError:
			m.logger.Error("[ERROR] Database or Internal error", zap.Error(err))
		default:
			m.logger.Warn("[WARN] Unhandled error code", zap.Error(err))
		}
		return e
	default:
		// Error no manejado
		m.logger.Error("[ERROR] Unhandled error", zap.Error(err))
		return &domainErrors.AppError{
			Code:    domainErrors.ErrInternalError,
			Message: "Internal server error",
			Cause:   err,
		}
	}
}
