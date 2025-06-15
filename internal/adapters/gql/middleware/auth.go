// Package middleware provides HTTP and GraphQL middleware components for the ASAM backend.
// It includes authentication, authorization, error handling, and request processing middleware.
package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
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
	"health":             true,
	"ping":               true,
}

// Regex patterns for detecting public operations (case-insensitive)
var (
	// Pattern for login/refreshToken mutations
	authOperationsRegex = regexp.MustCompile(`(?i)mutation\s+(login|refreshtoken)\s*[({]`)
	// Pattern for anonymous mutations with login/refreshToken
	anonymousAuthRegex = regexp.MustCompile(`(?i)mutation\s*\{\s*(login|refreshtoken)\s*\(`)
	// Pattern for introspection queries
	introspectionRegex = regexp.MustCompile(`(?i)(query\s+introspectionquery|__schema|__type)`)
	// Pattern for health check queries
	healthCheckRegex = regexp.MustCompile(`(?i)query\s+(health|ping)\s*[({]|\{\s*(health|ping)\s*}`)
	// Pattern for extracting operation name
	operationNameRegex = regexp.MustCompile(`(?i)^\s*(query|mutation|subscription)\s+(\w+)`)
)

// isExemptRequest verifica si la petición está exenta de autenticación
func isExemptRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions ||
		r.URL.Path == "/playground" ||
		r.URL.Path == "/graphql/playground" ||
		strings.HasSuffix(r.URL.Path, "/playground")
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

	// Preservar información del cliente si ya existe en el contexto
	// (desde el middleware HTTP)
	if ip := ctx.Value(constants.IPContextKey); ip == nil && clientIP != "" {
		ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
	}
	if ua := ctx.Value(constants.UserAgentContextKey); ua == nil && userAgent != "" {
		ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)
	}
	// Preservar device_name si existe
	if deviceName := ctx.Value("device_name"); deviceName != nil {
		ctx = context.WithValue(ctx, "device_name", deviceName)
	}

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
				// Pasar el contexto con la información del cliente a las operaciones públicas
				next.ServeHTTP(w, r)
				return
			}

			// Obtener y validar el token del header authorization
			token, err := validateAuthHeader(r.Header.Get("authorization"))
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

			// Obtener información del cliente del contexto (si fue establecida por middleware HTTP)
			clientIP := getClientIP(r)
			if ipFromCtx := r.Context().Value(constants.IPContextKey); ipFromCtx != nil {
				if ip, ok := ipFromCtx.(string); ok && ip != "" {
					clientIP = ip
				}
			}

			userAgent := r.UserAgent()
			if uaFromCtx := r.Context().Value(constants.UserAgentContextKey); uaFromCtx != nil {
				if ua, ok := uaFromCtx.(string); ok && ua != "" {
					userAgent = ua
				}
			}

			// Enriquecer el contexto con la información del usuario
			ctx := enrichContextWithUserInfo(r.Context(), user, token, clientIP, userAgent)

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
// Mejorado con regex patterns case-insensitive para mayor flexibilidad
func isPublicOperation(r *http.Request) (bool, string) {
	// Solo procesar solicitudes POST para GraphQL
	if r.Method != http.MethodPost {
		return true, "non-graphql"
	}

	// Verificar si es una consulta de introspección (común en herramientas como GraphiQL)
	if strings.Contains(r.URL.Path, "/playground") || strings.Contains(r.URL.Path, "/graphiql") {
		return true, "playground"
	}

	// Verificar que el body no sea nil
	if r.Body == nil {
		return false, "empty-body"
	}

	// Intentar parsear el cuerpo de la solicitud para obtener la operación
	var gqlRequest struct {
		OperationName string `json:"operationName"`
		Query         string `json:"query"`
	}

	// Leer el body actual
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// Si no se puede leer el body, asumir que necesita autenticación
		return false, "read-error"
	}

	// Restaurar el body para que pueda ser leído nuevamente
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	// Si el body está vacío, retornar false
	if len(bodyBytes) == 0 {
		return false, "empty-body"
	}

	if err := json.Unmarshal(bodyBytes, &gqlRequest); err != nil {
		// Si no podemos decodificar, asumir que necesita autenticación
		return false, "unparseable"
	}

	// Verificar si la operación está en la lista de operaciones públicas
	isPublic := false
	operationName := gqlRequest.OperationName

	// 1. Verificar por nombre de operación (case-insensitive)
	if operationName != "" {
		operationNameLower := strings.ToLower(operationName)
		for publicOp := range publicOperations {
			if strings.ToLower(publicOp) == operationNameLower {
				isPublic = true
				operationName = publicOp
				break
			}
		}
	}

	// 2. Si no es pública por nombre, verificar el contenido de la query usando regex
	if !isPublic && gqlRequest.Query != "" {
		// Check for auth operations (login/refreshToken)
		if authOperationsRegex.MatchString(gqlRequest.Query) {
			isPublic = true
			// Extract the operation name from the regex match
			matches := authOperationsRegex.FindStringSubmatch(gqlRequest.Query)
			if len(matches) > 1 {
				operationName = strings.ToLower(matches[1])
			}
		}

		// Check for anonymous mutations with login/refreshToken
		if !isPublic && anonymousAuthRegex.MatchString(gqlRequest.Query) {
			isPublic = true
			// Extract the operation name from the regex match
			matches := anonymousAuthRegex.FindStringSubmatch(gqlRequest.Query)
			if len(matches) > 1 {
				operationName = strings.ToLower(matches[1])
			}
		}

		// Check for introspection queries
		if !isPublic && introspectionRegex.MatchString(gqlRequest.Query) {
			isPublic = true
			operationName = "introspection"
		}

		// Check for health check queries
		if !isPublic && healthCheckRegex.MatchString(gqlRequest.Query) {
			isPublic = true
			// Extract the operation name from the regex match
			matches := healthCheckRegex.FindStringSubmatch(gqlRequest.Query)
			if len(matches) > 1 {
				operationName = strings.ToLower(matches[1])
			}
		}

		// 3. Si no se encontró por regex, intentar extraer el nombre de la operación
		if operationName == "" {
			matches := operationNameRegex.FindStringSubmatch(gqlRequest.Query)
			if len(matches) > 2 {
				extractedName := strings.ToLower(matches[2])
				if publicOperations[extractedName] {
					isPublic = true
					operationName = extractedName
				}
			}
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
