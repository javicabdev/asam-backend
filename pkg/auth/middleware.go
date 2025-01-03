package auth

import (
	"context"
	"encoding/json"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"net/http"
	"strings"
)

type Middleware struct {
	authService input.AuthService
}

func NewAuthMiddleware(authService input.AuthService) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

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
			sendError(w, "Authorization header required", http.StatusUnauthorized)
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
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type contextKey string

const UserContextKey contextKey = "user"

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
