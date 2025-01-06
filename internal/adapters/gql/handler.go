package gql

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/generated"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/middleware"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
)

// En internal/adapters/gql/handler.go

func NewHandler(authService input.AuthService, resolver *resolvers.Resolver, cfg *config.Config, logger logger.Logger, db *gorm.DB) http.Handler {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	})

	srv := handler.New(schema)
	srv.SetErrorPresenter(CustomErrorPresenter)

	// Configurar opciones básicas del servidor
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Crear middlewares
	authMiddleware := auth.NewAuthMiddleware(authService)
	errorMiddleware := middleware.NewErrorMiddleware(logger)
	validationMiddleware := middleware.NewValidationMiddleware()
	recoveryMiddleware := middleware.NewRecoveryMiddleware(logger)
	transactionMiddleware := middleware.NewTransactionMiddleware(db)
	rateLimiter := auth.NewRateLimiter(
		rate.Limit(cfg.RateLimitRPS),
		cfg.RateLimitBurst,
		cfg.RateLimitCleanup,
	)
	securityHeaders := auth.NewSecurityHeadersMiddleware()

	// Aquí iría el nuevo middleware de métricas
	metricsMiddleware := metrics.NewMetricsMiddleware()

	// Configurar métricas de base de datos
	metrics.RegisterMetrics(db)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers necesarios
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")

		// Manejar preflighted requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Solo permitir POST para queries/mutations
		if r.Method != http.MethodPost {
			// Excepción para el endpoint de métricas
			if r.URL.Path == "/metrics" {
				promhttp.Handler().ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Aplicar middlewares en cadena:
		// 1. Security Headers
		// 2. Recovery (para capturar panics)
		// 3. Rate Limiter
		// 4. Validation
		// 5. Transaction
		// 6. Error Handling
		// 7. Auth
		// 8. Metrics (nuevo)
		securityHeaders.Middleware(
			recoveryMiddleware.Handler(
				rateLimiter.Middleware(
					validationMiddleware.Handler(
						transactionMiddleware.Handler(
							errorMiddleware.Handler(
								authMiddleware.Handler(
									metricsMiddleware(srv),
								),
							),
						),
					),
				),
			),
		).ServeHTTP(w, r)
	})
}

// NewPlaygroundHandler crea un nuevo handler para el playground de GraphQL
func NewPlaygroundHandler() http.Handler {
	h := playground.Handler("ASAM GraphQL Playground", "/graphql")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configurar headers CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
