package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
)

// AuthMiddleware maneja la autenticación para las solicitudes GraphQL
func AuthMiddleware(authService input.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Obtener el token del header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No hay token, continuar sin autenticación
				next.ServeHTTP(w, r)
				return
			}

			// El formato debe ser "Bearer {token}"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				logger.Warn("Formato de token inválido",
					zap.String("authorization", authHeader),
					zap.String("ip", getClientIP(r)),
				)
				next.ServeHTTP(w, r)
				return
			}

			token := parts[1]

			// Validar el token
			user, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Warn("Token inválido",
					zap.Error(err),
					zap.String("ip", getClientIP(r)),
				)
				next.ServeHTTP(w, r)
				return
			}

			// Añadir información al contexto
			ctx := context.WithValue(r.Context(), constants.UserContextKey, user)
			ctx = context.WithValue(ctx, constants.AuthorizedContextKey, true)
			ctx = context.WithValue(ctx, "authorization", token) // Para la función getAccessTokenFromContext

			// Guardar información para auditoría
			ctx = context.WithValue(ctx, constants.UserIDContextKey, user.ID)
			ctx = context.WithValue(ctx, constants.UserRoleContextKey, user.Role)
			ctx = context.WithValue(ctx, constants.IPContextKey, getClientIP(r))
			ctx = context.WithValue(ctx, constants.UserAgentContextKey, r.UserAgent())

			// Continuar con la solicitud
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getClientIP obtiene la dirección IP real del cliente
func getClientIP(r *http.Request) string {
	// Primero checar X-Forwarded-For
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Tomar la primera IP (la original)
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Si no hay X-Forwarded-For, usar RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
