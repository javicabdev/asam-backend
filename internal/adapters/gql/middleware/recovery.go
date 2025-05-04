package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recupera de pánicos y los convierte en errores manejables
type RecoveryMiddleware struct {
	logger logger.Logger
	next   http.Handler
}

// NewRecoveryMiddleware crea un nuevo middleware de recuperación
func NewRecoveryMiddleware(logger logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Handler establece el siguiente handler en la cadena
func (m *RecoveryMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

// ServeHTTP implementa http.Handler
func (m *RecoveryMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if recovered := recover(); recovered != nil {
			// Capturar stack trace
			stack := debug.Stack()

			// Log detallado del pánico con campos estructurados
			m.logger.Error("Panic recovered in HTTP handler",
				zap.Any("error", recovered),
				zap.String("stack", string(stack)),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr),
			)

			// Crear un AppError estructurado con detalles del pánico
			appErr := errors.New(
				errors.ErrInternalError,
				"Internal server error",
			).WithCause(fmt.Errorf("%v", recovered))

			// Buscar el error handler en el contexto
			var handler *ErrorMiddleware
			if h, ok := r.Context().Value(ErrorHandlerKey{}).(*ErrorMiddleware); ok {
				handler = h
			} else {
				// Si no hay handler en el contexto, usar uno por defecto
				handler = NewErrorMiddleware(m.logger)
			}

			// Convertir el error del pánico a un error GraphQL estructurado
			graphQLErr := handler.HandleError(r.Context(), appErr)

			// Establecer código de respuesta y encabezados
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)

			// Escribir respuesta JSON con el error estructurado
			responseJSON := map[string]any{
				"errors": []any{
					map[string]any{
						"message": graphQLErr.Message,
						"path":    graphQLErr.Path,
						"extensions": map[string]any{
							"code": errors.ErrInternalError,
						},
					},
				},
			}

			// Escribir respuesta JSON
			if err := json.NewEncoder(w).Encode(responseJSON); err != nil {
				m.logger.Error("Failed to encode error response",
					zap.Error(err),
				)
			}
		}
	}()

	// Ejecutar el siguiente handler
	m.next.ServeHTTP(w, r)
}
