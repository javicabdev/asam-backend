package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Define un tipo personalizado para claves del contexto
type contextKey string

// Definir la clave para el token de autorización en el contexto
const authorizationKey contextKey = "authorization"

// Login Mutation.login implementa la mutación de login
func (r *Resolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	// Extraer username y password del input
	username := input.Username
	password := input.Password

	// Validación básica de entrada
	if username == "" || password == "" {
		return nil, errors.NewValidationError(
			"VALIDATION_FAILED: Usuario y contraseña son requeridos",
			map[string]string{
				"username": "El nombre de usuario es requerido",
				"password": "La contraseña es requerida",
			},
		)
	}

	// Check rate limiting
	allowed, lockoutDuration := r.loginRateLimiter.AllowLogin(ctx, username)
	if !allowed {
		if lockoutDuration > 0 {
			return nil, errors.NewBusinessError(
				errors.ErrUnauthorized,
				fmt.Sprintf("Demasiados intentos de inicio de sesión. Cuenta bloqueada por %v", lockoutDuration.Round(time.Minute)),
			)
		}
		return nil, errors.NewBusinessError(
			errors.ErrUnauthorized,
			"Demasiados intentos de inicio de sesión. Por favor, intente más tarde",
		)
	}

	// Llamada al servicio de autenticación
	tokenDetails, err := r.authService.Login(ctx, username, password)
	if err != nil {
		// Record failure for rate limiting
		r.loginRateLimiter.RecordFailure(ctx, username)
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "credenciales inválidas")
	}

	// Record success and reset rate limit counter
	r.loginRateLimiter.RecordSuccess(ctx, username)

	// Validar el token para obtener información del usuario
	userModel, err := r.authService.ValidateToken(ctx, tokenDetails.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error validando token")
	}

	// Mapear usuario de dominio a GraphQL
	user := mapUserToGQL(userModel)

	// Construir la respuesta tipada
	return &model.AuthResponse{
		User:         user,
		AccessToken:  tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
	}, nil
}

// Logout Mutation.logout implementa la mutación de logout
func (r *Resolver) Logout(ctx context.Context) (any, error) {
	// Obtener token del contexto
	token, err := getAccessTokenFromContext(ctx)
	if err != nil {
		errMsg := "No se pudo obtener el token de acceso: sesión no iniciada"
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	// Llamada al servicio de autenticación
	err = r.authService.Logout(ctx, token)
	if err != nil {
		errMsg := "Error al cerrar sesión: " + err.Error()
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	successMsg := "Sesión cerrada correctamente"
	return &model.MutationResponse{
		Success: true,
		Message: &successMsg,
	}, nil
}

// RefreshToken Mutation.refreshToken implementa la mutación de refreshToken
func (r *Resolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (any, error) {
	// Extraer refreshToken del input
	refreshToken := input.RefreshToken

	// Validación básica
	if refreshToken == "" {
		return nil, errors.NewValidationError(
			"VALIDATION_FAILED: Refresh token requerido",
			map[string]string{"refreshToken": "El token de refresco es requerido"},
		)
	}

	// Llamada al servicio de autenticación
	tokenDetails, err := r.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "token de refresco inválido o expirado")
	}

	// Construir la respuesta tipada
	return &model.TokenResponse{
		AccessToken:  tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
	}, nil
}

// Funciones auxiliares

// mapUserToGQL convierte un modelo de dominio User a un modelo generado por gqlgen
// Ya no es necesaria la conversión de roles porque el schema GraphQL ahora usa minúsculas
func mapUserToGQL(user *models.User) *models.User {
	// Simplemente retornar el usuario tal cual, ya que los roles coinciden
	return user
}

// getAccessTokenFromContext obtiene el token de acceso del contexto
func getAccessTokenFromContext(ctx context.Context) (string, error) {
	// Intentar obtener el token de las posibles ubicaciones en el contexto
	var token string

	if tokenVal, ok := ctx.Value(authorizationKey).(string); ok && tokenVal != "" {
		token = tokenVal
	}

	// Si no se encontró el token, devolver error
	if token == "" {
		return "", errors.NewBusinessError(errors.ErrUnauthorized, "token no encontrado en el contexto")
	}

	// Quitar el prefijo "Bearer " si existe
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validar que no sea vacío después de limpiar
	if token == "" {
		return "", errors.NewBusinessError(errors.ErrUnauthorized, "token vacío")
	}

	return token, nil
}
