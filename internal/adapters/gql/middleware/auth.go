package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// Tipo para las claves de contexto personalizadas
type contextKey string

// Clave personalizada para token de autorización
const authTokenKey contextKey = "authorization"

// publicOperations contiene las operaciones GraphQL que no requieren autenticación
var publicOperations = map[string]bool{
	"login":              true,
	"refreshToken":       true,
	"introspection":      true,
	"IntrospectionQuery": true,
}

// isExemptRequest verifica si la petición está exenta de autenticación
func isExemptRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions || r.URL.Path == "/graphql/playground"
}

// validateAuthHeader valida el encabezado de autorización y extrae el token
func validateAuthHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.NewUnauthorizedError()
	}

	// El formato debe ser "Bearer {token}"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
		return "", errors.NewUnauthorizedError()
	}

	return parts[1], nil
}

// enrichContextWithUserInfo enriquece el contexto con información del usuario
func enrichContextWithUserInfo(ctx context.Context, user *models.User, token string, clientIP string, userAgent string) context.Context {
	// Añadir información básica al contexto
	ctx = context.WithValue(ctx, constants.UserContextKey, user)
	ctx = context.WithValue(ctx, constants.AuthorizedContextKey, true)
	ctx = context.WithValue(ctx, authTokenKey, token) // Para la función getAccessTokenFromContext

	// Guardar información para auditoría
	ctx = context.WithValue(ctx, constants.UserIDContextKey, user.ID)
	ctx = context.WithValue(ctx, constants.UserRoleContextKey, user.Role)
	ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
	ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)

	return ctx
}

// handlePublicOperation maneja operaciones públicas que no requieren autenticación
func handlePublicOperation(w http.ResponseWriter, r *http.Request, next http.Handler, logger logger.Logger, operationName string) {
	logger.Debug("Operación pública permitida sin autenticación",
		zap.String("operation", operationName),
		zap.String("ip", getClientIP(r)),
	)
	next.ServeHTTP(w, r)
}

// handleAuthFailure maneja fallos de autenticación
func handleAuthFailure(w http.ResponseWriter, msg string, logger logger.Logger, operation string, clientIP string, err error) {
	if err != nil {
		logger.Warn("Error de autenticación",
			zap.Error(err),
			zap.String("operation", operation),
			zap.String("ip", clientIP),
		)
	} else {
		logger.Warn("Error de autenticación",
			zap.String("operation", operation),
			zap.String("ip", clientIP),
		)
	}
	respondWithAuthError(w, msg)
}

// AuthMiddleware maneja la autenticación para las solicitudes GraphQL
func AuthMiddleware(authService input.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verificar si es una petición exenta de autenticación
			if isExemptRequest(r) {
				next.ServeHTTP(w, r)
				return
			}

			// Verificar si es una operación pública (login, refresh token, etc.)
			isPublicOp, operationName := isPublicOperation(r)
			if isPublicOp {
				handlePublicOperation(w, r, next, logger, operationName)
				return
			}

			// Obtener y validar el token del header Authorization
			token, err := validateAuthHeader(r.Header.Get("Authorization"))
			if err != nil {
				handleAuthFailure(w, err.Error(), logger, operationName, getClientIP(r), nil)
				return
			}

			// Validar el token
			user, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				// Obtener mensaje de error más preciso
				msg := "Token inválido o expirado"
				if appErr, ok := errors.AsAppError(err); ok {
					msg = appErr.Message
				}

				handleAuthFailure(w, msg, logger, operationName, getClientIP(r), err)
				return
			}

			// Enriquecer el contexto con la información del usuario
			ctx := enrichContextWithUserInfo(r.Context(), user, token, getClientIP(r), r.UserAgent())

			// Log de acceso exitoso
			logger.Debug("Acceso autenticado exitoso",
				zap.Uint("user_id", user.ID),
				zap.String("username", user.Username),
				zap.String("role", string(user.Role)),
				zap.String("operation", operationName),
				zap.String("ip", getClientIP(r)),
			)

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

// isPublicOperation determina si la operación GraphQL es pública
func isPublicOperation(r *http.Request) (bool, string) {
	// Solo procesar solicitudes POST para GraphQL
	if r.Method != http.MethodPost {
		return true, "non-graphql"
	}

	// Verificar si es una consulta de introspección (común en herramientas como GraphiQL)
	if strings.Contains(r.URL.Path, "/playground") || strings.Contains(r.URL.Path, "/graphiql") {
		return true, "playground"
	}

	// Intentar parsear el cuerpo de la solicitud para obtener la operación
	var gqlRequest struct {
		OperationName string `json:"operationName"`
		Query         string `json:"query"`
	}

	// Decodificar sin consumir el body
	bodyBytes, err := r.GetBody()
	if err != nil {
		// Si no se puede obtener el body, asumir que necesita autenticación
		return false, "unknown"
	}

	defer func(bodyBytes io.ReadCloser) {
		_ = bodyBytes.Close()
	}(bodyBytes)

	if err := json.NewDecoder(bodyBytes).Decode(&gqlRequest); err != nil {
		// Si no podemos decodificar, asumir que necesita autenticación
		return false, "unparseable"
	}

	// Restaurar el body para el siguiente middleware
	r.Body = bodyBytes

	// Verificar si la operación está en la lista de operaciones públicas
	isPublic := false
	operationName := gqlRequest.OperationName

	// Verificar por nombre de operación
	if publicOperations[operationName] {
		isPublic = true
	}

	// Si no hay nombre de operación o no está en la lista, verificar la query
	if !isPublic && gqlRequest.Query != "" {
		// Check for login mutation
		if strings.Contains(gqlRequest.Query, "mutation login") ||
			strings.Contains(gqlRequest.Query, "mutation { login") {
			isPublic = true
			operationName = "login"
		}

		// Check for refreshToken mutation
		if strings.Contains(gqlRequest.Query, "mutation refreshToken") ||
			strings.Contains(gqlRequest.Query, "refreshToken(") {
			isPublic = true
			operationName = "refreshToken"
		}

		// Check for introspection query
		if strings.Contains(gqlRequest.Query, "__schema") ||
			strings.Contains(gqlRequest.Query, "__type") {
			isPublic = true
			operationName = "introspection"
		}
	}

	return isPublic, operationName
}

// respondWithAuthError envía una respuesta de error de autenticación en formato GraphQL
func respondWithAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	// Crear un error en formato GraphQL
	gqlError := &gqlerror.Error{
		Message: message,
		Extensions: map[string]any{
			"code": errors.ErrUnauthorized,
		},
	}

	// Construir respuesta JSON
	response := map[string]any{
		"errors": []*gqlerror.Error{gqlError},
		"data":   nil,
	}

	// Enviar respuesta
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Si ocurre un error al codificar, enviar una respuesta simple
		if _, writeErr := w.Write([]byte(`{"errors":[{"message":"Unauthorized","extensions":{"code":"UNAUTHORIZED"}}],"data":null}`)); writeErr != nil {
			// No podemos usar logger.Logger{} directamente, usemos un log simple
			log := zap.NewExample().Sugar()
			log.Errorf("Error writing unauthorized response: %v", writeErr)
		}
	}
}
