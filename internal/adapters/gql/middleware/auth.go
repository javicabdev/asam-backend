// Package middleware provides HTTP and GraphQL middleware components for the ASAM backend.
// It includes authentication, authorization, error handling, and request processing middleware.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// publicOperations contiene las operaciones GraphQL que no requieren autenticación.
// CORRECCIÓN: Todas las claves están ahora en minúsculas para que coincidan con la comprobación `strings.ToLower`.
var publicOperations = map[string]bool{
	"login":                   true,
	"refreshtoken":            true,
	"introspection":           true,
	"introspectionquery":      true,
	"health":                  true,
	"ping":                    true,
	"verifyemail":             true,
	"resendverificationemail": true,
	"requestpasswordreset":    true,
	"resetpasswordwithtoken":  true,
}

// gqlRequestBody is used to decode the operation name and query from the request body.
type gqlRequestBody struct {
	OperationName string `json:"operationName"`
	Query         string `json:"query"`
}

// AuthMiddleware es el punto de entrada del middleware de autenticación.
func AuthMiddleware(authService input.AuthService, logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleAuthRequest(w, r, next, authService, logger)
		})
	}
}

// handleAuthRequest es el orquestador principal de la lógica de autenticación.
func handleAuthRequest(w http.ResponseWriter, r *http.Request, next http.Handler, authService input.AuthService, logger logger.Logger) {
	if isExemptRequest(r) {
		next.ServeHTTP(w, r)
		return
	}

	operationName, err := getOperationName(r)
	if err != nil {
		logger.Error("[AUTH] Could not parse GraphQL request", zap.Error(err))
		respondWithBadRequest(w, "Invalid GraphQL request body")
		return
	}

	if operationName == "" {
		logger.Warn("[AUTH] Could not determine operation name. Rejecting as Bad Request.",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("ip", getClientIP(r)),
		)
		respondWithBadRequest(w, "Bad Request: Could not determine GraphQL operation.")
		return
	}

	if publicOperations[strings.ToLower(operationName)] {
		next.ServeHTTP(w, r)
		return
	}

	handlePrivateOperation(w, r, next, authService, logger, operationName)
}

// handlePrivateOperation gestiona la lógica para operaciones que requieren un token válido.
func handlePrivateOperation(w http.ResponseWriter, r *http.Request, next http.Handler, authService input.AuthService, logger logger.Logger, operationName string) {
	authHeader := r.Header.Get("authorization")
	token, err := validateAuthHeader(authHeader)
	if err != nil {
		handleAuthFailure(w, err.Error(), logger, operationName, getClientIP(r), nil)
		return
	}

	user, err := authService.ValidateToken(r.Context(), token)
	if err != nil {
		msg := "Token inválido o expirado"
		if appErr, ok := appErrors.AsAppError(err); ok {
			msg = appErr.Message
		}
		handleAuthFailure(w, msg, logger, operationName, getClientIP(r), err)
		return
	}

	if user == nil {
		handleAuthFailure(w, "Token válido pero no se encontró el usuario", logger, operationName, getClientIP(r), nil)
		return
	}

	ctx := enrichContextWithUserInfo(r.Context(), user, token, getClientIP(r), r.UserAgent())

	next.ServeHTTP(w, r.WithContext(ctx))
}

// --- Package-Shared Helper Functions ---

var operationNameRegex = regexp.MustCompile(`(?i)(?:mutation|query)\s+(\w+)`)

func getOperationName(r *http.Request) (string, error) {
	if r.Method != http.MethodPost || r.Body == nil {
		return "", nil
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("error reading request body: %w", err)
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if len(bodyBytes) == 0 {
		return "", nil
	}

	var reqBody gqlRequestBody
	if err := json.Unmarshal(bodyBytes, &reqBody); err != nil {
		return "", fmt.Errorf("error unmarshaling gql request body: %w", err)
	}

	if reqBody.OperationName != "" {
		return reqBody.OperationName, nil
	}

	if reqBody.Query != "" {
		matches := operationNameRegex.FindStringSubmatch(reqBody.Query)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", nil
}

func isExemptRequest(r *http.Request) bool {
	return r.Method == http.MethodOptions ||
		strings.Contains(r.URL.Path, "playground")
}

func validateAuthHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", appErrors.NewUnauthorizedError()
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", appErrors.NewUnauthorizedError()
	}
	return parts[1], nil
}

func enrichContextWithUserInfo(ctx context.Context, user *models.User, token, clientIP, userAgent string) context.Context {
	ctx = context.WithValue(ctx, constants.UserContextKey, user)
	ctx = context.WithValue(ctx, constants.AuthorizedContextKey, true)
	ctx = context.WithValue(ctx, constants.AuthTokenContextKey, token)
	ctx = context.WithValue(ctx, constants.UserIDContextKey, user.ID)
	ctx = context.WithValue(ctx, constants.UserRoleContextKey, user.Role)
	ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
	ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)
	return ctx
}

func handleAuthFailure(w http.ResponseWriter, msg string, logger logger.Logger, operation, clientIP string, err error) {
	logFields := []zap.Field{zap.String("operation", operation), zap.String("ip", clientIP)}
	if err != nil {
		logFields = append(logFields, zap.Error(err))
	}
	logger.Warn("Authentication error", logFields...)
	respondWithAuthError(w, msg)
}

func getClientIP(r *http.Request) string {
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		return strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return ip
}

func respondWithAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	response := map[string]any{
		"errors": []gqlerror.Error{
			{
				Message: message,
				Extensions: map[string]any{
					"code": appErrors.ErrUnauthorized,
				},
			},
		},
		"data": nil,
	}
	_ = json.NewEncoder(w).Encode(response)
}

func respondWithBadRequest(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	response := map[string]any{
		"errors": []gqlerror.Error{
			{
				Message: message,
				Extensions: map[string]any{
					"code": "BAD_REQUEST",
				},
			},
		},
		"data": nil,
	}
	_ = json.NewEncoder(w).Encode(response)
}
