// Package services implements the business logic for the ASAM backend.
// It contains service implementations that fulfill the input port interfaces.
package services

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

type authService struct {
	userRepo              output.UserRepository
	memberRepo            output.MemberRepository
	jwtUtil               *auth.JWTUtil
	tokenRepo             output.TokenRepository // Para gestionar tokens de refresh
	verificationTokenRepo output.VerificationTokenRepository
	emailVerificationSvc  input.EmailVerificationService
	logger                logger.Logger

	// Sliding expiration configuration
	slidingExpiration   bool
	slidingWindow       time.Duration
	absoluteMaxLifetime time.Duration
	inactivityTimeout   time.Duration
	maxTokensPerUser    int
}

// NewAuthService crea una nueva instancia del servicio de autenticación
// que implementa la interfaz input.AuthService.
func NewAuthService(
	userRepo output.UserRepository,
	memberRepo output.MemberRepository,
	jwtUtil *auth.JWTUtil,
	tokenRepo output.TokenRepository,
	verificationTokenRepo output.VerificationTokenRepository,
	emailVerificationSvc input.EmailVerificationService,
	logger logger.Logger,
	slidingExpiration bool,
	slidingWindow time.Duration,
	absoluteMaxLifetime time.Duration,
	inactivityTimeout time.Duration,
	maxTokensPerUser int,
) input.AuthService {
	return &authService{
		userRepo:              userRepo,
		memberRepo:            memberRepo,
		jwtUtil:               jwtUtil,
		tokenRepo:             tokenRepo,
		verificationTokenRepo: verificationTokenRepo,
		emailVerificationSvc:  emailVerificationSvc,
		logger:                logger,
		slidingExpiration:     slidingExpiration,
		slidingWindow:         slidingWindow,
		absoluteMaxLifetime:   absoluteMaxLifetime,
		inactivityTimeout:     inactivityTimeout,
		maxTokensPerUser:      maxTokensPerUser,
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
	// Normalize username if it's an email
	if strings.Contains(username, "@") {
		username = strings.ToLower(strings.TrimSpace(username))
	}

	// Registrar intento de login
	s.logger.Info("Login attempt",
		zap.String("username", username),
		zap.String("ip", getIPFromContext(ctx)),
		zap.String("user_agent", getUserAgentFromContext(ctx)),
	)

	// Validar credenciales
	user, err := s.validateCredentials(ctx, username, password)
	if err != nil {
		return nil, err
	}

	// Verificar estado del usuario
	if err := s.validateUserStatus(user); err != nil {
		return nil, err
	}

	// Para usuarios con rol USER, validar asociación con socio
	if user.Role == models.RoleUser {
		if err := s.validateMemberAssociation(ctx, user); err != nil {
			return nil, err
		}
	}

	// Generar y guardar tokens
	td, err := s.generateAndSaveTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	// Actualizar último login
	if err := s.updateLastLogin(ctx, user); err != nil {
		return nil, err
	}

	// Log login exitoso
	s.logSuccessfulLogin(user)

	return td, nil
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

// validateCredentials validates username/email and password
// Supports login with either username or email field
func (s *authService) validateCredentials(ctx context.Context, usernameOrEmail, password string) (*models.User, error) {
	// Intentar buscar usuario por username primero
	user, err := s.userRepo.FindByUsername(ctx, usernameOrEmail)
	if err != nil {
		s.logger.Error("Login failed: database error",
			zap.String("input", usernameOrEmail),
			zap.Error(err),
		)
		return nil, errors.DB(err, "error buscando usuario")
	}

	// Si no se encuentra por username, intentar por email
	if user == nil {
		user, err = s.userRepo.FindByEmail(ctx, usernameOrEmail)
		if err != nil {
			s.logger.Error("Login failed: database error searching by email",
				zap.String("input", usernameOrEmail),
				zap.Error(err),
			)
			return nil, errors.DB(err, "error buscando usuario")
		}
	}

	// Si aún no se encuentra, credenciales inválidas
	if user == nil {
		s.logger.Warn("Login failed: user not found",
			zap.String("input", usernameOrEmail),
		)
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "credenciales inválidas")
	}

	// Verificar contraseña
	if !user.CheckPassword(password) {
		s.logger.Warn("Login failed: invalid password",
			zap.String("input", usernameOrEmail),
			zap.Uint("user_id", user.ID),
		)
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "credenciales inválidas")
	}

	return user, nil
}

// validateUserStatus checks if the user is active
func (s *authService) validateUserStatus(user *models.User) error {
	if !user.IsActive {
		s.logger.Warn("Login failed: inactive user",
			zap.String("username", user.Username),
			zap.Uint("user_id", user.ID),
		)
		return errors.NewBusinessError(errors.ErrInvalidStatus, "usuario inactivo")
	}
	return nil
}

// validateMemberAssociation validates member association for USER role
func (s *authService) validateMemberAssociation(ctx context.Context, user *models.User) error {
	if user.MemberID == nil {
		s.logger.Warn("Login failed: user without associated member",
			zap.String("username", user.Username),
			zap.Uint("user_id", user.ID),
		)
		return errors.NewBusinessError(errors.ErrForbidden,
			"Tu usuario no está asociado a ningún socio. Contacta al administrador.")
	}

	// Verificar que el socio existe y está activo
	member, err := s.memberRepo.GetByID(ctx, *user.MemberID)
	if err != nil {
		s.logger.Error("Error fetching associated member",
			zap.Uint("member_id", *user.MemberID),
			zap.Error(err),
		)
		return errors.NewBusinessError(errors.ErrInternalError,
			"Error al verificar datos del socio")
	}

	if member == nil {
		s.logger.Error("Associated member not found",
			zap.Uint("member_id", *user.MemberID),
		)
		return errors.NewBusinessError(errors.ErrForbidden,
			"El socio asociado no existe. Contacta al administrador.")
	}

	if !member.IsActive() {
		s.logger.Warn("Login failed: inactive member",
			zap.String("username", user.Username),
			zap.Uint("user_id", user.ID),
			zap.Uint("member_id", member.ID),
		)
		return errors.NewBusinessError(errors.ErrForbidden,
			"El socio asociado está inactivo.")
	}

	// Precargar datos del socio para incluir en contexto
	user.Member = member
	return nil
}

// generateAndSaveTokens generates JWT tokens and saves refresh token
func (s *authService) generateAndSaveTokens(ctx context.Context, user *models.User) (*input.TokenDetails, error) {
	// Generar tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error generando tokens")
	}

	// Guardar refresh token con información adicional del contexto
	ctxWithInfo := s.enrichContextWithInfo(ctx)

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

// enrichContextWithInfo adds additional info to context
func (s *authService) enrichContextWithInfo(ctx context.Context) context.Context {
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
	return ctxWithInfo
}

// updateLastLogin updates the user's last login timestamp
func (s *authService) updateLastLogin(ctx context.Context, user *models.User) error {
	user.LastLogin = time.Now()
	err := s.userRepo.Update(ctx, user)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error actualizando último login")
	}
	return nil
}

// logSuccessfulLogin logs a successful login attempt
func (s *authService) logSuccessfulLogin(user *models.User) {
	logFields := []zap.Field{
		zap.String("username", user.Username),
		zap.Uint("user_id", user.ID),
		zap.String("role", string(user.Role)),
	}
	if user.MemberID != nil {
		logFields = append(logFields, zap.Uint("member_id", *user.MemberID))
	}
	s.logger.Info("Login successful", logFields...)
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

	// Validate that userID is not negative before conversion to uint
	if userID < 0 {
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "invalid user_id: negative value")
	}

	// 4. Verificar token en BD
	err = s.tokenRepo.ValidateRefreshToken(ctx, refreshUUID, uint(userID))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrUnauthorized, "refresh token no válido")
	}

	// 5. Obtener el token actual para verificar políticas de sliding expiration
	existingToken, err := s.tokenRepo.GetRefreshToken(ctx, refreshUUID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error obteniendo token existente")
	}

	// 6. Obtener usuario para verificar rol
	user, err := s.userRepo.FindByID(ctx, uint(userID))
	if err != nil {
		return nil, errors.DB(err, "error obteniendo usuario")
	}
	if user == nil {
		return nil, errors.NewBusinessError(errors.ErrNotFound, "usuario no encontrado")
	}

	// 7. Decidir si usar sliding expiration o crear un nuevo token
	var td *auth.TokenDetails
	var shouldExtendToken bool

	if s.slidingExpiration {
		shouldExtendToken, err = s.shouldApplySlidingExpiration(existingToken)
		if err != nil {
			// Si hay un error al verificar, mejor crear un nuevo token
			s.logger.Warn("Error checking sliding expiration eligibility, creating new token",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
			shouldExtendToken = false
		}
	}

	if shouldExtendToken {
		// Extender el token existente (sliding expiration)
		// Creamos un nuevo token pero con una expiración extendida
		td, err = s.createNewRefreshTokenWithSlidingExpiration(ctx, existingToken, user)
		if err != nil {
			return nil, err
		}
		s.logger.Info("Token extended using sliding expiration",
			zap.Uint("user_id", user.ID),
			zap.Int64("new_expires_at", td.RtExpires),
		)
	} else {
		// Crear un nuevo token (comportamiento tradicional)
		td, err = s.createNewRefreshToken(ctx, existingToken, user)
		if err != nil {
			return nil, err
		}
		s.logger.Info("New token created (sliding expiration limit reached or disabled)",
			zap.Uint("user_id", user.ID),
		)
	}

	// Aplicar límite de tokens por usuario si está configurado
	if s.maxTokensPerUser > 0 {
		if err := s.tokenRepo.EnforceTokenLimitPerUser(ctx, s.maxTokensPerUser); err != nil {
			// Log el error pero no fallar el refresh
			s.logger.Warn("Failed to enforce token limit after refresh",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
		}
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

// shouldApplySlidingExpiration verifica si se deben aplicar las políticas de sliding expiration
func (s *authService) shouldApplySlidingExpiration(token *models.RefreshToken) (bool, error) {
	now := time.Now()

	// Verificar si ha excedido el tiempo máximo absoluto de vida
	tokenAge := now.Sub(token.CreatedAt)
	if tokenAge >= s.absoluteMaxLifetime {
		s.logger.Info("Token exceeded absolute max lifetime, requiring new login",
			zap.Duration("token_age", tokenAge),
			zap.Duration("max_lifetime", s.absoluteMaxLifetime),
		)
		return false, nil
	}

	// Verificar si ha estado inactivo por mucho tiempo
	timeSinceLastUse := now.Sub(token.LastUsedAt)
	if timeSinceLastUse >= s.inactivityTimeout {
		s.logger.Info("Token exceeded inactivity timeout, requiring new login",
			zap.Duration("time_since_last_use", timeSinceLastUse),
			zap.Duration("inactivity_timeout", s.inactivityTimeout),
		)
		return false, nil
	}

	return true, nil
}

// createNewRefreshTokenWithSlidingExpiration crea un nuevo token respetando sliding expiration limits
func (s *authService) createNewRefreshTokenWithSlidingExpiration(ctx context.Context, oldToken *models.RefreshToken, user *models.User) (*auth.TokenDetails, error) {
	// Calcular nueva fecha de expiración (extender por sliding window desde ahora)
	newExpires := time.Now().Add(s.slidingWindow).Unix()

	// Asegurar que no exceda el límite absoluto desde la creación original
	maxAllowedExpires := oldToken.CreatedAt.Add(s.absoluteMaxLifetime).Unix()
	if newExpires > maxAllowedExpires {
		newExpires = maxAllowedExpires
		s.logger.Info("Sliding expiration capped at absolute max lifetime",
			zap.Uint("user_id", user.ID),
			zap.Time("max_allowed_expires", time.Unix(maxAllowedExpires, 0)),
		)
	}

	// Generar nuevo par de tokens con la expiración calculada
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error generando nuevos tokens")
	}

	// Actualizar la expiración del refresh token al valor calculado
	td.RtExpires = newExpires

	// Eliminar el token antiguo
	err = s.tokenRepo.DeleteRefreshToken(ctx, oldToken.UUID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error eliminando refresh token antiguo")
	}

	// Preparar contexto con información adicional
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

	// Guardar el nuevo refresh token con la expiración extendida
	err = s.tokenRepo.SaveRefreshToken(ctxWithInfo, td.RefreshUUID, user.ID, newExpires)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error guardando nuevo refresh token")
	}

	return td, nil
}

// createNewRefreshToken crea un nuevo token de refresco (comportamiento tradicional)
func (s *authService) createNewRefreshToken(ctx context.Context, oldToken *models.RefreshToken, user *models.User) (*auth.TokenDetails, error) {
	// Generar nuevo par de tokens
	td, err := s.jwtUtil.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error generando nuevos tokens")
	}

	// Eliminar refresh token antiguo
	err = s.tokenRepo.DeleteRefreshToken(ctx, oldToken.UUID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error eliminando refresh token antiguo")
	}

	// Preparar contexto con información adicional
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

	// Guardar nuevo refresh token
	err = s.tokenRepo.SaveRefreshToken(ctxWithInfo, td.RefreshUUID, user.ID, td.RtExpires)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error guardando nuevo refresh token")
	}

	return td, nil
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
		// Try to get user_id as other types
		var userIDInt uint
		switch v := claims["user_id"].(type) {
		case int:
			// Validate that the value is not negative before conversion
			if v < 0 {
				return nil, errors.NewBusinessError(errors.ErrUnauthorized, "invalid user_id: negative value")
			}
			userIDInt = uint(v)
		case int64:
			// Validate that the value is not negative before conversion
			if v < 0 {
				return nil, errors.NewBusinessError(errors.ErrUnauthorized, "invalid user_id: negative value")
			}
			userIDInt = uint(v)
		case uint:
			userIDInt = v
		case uint64:
			userIDInt = uint(v)
		default:

			return nil, errors.NewBusinessError(errors.ErrUnauthorized, "user_id no encontrado en token")
		}
		userID = float64(userIDInt)
	}

	// Validate that userID is not negative before conversion to uint
	if userID < 0 {
		return nil, errors.NewBusinessError(errors.ErrUnauthorized, "invalid user_id: negative value")
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

// ResetPasswordWithToken resets a user's password using a valid reset token
func (s *authService) ResetPasswordWithToken(ctx context.Context, token string, newPassword string) error {
	// Verify the reset token
	verificationToken, err := s.emailVerificationSvc.VerifyPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Get the user
	user, err := s.userRepo.FindByID(ctx, verificationToken.UserID)
	if err != nil {
		return errors.DB(err, "error obteniendo usuario")
	}
	if user == nil {
		return errors.NewBusinessError(errors.ErrNotFound, "usuario no encontrado")
	}

	// Update the password
	if err := user.SetPassword(newPassword); err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error estableciendo nueva contraseña")
	}

	// Save the updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "error actualizando usuario")
	}

	// Mark the token as used using atomic operation
	if err := s.verificationTokenRepo.MarkTokenAsUsed(ctx, verificationToken.ID); err != nil {
		// Log warning but don't fail since password was already reset
		s.logger.Warn("Failed to mark reset token as used", zap.Uint("tokenID", verificationToken.ID), zap.Error(err))
	}

	// Log the password reset
	s.logger.Info("Password reset successful",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("ip", getIPFromContext(ctx)),
	)

	return nil
}
