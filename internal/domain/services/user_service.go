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
	memberRepo   output.MemberRepository
	tokenRepo    output.VerificationTokenRepository
	emailService output.EmailService
	logger       logger.Logger
	baseURL      string // Base URL para construir links
}

// NewUserService creates a new user management service
func NewUserService(
	userRepo output.UserRepository,
	memberRepo output.MemberRepository,
	tokenRepo output.VerificationTokenRepository,
	emailService output.EmailService,
	logger logger.Logger,
	baseURL string,
) input.UserService {
	return &userService{
		userRepo:     userRepo,
		memberRepo:   memberRepo,
		tokenRepo:    tokenRepo,
		emailService: emailService,
		logger:       logger,
		baseURL:      baseURL,
	}
}

// CreateUser creates a new user with the given details
func (s *userService) CreateUser(ctx context.Context, username, email, password string, role models.Role, memberID *uint) (*models.User, error) {
	// Normalize username first (especially important for emails)
	username = strings.TrimSpace(username)
	if strings.Contains(username, "@") {
		username = strings.ToLower(username)
	}

	// Validate all inputs
	if err := s.validateCreateUserInput(ctx, username, password, role, memberID); err != nil {
		return nil, err
	}

	// Create user with validated data
	return s.createUserWithValidatedData(ctx, username, email, password, role, memberID)
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

	// Apply all updates
	if err := s.applyUserUpdates(ctx, user, updates); err != nil {
		return nil, err
	}

	// Update user in database
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

// applyUserUpdates applies all updates to the user object
func (s *userService) applyUserUpdates(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	// Update username if provided
	if err := s.updateUsername(ctx, user, updates); err != nil {
		return err
	}

	// Update email if provided
	if err := s.updateEmail(ctx, user, updates); err != nil {
		return err
	}

	// Update password if provided
	if err := s.updatePassword(user, updates); err != nil {
		return err
	}

	// Update role and memberID together (they must be consistent)
	if err := s.updateRoleAndMember(ctx, user, updates); err != nil {
		return err
	}

	// Update isActive if provided
	s.updateIsActive(user, updates)

	return nil
}

// updateUsername handles username update logic
func (s *userService) updateUsername(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	username, ok := updates["username"].(string)
	if !ok || username == "" {
		return nil
	}

	// Normalize username
	username = s.normalizeUsername(username)

	// Validate username
	if err := s.validateUsername(username); err != nil {
		return err
	}

	// Check if username is changing
	if username != user.Username {
		// Check if new username is available
		if err := s.checkUsernameAvailability(ctx, username, user.ID); err != nil {
			return err
		}

		// Update email verification status if changing to email
		if strings.Contains(username, "@") {
			user.EmailVerified = false
			user.EmailVerifiedAt = nil
		}
	}

	user.Username = username
	return nil
}

// updateEmail handles email update logic
func (s *userService) updateEmail(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	email, ok := updates["email"].(string)
	if !ok || email == "" {
		return nil
	}

	// Normalize email
	email = strings.TrimSpace(strings.ToLower(email))

	// Check if email is changing
	if email != user.Email {
		// Check if new email is available
		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil && !errors.IsNotFoundError(err) {
			return errors.DB(err, "error checking email availability")
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return errors.NewValidationError(
				"Email already exists",
				map[string]string{"email": "This email is already in use"},
			)
		}

		// Reset email verification status when email changes
		user.EmailVerified = false
		user.EmailVerifiedAt = nil

		s.logger.Info("Email changed, verification status reset",
			zap.Uint("user_id", user.ID),
			zap.String("old_email", user.Email),
			zap.String("new_email", email),
		)
	}

	user.Email = email
	return nil
}

// normalizeUsername normalizes the username format
func (s *userService) normalizeUsername(username string) string {
	username = strings.TrimSpace(username)
	if strings.Contains(username, "@") {
		username = strings.ToLower(username)
	}
	return username
}

// checkUsernameAvailability checks if a username is available for a specific user
func (s *userService) checkUsernameAvailability(ctx context.Context, username string, userID uint) error {
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil && !errors.IsNotFoundError(err) {
		return errors.DB(err, "error checking username availability")
	}
	if existingUser != nil && existingUser.ID != userID {
		return errors.NewValidationError(
			"Username already exists",
			map[string]string{"username": "This username is already taken"},
		)
	}
	return nil
}

// updatePassword handles password update logic
func (s *userService) updatePassword(user *models.User, updates map[string]interface{}) error {
	password, ok := updates["password"].(string)
	if !ok || password == "" {
		return nil
	}

	if err := s.validatePassword(password); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error hashing password")
	}

	user.Password = string(hashedPassword)
	return nil
}

// updateRoleAndMember handles role and member association update
func (s *userService) updateRoleAndMember(ctx context.Context, user *models.User, updates map[string]interface{}) error {
	// Determine final values
	finalRole, finalMemberID, hasRole, hasMemberID := s.determineFinalRoleAndMember(user, updates)

	// If neither role nor memberID is being updated, nothing to do
	if !hasRole && !hasMemberID {
		return nil
	}

	// Validate the combination
	if err := s.validateRoleMemberCombination(ctx, user, finalRole, finalMemberID); err != nil {
		return err
	}

	// Apply the updates
	user.Role = finalRole
	user.MemberID = finalMemberID

	return nil
}

// updateIsActive handles active status update
func (s *userService) updateIsActive(user *models.User, updates map[string]interface{}) {
	if isActive, ok := updates["isActive"].(bool); ok {
		user.IsActive = isActive
	}
}

// DeleteUser permanently deletes a user from the database
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
		adminCount, err := s.userRepo.CountUsersByRole(ctx, models.RoleAdmin)
		if err != nil {
			s.logger.Warn("Could not count admin users", zap.Error(err))
		} else if adminCount <= 1 {
			return errors.New(errors.ErrInvalidOperation, "Cannot delete the last admin user")
		}

		s.logger.Warn("Deleting an admin user",
			zap.Uint("user_id", user.ID),
			zap.String("username", user.Username),
		)
	}

	// Permanently delete the user (will cascade delete tokens)
	// This will fail if user has member_id set due to OnDelete:RESTRICT constraint
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return errors.DB(err, "error deleting user")
	}

	s.logger.Info("User permanently deleted",
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

// GetUserByEmail retrieves a user by email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Validate email format
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	// First, try to find user by username (in case username is the email)
	user, err := s.userRepo.FindByUsername(ctx, email)
	if err != nil {
		return nil, errors.DB(err, "error finding user by username")
	}

	// If found by username, return it
	if user != nil {
		// Clear password before returning
		user.Password = ""
		return user, nil
	}

	// If not found by username, try to find by email field
	user, err = s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.DB(err, "error finding user by email")
	}

	if user == nil {
		return nil, errors.NewNotFoundError("user")
	}

	// Clear password before returning
	user.Password = ""
	return user, nil
}

// ListUsers retrieves a paginated list of users
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10 // Default page size
	}

	// Get users from repository
	users, totalCount, err := s.userRepo.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, 0, errors.DB(err, "error listing users")
	}

	// Clear passwords before returning
	for _, user := range users {
		user.Password = ""
	}

	return users, totalCount, nil
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
		Type:      string(models.TokenTypeEmailVerification),
		Email:     user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours expiration
	}

	// Delete any existing verification tokens for this user
	if err := s.tokenRepo.InvalidateUserTokens(ctx, user.ID, string(models.TokenTypeEmailVerification)); err != nil {
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
	token, err := s.tokenRepo.GetByToken(ctx, tokenValue)
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
	// TODO: Implement CountActiveTokensByUser in repository
	// For now, skip rate limiting check
	/*
		count, err := s.tokenRepo.CountActiveTokensByUser(ctx, user.ID, models.TokenTypePasswordReset)
		if err != nil {
			return errors.DB(err, "error checking token count")
		}
		if count >= 3 {
			return errors.NewBusinessError(errors.ErrRateLimitExceeded, "too many password reset requests")
		}
	*/

	// Generate reset token
	tokenValue, err := utils.GeneratePasswordResetToken()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "error generating reset token")
	}

	// Create token in database
	token := &models.VerificationToken{
		Token:     tokenValue,
		UserID:    user.ID,
		Type:      string(models.TokenTypePasswordReset),
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
	token, err := s.tokenRepo.GetByToken(ctx, tokenValue)
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
	if token.Type != string(models.TokenTypePasswordReset) {
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
	if err := s.tokenRepo.InvalidateUserTokens(ctx, user.ID, string(models.TokenTypePasswordReset)); err != nil {
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

// passwordComplexity holds the results of password complexity analysis
type passwordComplexity struct {
	hasUpper   bool
	hasLower   bool
	hasNumber  bool
	hasSpecial bool
}

// validatePassword validates password requirements
func (s *userService) validatePassword(password string) error {
	// Validate password length
	if err := s.validatePasswordLength(password); err != nil {
		return err
	}

	// Check password complexity
	complexity := s.analyzePasswordComplexity(password)

	// Validate complexity requirements
	if err := s.validatePasswordComplexity(complexity); err != nil {
		return err
	}

	return nil
}

// validatePasswordLength validates password length requirements
func (s *userService) validatePasswordLength(password string) error {
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

	return nil
}

// analyzePasswordComplexity analyzes the character types in a password
func (s *userService) analyzePasswordComplexity(password string) passwordComplexity {
	complexity := passwordComplexity{}
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	for _, char := range password {
		switch {
		case isUpperCase(char):
			complexity.hasUpper = true
		case isLowerCase(char):
			complexity.hasLower = true
		case isNumber(char):
			complexity.hasNumber = true
		case strings.ContainsRune(specialChars, char):
			complexity.hasSpecial = true
		}
	}

	return complexity
}

// validatePasswordComplexity validates that password meets complexity requirements
func (s *userService) validatePasswordComplexity(complexity passwordComplexity) error {
	if !complexity.hasUpper || !complexity.hasLower || !complexity.hasNumber {
		return errors.NewValidationError(
			"Password does not meet complexity requirements",
			map[string]string{
				"password": "Password must contain at least one uppercase letter, one lowercase letter, and one number",
			},
		)
	}
	return nil
}

// Helper functions for character type checking
func isUpperCase(char rune) bool {
	return char >= 'A' && char <= 'Z'
}

func isLowerCase(char rune) bool {
	return char >= 'a' && char <= 'z'
}

func isNumber(char rune) bool {
	return char >= '0' && char <= '9'
}
