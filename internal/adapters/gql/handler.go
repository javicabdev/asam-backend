package gql

import (
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"golang.org/x/time/rate"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/generated"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/config"
)

// NewHandler crea un nuevo handler de GraphQL
func NewHandler(authService input.AuthService, resolver *resolvers.Resolver, cfg *config.Config) http.Handler {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: resolver,
	})

	srv := handler.New(schema)

	// Configurar opciones básicas del servidor
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

	// Crear middleware de autenticación
	authMiddleware := auth.NewAuthMiddleware(authService)

	// Crear rate limiter con configuración
	rateLimiter := auth.NewRateLimiter(
		rate.Limit(cfg.RateLimitRPS),
		cfg.RateLimitBurst,
		cfg.RateLimitCleanup,
	)

	securityHeaders := auth.NewSecurityHeadersMiddleware()

	// Middleware para manejar CORS y headers
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
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Aplicar middlewares en cadena:
		// 1. Security Headers (más externo)
		// 2. Rate Limiter
		// 3. Auth (más interno)
		securityHeaders.Middleware(
			rateLimiter.Middleware(
				authMiddleware.Handler(srv),
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
