// internal/adapters/gql/middleware/recovery.go

package middleware

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

type RecoveryMiddleware struct {
	logger *zap.Logger
	next   http.Handler
}

func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

func (m *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

func (m *RecoveryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			// Log el panic con el stack trace
			stack := debug.Stack()
			m.logger.Panic("PANIC",
				zap.Any("error", err),
				zap.String("stack", string(stack)),
			)

			// Crear error de aplicación
			appErr := errors.NewBusinessError(
				errors.ErrInternalError,
				fmt.Sprintf("Internal server error: %v", err),
			)

			// Establecer el error en el contexto para que lo maneje el error middleware
			ctx := context.WithValue(r.Context(), "error", appErr)
			r = r.WithContext(ctx)

			// Responder con error 500
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()

	m.next.ServeHTTP(w, r)
}
