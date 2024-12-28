package gql

import (
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/generated"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
)

// NewHandler crea un nuevo handler de GraphQL
func NewHandler() http.Handler {
	schema := generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolvers.Resolver{},
	})

	srv := handler.New(schema)

	// Configurar opciones básicas del servidor
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.MultipartForm{})

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

		srv.ServeHTTP(w, r)
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
