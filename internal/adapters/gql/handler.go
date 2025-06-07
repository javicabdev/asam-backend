// Package gql provides GraphQL server implementation for the ASAM backend.
// It includes handlers, middleware configuration, and integration with the resolver layer.
package gql

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/generated"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

// corsMiddleware es un middleware simple para CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configurar headers CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, authorization")

		// Manejar preflighted requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NewHandler configura y retorna un nuevo handler GraphQL
func NewHandler(
	authService input.AuthService,
	resolver *resolvers.Resolver,
	cfg *config.Config,
	appLogger logger.Logger,
	db *gorm.DB,
) http.Handler {
	// Crear un errorHandler compartido
	errorHandler := middleware.NewErrorMiddleware(appLogger)

	// Crear el schema GraphQL
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	})

	// Configurar el servidor GraphQL
	srv := handler.New(schema)

	// Configurar error presenter que usa el mismo errorHandler
	srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
		// Asegurar que se guarde el errorHandler en el contexto para uso posterior
		ctx = context.WithValue(ctx, middleware.ErrorHandlerKey{}, errorHandler)
		return errorHandler.HandleError(ctx, err)
	})

	// Configurar función de recuperación que también usa el errorHandler
	srv.SetRecoverFunc(func(ctx context.Context, err any) error {
		// Asegurar que se guarde el errorHandler en el contexto para uso posterior
		ctx = context.WithValue(ctx, middleware.ErrorHandlerKey{}, errorHandler)

		appLogger.Error("GraphQL panic recovered",
			zap.Any("error", err),
			zap.String("path", graphql.GetPath(ctx).String()))

		return &gqlerror.Error{
			Path:    graphql.GetPath(ctx),
			Message: "Internal server error",
			Extensions: map[string]any{
				"code": appErrors.ErrInternalError,
			},
		}
	})

	// Registrar métricas para la base de datos
	metrics.RegisterMetrics(db)

	// Crear middlewares individuales
	// Usar nuestro nuevo middleware de autenticación basado en JWT
	authMiddleware := middleware.AuthMiddleware(authService, appLogger)
	validationMiddleware := middleware.NewValidationMiddleware()
	recoveryMiddleware := middleware.NewRecoveryMiddleware(appLogger)
	transactionMiddleware := middleware.NewTransactionMiddleware(db)
	rateLimiter := auth.NewRateLimiter(
		rate.Limit(cfg.RateLimitRPS),
		cfg.RateLimitBurst,
		cfg.RateLimitCleanup,
		appLogger,
	)
	securityHeaders := auth.NewSecurityHeadersMiddleware()
	metricsMiddleware := metrics.NewMetricsMiddleware()

	// Construir la cadena de middleware con manejo de errores coherente
	var handlerChain http.Handler = srv

	// Orden de middleware revisado para manejo de errores coherente:
	handlerChain = transactionMiddleware.Handler(handlerChain) // Transacciones (más interno)
	handlerChain = metricsMiddleware(handlerChain)             // Métricas después de las transacciones

	// El middleware de errores debe ir ANTES de middlewares que pueden generar errores
	// pero después de middlewares que modifican el flujo (como transacciones)
	handlerChain = errorHandler.Handler(handlerChain) // *** Error handling aquí ***

	handlerChain = authMiddleware(handlerChain)               // Autenticación JWT - errores manejados por errorHandler
	handlerChain = validationMiddleware.Handler(handlerChain) // Validación
	handlerChain = rateLimiter.Middleware(handlerChain)       // Rate limiting
	handlerChain = recoveryMiddleware.Handler(handlerChain)   // Recuperación de pánicos - alimenta a errorHandler
	handlerChain = securityHeaders.Middleware(handlerChain)   // Headers de seguridad

	// Crear handler de validación de función
	methodHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Solo aceptar POST (OPTIONS es manejado por CORS)
		if r.Method != http.MethodPost && r.Method != http.MethodOptions {
			ctx := context.WithValue(r.Context(), middleware.ErrorHandlerKey{}, errorHandler)
			err := appErrors.New(appErrors.ErrInvalidOperation, "Method not allowed. GraphQL only accepts POST requests")
			gqlErr := errorHandler.HandleError(ctx, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			writeJSON(w, map[string]any{
				"errors": []*gqlerror.Error{gqlErr},
			})
			return
		}

		// Establecer Content-Type para respuestas válidas
		w.Header().Set("Content-Type", "application/json")

		// Pasar al handler chain
		handlerChain.ServeHTTP(w, r)
	})

	// Aplicar CORS como capa más externa
	return corsMiddleware(methodHandler)
}

// writeJSON es un helper para escribir JSON en la respuesta
func writeJSON(w http.ResponseWriter, data any) {
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}

// NewPlaygroundHandler crea un nuevo handler para el playground GraphQL
func NewPlaygroundHandler() http.Handler {
	playgroundHandler := playground.Handler("ASAM GraphQL Playground", "/graphql")

	// Aplicar CORS al playgroundHandler para consistencia
	return corsMiddleware(playgroundHandler)
}
