// Package gql provides GraphQL server implementation for the ASAM backend.
// It includes handlers, middleware configuration, and integration with the resolver layer.
package gql

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/generated"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/metrics"
)

// corsMiddleware is a more comprehensive CORS middleware that handles all necessary headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configure CORS headers
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
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

	// Habilitar introspección solo en desarrollo
	if cfg.Environment == "development" {
		srv.Use(extension.Introspection{})
		appLogger.Info("GraphQL introspection enabled for development")
	}

	// Configure transport options to ensure all transports are properly enabled
	srv.AddTransport(&transport.Options{})
	srv.AddTransport(&transport.GET{})
	srv.AddTransport(&transport.POST{})
	srv.AddTransport(&transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
	})

	// Add a middleware to preserve HTTP context values
	// This is critical for authentication to work properly
	srv.Use(extension.FixedComplexityLimit(cfg.GQLComplexityLimit))

	// Add field middleware to log context at field level
	srv.AroundFields(middleware.FieldContextMiddleware(appLogger))

	// Add middleware to clean __typename from inputs (for Apollo Client compatibility)
	srv.AroundFields(middleware.TypenameCleanerMiddleware(appLogger))

	// Use AroundOperations to ensure context is properly propagated
	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		// Extract operation name for logging
		opCtx := graphql.GetOperationContext(ctx)
		opName := "unknown"
		if opCtx != nil {
			opName = opCtx.OperationName
		}

		// Extract user from context if it exists
		user, userOk := ctx.Value(constants.UserContextKey).(*models.User)
		token, _ := ctx.Value(constants.AuthTokenContextKey).(string)
		authorized, _ := ctx.Value(constants.AuthorizedContextKey).(bool)

		// Enhanced logging for sendVerificationEmail
		if opName == "SendVerificationEmail" || opName == "sendVerificationEmail" {
			appLogger.Info("[GRAPHQL-DEBUG] SendVerificationEmail operation context",
				zap.String("operation", opName),
				zap.Bool("hasUser", userOk && user != nil),
				zap.Bool("hasToken", token != ""),
				zap.Bool("authorized", authorized),
				zap.Bool("userContextKeyExists", ctx.Value(constants.UserContextKey) != nil),
				zap.String("userContextType", fmt.Sprintf("%T", ctx.Value(constants.UserContextKey))),
			)

			if user != nil {
				appLogger.Info("[GRAPHQL-DEBUG] User details in context",
					zap.Uint("userID", user.ID),
					zap.String("username", user.Username),
					zap.String("email", user.Email),
					zap.Bool("emailVerified", user.EmailVerified),
					zap.String("role", string(user.Role)),
				)
			} else {
				appLogger.Error("[GRAPHQL-DEBUG] User is nil in context for SendVerificationEmail")
			}
		} else {
			// Log normal operations
			appLogger.Debug("GraphQL AroundOperations: Context check",
				zap.String("operation", opName),
				zap.Bool("hasUser", userOk && user != nil),
				zap.Bool("hasToken", token != ""),
				zap.Bool("authorized", authorized),
			)
		}

		// Continue with the operation
		return next(ctx)
	})

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
	// IMPORTANTE: El orden es crítico para la propagación del contexto

	// Primero aplicamos el middleware de autenticación para que el contexto tenga el usuario
	handlerChain = authMiddleware(handlerChain) // Autenticación JWT - establece el usuario en el contexto

	// Luego aplicamos los demás middlewares
	handlerChain = transactionMiddleware.Handler(handlerChain) // Transacciones
	handlerChain = metricsMiddleware(handlerChain)             // Métricas
	handlerChain = errorHandler.Handler(handlerChain)          // Error handling
	handlerChain = validationMiddleware.Handler(handlerChain)  // Validación
	handlerChain = rateLimiter.Middleware(handlerChain)        // Rate limiting
	handlerChain = recoveryMiddleware.Handler(handlerChain)    // Recuperación de pánicos
	handlerChain = securityHeaders.Middleware(handlerChain)    // Headers de seguridad

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
