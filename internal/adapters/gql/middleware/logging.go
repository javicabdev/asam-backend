package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
)

func LoggingMiddleware() graphql.ResponseMiddleware {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		start := time.Now()
		resp := next(ctx)

		// Extraer información de la operación
		op := graphql.GetOperationContext(ctx)

		// Convertir Operation a string
		operationType := "unknown"
		if op.Operation != nil {
			operationType = string(op.Operation.Operation)
		}

		// Preparar campos para el log
		fields := []zap.Field{
			zap.String("operation_name", op.OperationName),
			zap.String("operation_type", operationType),
			zap.Duration("duration", time.Since(start)),
		}

		// Añadir errores si existen
		if len(resp.Errors) > 0 {
			fields = append(fields, zap.Any("errors", resp.Errors))
			logger.Error("GraphQL operation completed with errors", fields...)
		} else {
			logger.Info("GraphQL operation completed successfully", fields...)
		}

		return resp
	}
}
