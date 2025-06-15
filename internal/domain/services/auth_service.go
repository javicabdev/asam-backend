// Package services implements the business logic for the ASAM backend.
// It contains service implementations that fulfill the input port interfaces.
package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

type authService struct {
	userRepo  output.UserRepository
	jwtUtil   *auth.JWTUtil
	tokenRepo output.TokenRepository // Para gestionar tokens de refresh
	logger    logger.Logger
}

// NewAuthService crea una nueva instancia del servicio de autenticación
// que implementa la interfaz input.AuthService.
func NewAuthService(
	userRepo output.UserRepository,
	jwtUtil *auth.JWTUtil,
	tokenRepo output.TokenRepository,
	logger logger.Logger,
) input.AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtUtil:   jwtUtil,
		tokenRepo: tokenRepo,
		logger:    logger,
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
	s.logger.Info("Login attempt",
		zap.String("username", username),
		zap.String(string(constants.IPContextKey), getIPFromContext(ctx)),
		zap.String(string(constants.UserAgentContextKey), getUserAgentFromContext(ctx)),
	)

	// Buscar usuario por username
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Login failed: database error",
			zap.String("username", username),
			zap.Error(err),
		)
		return nil, errors.DB(err, "error buscando usuario")
	}
	if user == nil {
		s.logger.Warn("Login failed: user not found",
			zap.String("username", username),
		)
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "credenciales inválidas")
	}

	// Verificar contraseña
	if !user.CheckPassword(password) {
		// Debug: verificar qué está pasando
		s.logger.Warn("Login failed: invalid password",
			zap.String("username", username),
			zap.Uint("user_id", user.ID),
			zap.String("password_length", fmt.Sprintf("%d", len(password))),
			zap.String("hash_length", fmt.Sprintf("%d", len(user.Password))),
			zap.String("hash_prefix", user.Password[:10]), // Solo los primeros 10 caracteres del hash
		)

		// Generar un hash correcto para admin123
		correctHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		s.logger.Info("Generated correct hash for admin123",
			zap.String("correct_hash", string(correctHash)),
		)

		// Verificar si el hash generado funciona
		err = bcrypt.CompareHashAndPassword(correctHash, []byte("admin123"))
		s.logger.Info("Testing generated hash",
			zap.Bool("generated_hash_valid", err == nil),
		)

		// Comparar con el hash de la BD
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("admin123"))
		s.logger.Info("Testing DB hash directly",
			zap.Bool("db_hash_valid", err == nil),
			zap.String("db_hash_full", user.Password),
		)

		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "credenciales inválidas")
	}

	// Verificar que el usuario esté activo
	if !user.IsActive {
		s.logger.Warn("Login failed: inactive user",
			zap.String("username", username),
			zap.Uint("user_id", user.ID),
		)
		return nil, errors.NewBusinessError(errors.ErrInvalidStatus, "usuario inactivo")
	}

	// Generar tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error generando tokens")
	}

	// Guardar refresh token con información adicional del contexto
	ctxWithInfo := ctx
	if ip := ctx.Value(constants.IPContextKey); ip != nil {
		ctxWithInfo = context.WithValue(ctxWithInfo, constants.IPAddressContextKey, ip)
	}
	if ua := ctx.Value(constants.UserAgentContextKey); ua != nil {
		ctxWithInfo = context.WithValue(ctxWithInfo, constants.UserAgentContextKey, ua)
	}
	if device, ok := ctx.Value(constants.DeviceNameContextKey).(string); ok {
		ctxWithInfo = context.WithValue(ctxWithInfo, constants.DeviceNameContextKey, device)
	}

	err = s.tokenRepo.SaveRefreshToken(ctxWithInfo, td.RefreshUUID, user.ID, td.RtExpires)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error guardando refresh token")
	}

	// Aplicar límite de tokens por usuario si está configurado
	maxTokens := 5 // Valor por defecto, idealmente vendría de la configuración
	if err := s.tokenRepo.EnforceTokenLimitPerUser(ctx, maxTokens); err != nil {
		// Log el error pero no fallar el login
		s.logger.Warn("Failed to enforce token limit after login",
			zap.Error(err),
			zap.Uint("user_id", user.ID),
		)
	}

	// Actualizar último login
	user.LastLogin = time.Now()
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error actualizando último login")
	}

	// Al final del login exitoso:
	s.logger.Info("Login successful",
		zap.String("username", username),
		zap.Uint("user_id", user.ID),
		zap.String("role", string(user.Role)),
	)

	// Convertir auth.TokenDetails a input.TokenDetails
	return &input.TokenDetails{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		AccessUUID:   td.AccessUUID,
		RefreshUUID:  td.RefreshUUID,
		AtExpires:    td.AtExpires,
		RtExpires:    td.RtExpires,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	// 1. Validar access token
	token, err := s.jwtUtil.ValidateToken(accessToken, false)
	if err != nil {
		return errors.Wrap(err, errors.ErrUnauthorized, "token inválido")
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error extrayendo claims")
	}

	// 3. Eliminar refresh token asociado
	accessUUID, ok := claims["uuid"].(string)
	if !ok {
		return errors.NewBusinessError(errors.ErrUnauthorized, "uuid no encontrado en token")
	}

	err = s.tokenRepo.DeleteRefreshToken(ctx, accessUUID)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error eliminando refresh token")
	}

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*input.TokenDetails, error) {
	// 1. Validar refresh token
	token, err := s.jwtUtil.ValidateToken(refreshToken, true)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "refresh token inválido")
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error extrayendo claims")
	}

	// 3. Verificar que el refresh token exista en BD
	refreshUUID, ok := claims["uuid"].(string)
	if !ok {
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "uuid no encontrado en token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "user_id no encontrado en token")
	}

	// 4. Verificar token en BD
	err = s.tokenRepo.ValidateRefreshToken(ctx, refreshUUID, uint(userID))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "refresh token no válido")
	}

	// 5. Obtener usuario para verificar rol
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, errors.DB(err, "error obteniendo usuario")
	}
	if user == nil {
		return nil, errors.NewBusinessError(errors.ErrNotFound, "usuario no encontrado")
	}

	// 6. Generar nuevo par de tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error generando nuevos tokens")
	}

	// 7. Actualizar refresh token en BD con información del contexto
	err = s.tokenRepo.DeleteRefreshToken(ctx, refreshUUID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error eliminando refresh token antiguo")
	}

	// Preparar contexto con información adicional
	ctxWithInfo := ctx
	if ip := ctx.Value(constants.IPContextKey); ip != nil {
		ctxWithInfo = context.WithValue(ctxWithInfo, "ip_address", ip)
	}
	if ua := ctx.Value(constants.UserAgentContextKey); ua != nil {
		ctxWithInfo = context.WithValue(ctxWithInfo, string(constants.UserAgentContextKey), ua)
	}
	if device, ok := ctx.Value(constants.DeviceNameContextKey).(string); ok {
		ctxWithInfo = context.WithValue(ctxWithInfo, string(constants.DeviceNameContextKey), device)
	}

	err = s.tokenRepo.SaveRefreshToken(ctxWithInfo, td.RefreshUUID, user.ID, td.RtExpires)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error guardando nuevo refresh token")
	}

	// Convertir auth.TokenDetails a input.TokenDetails
	return &input.TokenDetails{
		AccessToken:  td.AccessToken,
		RefreshToken: td.RefreshToken,
		AccessUUID:   td.AccessUUID,
		RefreshUUID:  td.RefreshUUID,
		AtExpires:    td.AtExpires,
		RtExpires:    td.RtExpires,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// 1. Validar token
	token, err := s.jwtUtil.ValidateToken(tokenString, false)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "token inválido")
	}

	// 2. Extraer claims
	claims, err := s.jwtUtil.ExtractClaims(token)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error extrayendo claims")
	}

	// 3. Obtener user_id
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "user_id no encontrado en token")
	}

	// 4. Buscar usuario
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, errors.DB(err, "error obteniendo usuario")
	}
	if user == nil {
		return nil, errors.NewBusinessError(errors.ErrNotFound, "usuario no encontrado")
	}

	return user, nil
}
