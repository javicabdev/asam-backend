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

// publicOperations contiene las operaciones GraphQL que no requieren autenticación
var publicOperations = map[string]bool{
	"login":                   true,
	"refreshToken":            true,
	"introspection":           true,
	"IntrospectionQuery":      true,
	"health":                  true,
	"ping":                    true,
	"verifyEmail":             true,
	"resendVerificationEmail": true,
	"requestPasswordReset":    true,
	"resetPasswordWithToken":  true,
}

// Regex patterns for detecting public operations (case-insensitive)
var (
	authOperationsRegex             = regexp.MustCompile(`(?i)mutation\s+(login|refreshtoken)\s*[({]`)
	anonymousAuthRegex              = regexp.MustCompile(`(?i)mutation\s*\{\s*(login|refreshtoken)\s*\(`)
	emailVerificationRegex          = regexp.MustCompile(`(?i)mutation\s+(verifyemail|resendverificationemail)\s*[({]`)
	anonymousEmailVerificationRegex = regexp.MustCompile(`(?i)mutation\s*\{\s*(verifyemail|resendverificationemail)\s*\(`)
	passwordResetRegex              = regexp.MustCompile(`(?i)mutation\s+(requestpasswordreset|resetpasswordwithtoken)\s*[({]`)
	anonymousPasswordResetRegex     = regexp.MustCompile(`(?i)mutation\s*\{\s*(requestpasswordreset|resetpasswordwithtoken)\s*\(`)
	introspectionRegex              = regexp.MustCompile(`(?i)(query\s+introspectionquery|__schema|__type)`)
	healthCheckRegex                = regexp.MustCompile(`(?i)query\s+(health|ping)\s*[({]|\{\s*(health|ping)\s*}`)
	operationNameRegex              = regexp.MustCompile(`(?i)^\s*(query|mutation|subscription)\s+(\w+)`)
)

// AuthMiddleware es el punto de entrada del middleware de autenticación.
// Su única responsabilidad es delegar el manejo de la petición a una función de lógica dedicada.
// Esto mantiene el código limpio y reduce la complejidad ciclomática.
func AuthMiddleware(authService input.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleAuthRequest(w, r, next, authService, logger)
		})
	}
}

// handleAuthRequest es el orquestador principal de la lógica de autenticación.
// Decide si una petición debe pasar directamente, si es pública, o si necesita validación de token.
func handleAuthRequest(w http.ResponseWriter, r *http.Request, next http.Handler, authService input.AuthService, logger logger.Logger) {
	// 1. Peticiones exentas (OPTIONS, playground) no necesitan más procesamiento.
	if isExemptRequest(r) {
		next.ServeHTTP(w, r)
		return
	}

	// 2. Determinar si la operación es pública (login, etc.).
	isPublic, operationName := isPublicOperation(r)
	if isPublic {
		logger.Debug("Public operation detected, skipping token validation", zap.String("operation", operationName))
		next.ServeHTTP(w, r)
		return
	}

	// 3. Si no es exenta ni pública, es una operación privada que requiere autenticación.
	handlePrivateOperation(w, r, next, authService, logger, operationName)
}

// handlePrivateOperation gestiona la lógica para operaciones que requieren un token válido.
func handlePrivateOperation(w http.ResponseWriter, r *http.Request, next http.Handler, authService input.AuthService, logger logger.Logger, operationName string) {
	// a. Validar el encabezado de autorización y extraer el token.
	authHeader := r.Header.Get("authorization")
	token, err := validateAuthHeader(authHeader)
	if err != nil {
		handleAuthFailure(w, err.Error(), logger, operationName, getClientIP(r), nil)
		return
	}

	// b. Validar el token con el servicio de autenticación.
	user, err := authService.ValidateToken(r.Context(), token)
	if err != nil {
		msg := "Token inválido o expirado"
		if appErr, ok := errors.AsAppError(err); ok {
			msg = appErr.Message
		}
		handleAuthFailure(w, msg, logger, operationName, getClientIP(r), err)
		return
	}

	// c. Éxito: Enriquecer el contexto con la información del usuario y continuar.
	ctx := enrichContextWithUserInfo(r.Context(), user, token, getClientIP(r), r.UserAgent())

	logger.Debug("Authenticated access successful",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("role", string(user.Role)),
		zap.String("operation", operationName),
		zap.String("ip", getClientIP(r)),
	)

	next.ServeHTTP(w, r.WithContext(ctx))
}

// --- FUNCIONES AUXILIARES (Sin cambios respecto al original) ---

// isExemptRequest verifica si la petición está exenta de autenticación por su naturaleza.
func isExemptRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions ||
		r.URL.Path == "/playground" ||
		r.URL.Path == "/graphql/playground" ||
		strings.HasSuffix(r.URL.Path, "/playground")
}

// validateAuthHeader valida el encabezado de autorización y extrae el token.
func validateAuthHeader(authHeader string) (token string, err error) {
	if authHeader == "" {
		return "", errors.NewUnauthorizedError()
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", errors.NewUnauthorizedError()
	}

	return parts[1], nil
}

// enrichContextWithUserInfo enriquece el contexto con información del usuario.
func enrichContextWithUserInfo(ctx context.Context, user *models.User, token string, clientIP string, userAgent string) context.Context {
	ctx = context.WithValue(ctx, constants.UserContextKey, user)
	ctx = context.WithValue(ctx, constants.AuthorizedContextKey, true)
	ctx = context.WithValue(ctx, constants.AuthTokenContextKey, token)
	ctx = context.WithValue(ctx, constants.UserIDContextKey, user.ID)
	ctx = context.WithValue(ctx, constants.UserRoleContextKey, user.Role)
	if ip := ctx.Value(constants.IPContextKey); ip == nil && clientIP != "" {
		ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
	}
	if ua := ctx.Value(constants.UserAgentContextKey); ua == nil && userAgent != "" {
		ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)
	}
	return ctx
}

// handleAuthFailure maneja fallos de autenticación registrando el error y respondiendo al cliente.
func handleAuthFailure(w http.ResponseWriter, msg string, logger logger.Logger, operation string, clientIP string, err error) {
	logFields := []zap.Field{
		zap.String("operation", operation),
		zap.String("ip", clientIP),
	}
	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}
	logger.Warn("Authentication error", logFields...)
	respondWithAuthError(w, msg)
}

// getClientIP obtiene la dirección IP real del cliente.
func getClientIP(r *http.Request) string {
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

type gqlRequest struct {
	OperationName string `json:"operationName"`
	Query         string `json:"query"`
}

// isPublicOperation determina si la operación GraphQL es pública leyendo el cuerpo de la petición.
func isPublicOperation(r *http.Request) (isPublic bool, operationName string) {
	if r.Method != http.MethodPost {
		return true, "non-graphql-request"
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return false, "read-error"
	}
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // Restore body

	if len(bodyBytes) == 0 {
		return false, "empty-body"
	}

	var req gqlRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		// If unmarshal fails, it might be a query without variables. Check with regex.
		isPublic, opName := checkByQueryContent(string(bodyBytes))
		return isPublic, opName
	}

	if isPublic, opName := checkByOperationName(req.OperationName); isPublic {
		return true, opName
	}

	if req.Query != "" {
		return checkByQueryContent(req.Query)
	}

	return false, req.OperationName
}

// checkByOperationName verifica si el nombre de la operación es público.
func checkByOperationName(name string) (isPublic bool, operationName string) {
	if name == "" {
		return false, ""
	}
	if publicOperations[strings.ToLower(name)] {
		return true, name
	}
	return false, name
}

// checkByQueryContent verifica el contenido de la query usando regex.
func checkByQueryContent(query string) (isPublic bool, operationName string) {
	patterns := []struct {
		regex       *regexp.Regexp
		defaultName string
	}{
		{authOperationsRegex, "login"},
		{anonymousAuthRegex, "login"},
		{emailVerificationRegex, "emailVerification"},
		{anonymousEmailVerificationRegex, "emailVerification"},
		{passwordResetRegex, "passwordReset"},
		{anonymousPasswordResetRegex, "passwordReset"},
		{introspectionRegex, "introspection"},
		{healthCheckRegex, "health"},
	}

	for _, p := range patterns {
		if p.regex.MatchString(query) {
			matches := p.regex.FindStringSubmatch(query)
			opName := p.defaultName
			if len(matches) > 1 {
				opName = strings.ToLower(matches[1])
			}
			return true, opName
		}
	}

	matches := operationNameRegex.FindStringSubmatch(query)
	if len(matches) > 2 {
		extractedName := strings.ToLower(matches[2])
		if publicOperations[extractedName] {
			return true, extractedName
		}
		return false, extractedName
	}

	return false, ""
}

// respondWithAuthError envía una respuesta de error de autenticación en formato GraphQL.
func respondWithAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]any{
		"errors": []gqlerror.Error{
			{
				Message: message,
				Extensions: map[string]any{
					"code": errors.ErrUnauthorized,
				},
			},
		},
		"data": nil,
	}

	_ = json.NewEncoder(w).Encode(response)
}
