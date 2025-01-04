package services

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

type authService struct {
	userRepo  output.UserRepository
	jwtUtil   *auth.JWTUtil
	tokenRepo output.TokenRepository // Para gestionar tokens de refresh
}

func NewAuthService(
	userRepo output.UserRepository,
	jwtUtil *auth.JWTUtil,
	tokenRepo output.TokenRepository,
) input.AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtUtil:   jwtUtil,
		tokenRepo: tokenRepo,
	}
}

// Helper functions
func getIPFromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(constants.IPContextKey).(string); ok {
		return ip
	}
	return "unknown"
}

func getUserAgentFromContext(ctx context.Context) string {
	if ua, ok := ctx.Value(constants.UserAgentContextKey).(string); ok {
		return ua
	}
	return "unknown"
}

func (s *authService) Login(ctx context.Context, username, password string) (*input.TokenDetails, error) {
	// Registrar intento de login
	logger.Info("Login attempt",
		zap.String("username", username),
		zap.String("ip", getIPFromContext(ctx)),
		zap.String("user_agent", getUserAgentFromContext(ctx)),
	)

	// Buscar usuario por username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		logger.Error("Login failed: database error",
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error buscando usuario: %w", err)
	}
	if user == nil {
		logger.Warn("Login failed: user not found",
			zap.String("username", username),
		)
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Verificar contraseña
	if !user.CheckPassword(password) {
		logger.Warn("Login failed: invalid password",
			zap.String("username", username),
			zap.Uint("user_id", user.ID),
		)
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		logger.Warn("Login failed: inactive user",
			zap.String("username", username),
			zap.Uint("user_id", user.ID),
		)
		return nil, fmt.Errorf("usuario inactivo")
	}

	// Generar tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("error generando tokens: %w", err)
	}

	// Guardar refresh token
	err = s.tokenRepo.SaveRefreshToken(ctx, td.RefreshUuid, user.ID, td.RtExpires)
	if err != nil {
		return nil, fmt.Errorf("error guardando refresh token: %w", err)
	}

	// Actualizar último login
	user.LastLogin = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error actualizando último login: %w", err)
	}

	// Al final del login exitoso:
	logger.Info("Login successful",
		zap.String("username", username),
		zap.Uint("user_id", user.ID),
		zap.String("role", string(user.Role)),
	)

	// Convertir auth.TokenDetails a input.TokenDetails
	return &input.TokenDetails{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		AccessUuid:   td.AccessUuid,
		RefreshUuid:  td.RefreshUuid,
		AtExpires:    td.AtExpires,
		RtExpires:    td.RtExpires,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	// 1. Validar access token
	token, err := s.jwtUtil.ValidateToken(accessToken, false)
	if err != nil {
		return fmt.Errorf("token inválido: %w", err)
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return err
	}

	// 3. Eliminar refresh token asociado
	accessUuid, ok := claims["uuid"].(string)
	if !ok {
		return fmt.Errorf("uuid no encontrado en token")
	}

	err = s.tokenRepo.DeleteRefreshToken(ctx, accessUuid)
	if err != nil {
		return fmt.Errorf("error eliminando refresh token: %w", err)
	}

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*input.TokenDetails, error) {
	// 1. Validar refresh token
	token, err := s.jwtUtil.ValidateToken(refreshToken, true)
	if err != nil {
		return nil, fmt.Errorf("refresh token inválido: %w", err)
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return nil, err
	}

	// 3. Verificar que el refresh token exista en BD
	refreshUuid, ok := claims["uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("uuid no encontrado en token")
	}

	userId, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("user_id no encontrado en token")
	}

	// 4. Verificar token en BD
	err = s.tokenRepo.ValidateRefreshToken(ctx, refreshUuid, uint(userId))
	if err != nil {
		return nil, fmt.Errorf("refresh token no válido: %w", err)
	}

	// 5. Obtener usuario para verificar rol
	user, err := s.userRepo.FindByID(ctx, uint(userId))
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	// 6. Generar nuevo par de tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("error generando nuevos tokens: %w", err)
	}

	// 7. Actualizar refresh token en BD
	err = s.tokenRepo.DeleteRefreshToken(ctx, refreshUuid)
	if err != nil {
		return nil, fmt.Errorf("error eliminando refresh token antiguo: %w", err)
	}

	err = s.tokenRepo.SaveRefreshToken(ctx, td.RefreshUuid, user.ID, td.RtExpires)
	if err != nil {
		return nil, fmt.Errorf("error guardando nuevo refresh token: %w", err)
	}

	// Convertir auth.TokenDetails a input.TokenDetails
	return &input.TokenDetails{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		AccessUuid:   td.AccessUuid,
		RefreshUuid:  td.RefreshUuid,
		AtExpires:    td.AtExpires,
		RtExpires:    td.RtExpires,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// 1. Validar token
	token, err := s.jwtUtil.ValidateToken(tokenString, false)
	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return nil, err
	}

	// 3. Obtener user_id
	userId, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("user_id no encontrado en token")
	}

	// 4. Buscar usuario
	user, err := s.userRepo.FindByID(ctx, uint(userId))
	if err != nil {
		return nil, fmt.Errorf("error obteniendo usuario: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("usuario no encontrado")
	}

	return user, nil
}
