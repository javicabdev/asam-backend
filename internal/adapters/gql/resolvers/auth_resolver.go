package resolvers

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

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

	// Llamada al servicio de autenticación
	tokenDetails, err := r.authService.Login(ctx, username, password)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "credenciales inválidas")
	}

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
// Esta función simplemente devuelve el mismo user que recibe,
// ya que estamos usando directamente el modelo del dominio en GraphQL
func mapUserToGQL(user *models.User) *models.User {
	return user
}

// getAccessTokenFromContext obtiene el token de acceso del contexto
func getAccessTokenFromContext(ctx context.Context) (string, error) {
	// Primero intentar obtener del contexto con la clave que usa el middleware
	token, ok := ctx.Value("authorization").(string)
	if !ok || token == "" {
		// Intentar buscar en los headers
		if authHeader, ok := ctx.Value("Authorization").(string); ok && authHeader != "" {
			token = authHeader
		} else {
			return "", errors.NewBusinessError(errors.ErrUnauthorized, "token no encontrado en el contexto")
		}
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
