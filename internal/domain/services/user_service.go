package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/utils"
)

// userService implements user management operations
type userService struct {
	userRepo     output.UserRepository
	tokenRepo    output.VerificationTokenRepository
	emailService output.EmailService
	logger       logger.Logger
	baseURL      string // Base URL para construir links
}

// NewUserService creates a new user management service
func NewUserService(
	userRepo output.UserRepository,
	tokenRepo output.VerificationTokenRepository,
	emailService output.EmailService,
	logger logger.Logger,
	baseURL string,
) input.UserService {
	return &userService{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		logger:       logger,
		baseURL:      baseURL,
	}
}

// CreateUser creates a new user with the given details
func (s *userService) CreateUser(ctx context.Context, username, password string, role models.Role) (*models.User, error) {
	// Normalize username first (especially important for emails)
	username = strings.TrimSpace(username)
	if strings.Contains(username, "@") {
		username = strings.ToLower(username)
	}

	// Validate all inputs before any expensive operations
	if err := s.validateUsername(username); err != nil {
		return nil, err
	}

	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// Now check if username already exists (after all validations pass)
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.IsNotFoundError(err) {
		return nil, errors.DB(err, "error checking existing username")
	}
	if existingUser != nil {
		return nil, errors.NewValidationError(
			"Username already exists",
			map[string]string{"username": "This username is already taken"},
		)
	}

	// Hash password after all validations pass
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Create user
	user := &models.User{
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
		IsActive: true,
		// Por defecto, si el username es un email, no está verificado
		EmailVerified: !strings.Contains(username, "@"),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, errors.DB(err, "error creating user")
	}

	s.logger.Info("User created successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("role", string(user.Role)),
	)

	// Si el username es un email, enviar email de verificación
	if strings.Contains(username, "@") && !user.EmailVerified {
		// Intentar enviar email de verificación, pero no fallar si hay error
		if err := s.SendVerificationEmail(ctx, user.ID); err != nil {
			s.logger.Error("Failed to send verification email",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
		}
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// UpdateUser updates an existing user's details
func (s *userService) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Apply updates
	if username, ok := updates["username"].(string); ok && username != "" {
		// Normalize username first
		username = strings.TrimSpace(username)
		if strings.Contains(username, "@") {
			username = strings.ToLower(username)
		}

		// Validate username before any DB operations
		if err := s.validateUsername(username); err != nil {
			return nil, err
		}

		// Check if new username is taken by another user (only if it's different)
		if username != user.Username {
			existingUser, err := s.userRepo.FindByUsername(ctx, username)
			if err != nil && !errors.IsNotFoundError(err) {
				return nil, errors.DB(err, "error checking username availability")
			}
			if existingUser != nil && existingUser.ID != user.ID {
				return nil, errors.NewValidationError(
					"Username already exists",
					map[string]string{"username": "This username is already taken"},
				)
			}

			// Si cambia a un email, marcar como no verificado
			if strings.Contains(username, "@") {
				user.EmailVerified = false
				user.EmailVerifiedAt = nil
			}
		}
		user.Username = username
	}

	if password, ok := updates["password"].(string); ok && password != "" {
		if err := s.validatePassword(password); err != nil {
			return nil, err
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrInternalError, "error hashing password")
		}
		user.Password = string(hashedPassword)
	}

	if role, ok := updates["role"].(models.Role); ok {
		user.Role = role
	}

	if isActive, ok := updates["isActive"].(bool); ok {
		user.IsActive = isActive
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.DB(err, "error updating user")
	}

	s.logger.Info("User updated successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Prevent deleting the last admin
	if user.Role == models.RoleAdmin {
		// Count remaining admins (this would need a new repository method)
		// For now, we'll just log a warning
		s.logger.Warn("Deleting an admin user",
			zap.Uint("user_id", user.ID),
			zap.String("username", user.Username),
		)
	}

	// Soft delete by deactivating the user
	user.IsActive = false
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error deleting user")
	}

	s.logger.Info("User deleted successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(_ context.Context, _, _ int) ([]*models.User, error) {
	// For now, we'll return all users and handle pagination in memory
	// In a real implementation, this would be done at the repository level

	// This is a simplified implementation - you would need to add a ListUsers method to UserRepository
	// for proper pagination support

	// Placeholder implementation
	users := make([]*models.User, 0)

	// Clear passwords before returning
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

// ChangePassword changes a user's password
func (s *userService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Verify current password
	if !user.CheckPassword(currentPassword) {
		return errors.NewBusinessError(errors.ErrUnauthorized, "current password is incorrect")
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error updating password")
	}

	// Send notification email if username is an email
	if strings.Contains(user.Username, "@") {
		// Extract username for email
		displayName := utils.ExtractUsernameFromEmail(user.Username)
		if err := s.emailService.SendPasswordChangedEmail(ctx, user.Username, displayName); err != nil {
			s.logger.Error("Failed to send password changed email",
				zap.Error(err),
				zap.Uint("user_id", user.ID),
			)
		}
	}

	s.logger.Info("Password changed successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// ResetPassword resets a user's password (admin function)
func (s *userService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	// Validate new password first (before any DB operations)
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error resetting password")
	}

	s.logger.Info("Password reset successfully",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// SendVerificationEmail sends a verification email to the user
func (s *userService) SendVerificationEmail(ctx context.Context, userID uint) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Check if username is an email
	if !strings.Contains(user.Username, "@") {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "user does not have an email username")
	}

	// Check if already verified
	if user.EmailVerified {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "email already verified")
	}

	// Generate verification token
	tokenValue, err := utils.GenerateVerificationToken()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error generating verification token")
	}

	// Create token in database
	token := &models.VerificationToken{
		Token:     tokenValue,
		UserID:    user.ID,
		Type:      models.TokenTypeEmailVerification,
		Email:     user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiration
	}

	// Delete any existing verification tokens for this user
	if err := s.tokenRepo.DeleteUserTokensByType(ctx, user.ID, models.TokenTypeEmailVerification); err != nil {
		s.logger.Error("Failed to delete existing verification tokens",
			zap.Error(err),
			zap.Uint("user_id", user.ID),
		)
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return errors.DB(err, "error creating verification token")
	}

	// Build verification URL
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, tokenValue)

	// Send email
	displayName := utils.ExtractUsernameFromEmail(user.Username)
	if err := s.emailService.SendVerificationEmail(ctx, user.Username, displayName, verificationURL); err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error sending verification email")
	}

	s.logger.Info("Verification email sent",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Username),
	)

	return nil
}

// VerifyEmail verifies a user's email with the provided token
func (s *userService) VerifyEmail(ctx context.Context, tokenValue string) error {
	// Find token
	token, err := s.tokenRepo.FindByToken(ctx, tokenValue)
	if err != nil {
		return errors.DB(err, "error finding verification token")
	}
	if token == nil {
		return errors.NewBusinessError(errors.ErrNotFound, "invalid or expired token")
	}

	// Check if token is valid
	if !token.IsValid() {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "token is expired or already used")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Verify email matches
	if user.Username != token.Email {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "email mismatch")
	}

	// Update user
	user.EmailVerified = true
	now := time.Now()
	user.EmailVerifiedAt = &now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error updating user")
	}

	// Mark token as used
	token.MarkAsUsed()
	if err := s.tokenRepo.Update(ctx, token); err != nil {
		s.logger.Error("Failed to mark token as used",
			zap.Error(err),
			zap.String("token", tokenValue),
		)
	}

	s.logger.Info("Email verified successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Username),
	)

	return nil
}

// RequestPasswordReset initiates a password reset for the given email
func (s *userService) RequestPasswordReset(ctx context.Context, email string) error {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return err
	}

	// Find user by email
	user, err := s.userRepo.FindByUsername(ctx, email)
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.DB(err, "error finding user")
	}

	// Don't reveal if user exists or not for security
	if user == nil {
		s.logger.Warn("Password reset requested for non-existent email",
			zap.String("email", email),
		)
		// Return success to prevent email enumeration
		return nil
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Password reset requested for inactive user",
			zap.String("email", email),
			zap.Uint("user_id", user.ID),
		)
		// Return success to prevent email enumeration
		return nil
	}

	// Check rate limiting (max 3 tokens per hour)
	count, err := s.tokenRepo.CountActiveTokensByUser(ctx, user.ID, models.TokenTypePasswordReset)
	if err != nil {
		return errors.DB(err, "error checking token count")
	}
	if count >= 3 {
		return errors.NewBusinessError(errors.ErrRateLimitExceeded, "too many password reset requests")
	}

	// Generate reset token
	tokenValue, err := utils.GeneratePasswordResetToken()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error generating reset token")
	}

	// Create token in database
	token := &models.VerificationToken{
		Token:     tokenValue,
		UserID:    user.ID,
		Type:      models.TokenTypePasswordReset,
		Email:     user.Username,
		ExpiresAt: time.Now().Add(1 * time.Hour), // 1 hour expiration
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return errors.DB(err, "error creating reset token")
	}

	// Build reset URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, tokenValue)

	// Send email
	displayName := utils.ExtractUsernameFromEmail(user.Username)
	if err := s.emailService.SendPasswordResetEmail(ctx, user.Username, displayName, resetURL); err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error sending reset email")
	}

	s.logger.Info("Password reset email sent",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Username),
	)

	return nil
}

// ResetPasswordWithToken resets password using a valid token
func (s *userService) ResetPasswordWithToken(ctx context.Context, tokenValue, newPassword string) error {
	// Validate new password first (before any DB operations)
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Find token
	token, err := s.tokenRepo.FindByToken(ctx, tokenValue)
	if err != nil {
		return errors.DB(err, "error finding reset token")
	}
	if token == nil {
		return errors.NewBusinessError(errors.ErrNotFound, "invalid or expired token")
	}

	// Check if token is valid
	if !token.IsValid() {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "token is expired or already used")
	}

	// Check token type
	if token.Type != models.TokenTypePasswordReset {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "invalid token type")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return errors.DB(err, "error finding user")
	}
	if user == nil {
		return errors.NewNotFoundError("user")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	// Update password
	user.Password = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.DB(err, "error updating password")
	}

	// Mark token as used
	token.MarkAsUsed()
	if err := s.tokenRepo.Update(ctx, token); err != nil {
		s.logger.Error("Failed to mark token as used",
			zap.Error(err),
			zap.String("token", tokenValue),
		)
	}

	// Delete all other password reset tokens for this user
	if err := s.tokenRepo.DeleteUserTokensByType(ctx, user.ID, models.TokenTypePasswordReset); err != nil {
		s.logger.Error("Failed to delete other reset tokens",
			zap.Error(err),
			zap.Uint("user_id", user.ID),
		)
	}

	// Send notification email
	displayName := utils.ExtractUsernameFromEmail(user.Username)
	if err := s.emailService.SendPasswordChangedEmail(ctx, user.Username, displayName); err != nil {
		s.logger.Error("Failed to send password changed email",
			zap.Error(err),
			zap.Uint("user_id", user.ID),
		)
	}

	s.logger.Info("Password reset successfully with token",
		zap.Uint("user_id", user.ID),
		zap.String("username", user.Username),
	)

	return nil
}

// ResendVerificationEmail resends the verification email
func (s *userService) ResendVerificationEmail(ctx context.Context, email string) error {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Find user by email
	user, err := s.userRepo.FindByUsername(ctx, email)
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.DB(err, "error finding user")
	}

	// Don't reveal if user exists or not for security
	if user == nil {
		s.logger.Warn("Verification resend requested for non-existent email",
			zap.String("email", email),
		)
		// Return success to prevent email enumeration
		return nil
	}

	// Check if already verified
	if user.EmailVerified {
		return errors.NewBusinessError(errors.ErrInvalidRequest, "email already verified")
	}

	// Send verification email
	return s.SendVerificationEmail(ctx, user.ID)
}

// validateUsername validates username requirements
// Accepts both regular usernames and email addresses
func (s *userService) validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return errors.NewValidationError(
			"Username is required",
			map[string]string{"username": "Username cannot be empty"},
		)
	}

	if len(username) < 3 {
		return errors.NewValidationError(
			"Username too short",
			map[string]string{"username": "Username must be at least 3 characters long"},
		)
	}

	if len(username) > 100 { // Increased to accommodate email addresses
		return errors.NewValidationError(
			"Username too long",
			map[string]string{"username": "Username must not exceed 100 characters"},
		)
	}

	// Check if it looks like an email
	if strings.Contains(username, "@") {
		return s.validateEmail(username)
	}

	// Otherwise validate as regular username
	return s.validateRegularUsername(username)
}

// validateRegularUsername validates non-email usernames
func (s *userService) validateRegularUsername(username string) error {
	// Check for valid characters (alphanumeric, underscore, hyphen, dot)
	for _, char := range username {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '_' && char != '-' && char != '.' {
			return errors.NewValidationError(
				"Invalid username format",
				map[string]string{"username": "Username can only contain letters, numbers, underscore, hyphen, and dot"},
			)
		}
	}

	return nil
}

// validateEmail validates email format
func (s *userService) validateEmail(email string) error {
	// Basic email validation regex
	// This regex covers most common email formats
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error validating email format")
	}

	if !matched {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Please provide a valid email address"},
		)
	}

	// Additional validation rules
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Email must contain exactly one @ symbol"},
		)
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Validate local part
	if len(localPart) == 0 || len(localPart) > 64 {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Email local part must be between 1 and 64 characters"},
		)
	}

	// Check for consecutive dots
	if strings.Contains(email, "..") {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Email cannot contain consecutive dots"},
		)
	}

	// Check if email starts or ends with a dot
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Email local part cannot start or end with a dot"},
		)
	}

	// Validate domain part
	if len(domainPart) < 3 || len(domainPart) > 255 {
		return errors.NewValidationError(
			"Invalid email format",
			map[string]string{"username": "Email domain must be between 3 and 255 characters"},
		)
	}

	return nil
}

// validatePassword validates password requirements
func (s *userService) validatePassword(password string) error {
	if password == "" {
		return errors.NewValidationError(
			"Password is required",
			map[string]string{"password": "Password cannot be empty"},
		)
	}

	if len(password) < 8 {
		return errors.NewValidationError(
			"Password too short",
			map[string]string{"password": "Password must be at least 8 characters long"},
		)
	}

	if len(password) > 100 {
		return errors.NewValidationError(
			"Password too long",
			map[string]string{"password": "Password must not exceed 100 characters"},
		)
	}

	// Check password complexity
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return errors.NewValidationError(
			"Password does not meet complexity requirements",
			map[string]string{
				"password": "Password must contain at least one uppercase letter, one lowercase letter, and one number",
			},
		)
	}

	// hasSpecial is optional but recommended
	if !hasSpecial {
		s.logger.Debug("Password does not contain special characters",
			zap.String("recommendation", "Consider adding special characters for better security"),
		)
	}

	return nil
}
