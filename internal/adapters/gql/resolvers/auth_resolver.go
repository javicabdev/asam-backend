package resolvers

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"time"
)

// Helper function to convert string to pointer
func strPtr(s string) *string {
	return &s
}

// Mutation.login implementa la mutación de login
func (r *Resolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	// Llamada al servicio de autenticación
	tokenDetails, err := r.authService.Login(ctx, input.Username, input.Password)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "credenciales inválidas")
	}

	// Validar el token para obtener información del usuario
	user, err := r.authService.ValidateToken(ctx, tokenDetails.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error validando token")
	}

	// Construir la respuesta
	return &model.AuthResponse{
		User:         mapUserToGQL(user),
		AccessToken:  model.JWT(tokenDetails.AccessToken),
		RefreshToken: model.JWT(tokenDetails.RefreshToken),
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
	}, nil
}

// Mutation.logout implementa la mutación de logout
func (r *Resolver) Logout(ctx context.Context) (*model.MutationResponse, error) {
	// Obtener token del contexto
	token, err := getAccessTokenFromContext(ctx)
	if err != nil {
		return &model.MutationResponse{
			Success: false,
			Error:   strPtr("No se pudo obtener el token de acceso"),
		}, nil
	}

	// Llamada al servicio de autenticación
	err = r.authService.Logout(ctx, token)
	if err != nil {
		return &model.MutationResponse{
			Success: false,
			Error:   strPtr("Error al cerrar sesión: " + err.Error()),
		}, nil
	}

	return &model.MutationResponse{
		Success: true,
		Message: strPtr("Sesión cerrada correctamente"),
	}, nil
}

// Mutation.refreshToken implementa la mutación de refreshToken
func (r *Resolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (*model.TokenResponse, error) {
	// Llamada al servicio de autenticación
	tokenDetails, err := r.authService.RefreshToken(ctx, string(input.RefreshToken))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "token de refresco inválido")
	}

	// Construir la respuesta
	return &model.TokenResponse{
		AccessToken:  model.JWT(tokenDetails.AccessToken),
		RefreshToken: model.JWT(tokenDetails.RefreshToken),
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
	}, nil
}

// Funciones auxiliares

// mapUserToGQL convierte un modelo de dominio User a un modelo GQL User
func mapUserToGQL(user *models.User) *model.User {
	if user == nil {
		return nil
	}

	role := model.UserRoleUser
	if user.Role == models.RoleAdmin {
		role = model.UserRoleAdmin
	}

	return &model.User{
		ID:        int(user.ID),
		Username:  user.Username,
		Role:      role,
		IsActive:  user.IsActive,
		LastLogin: &user.LastLogin,
	}
}

// getAccessTokenFromContext obtiene el token de acceso del contexto
func getAccessTokenFromContext(ctx context.Context) (string, error) {
	// En una implementación real, este token vendría del middleware de autenticación
	// Por ahora, asumimos que se incluye en el header Authorization
	// Esto se implementará en el middleware de autenticación posteriormente

	// Ejemplo básico, en la implementación real será más completo
	token, ok := ctx.Value("authorization").(string)
	if !ok || token == "" {
		return "", errors.NewBusinessError(errors.ErrUnauthorized, "token no encontrado en el contexto")
	}

	// Quitar el prefijo "Bearer " si existe
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	return token, nil
}
