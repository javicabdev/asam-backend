package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/constants"
)

// Middleware provides authentication middleware for HTTP requests.
type Middleware struct {
	authService input.AuthService
}

// Handler returns an HTTP middleware function that authenticates requests
// by validating JWT tokens from the authorization header and adding user info to the request context.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Si es una operación de introspección de GraphQL, permitir sin auth
		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Obtener token del header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Para operaciones que no requieren autenticación (login, register),
			// continuar con el contexto actualizado
			next.ServeHTTP(w, r)
			return
		}

		// Validar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendError(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Validar token
		user, err := m.authService.ValidateToken(r.Context(), parts[1])
		if err != nil {
			sendError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Añadir usuario al contexto
		ctx := context.WithValue(r.Context(), constants.UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type errorResponse struct {
	Error string `json:"error"`
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(errorResponse{Error: message})
	if err != nil {
		return
	}
}
