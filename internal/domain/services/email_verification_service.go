package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// emailVerificationService implements the email verification service business logic
type emailVerificationService struct {
	logger              logger.Logger
	emailNotifier       input.EmailNotificationService
	tokenRepo           output.VerificationTokenRepository
	userRepo            output.UserRepository
	verificationBaseURL string
	resetBaseURL        string
}

// NewEmailVerificationService creates a new email verification service instance
func NewEmailVerificationService(
	logger logger.Logger,
	emailNotifier input.EmailNotificationService,
	tokenRepo output.VerificationTokenRepository,
	userRepo output.UserRepository,
	verificationBaseURL string,
	resetBaseURL string,
) input.EmailVerificationService {
	return &emailVerificationService{
		logger:              logger,
		emailNotifier:       emailNotifier,
		tokenRepo:           tokenRepo,
		userRepo:            userRepo,
		verificationBaseURL: verificationBaseURL,
		resetBaseURL:        resetBaseURL,
	}
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateToken creates a token with the given parameters
func (s *emailVerificationService) generateToken(ctx context.Context, userID uint, email string, tokenType models.TokenType, expiresIn time.Duration) (string, error) {
	// Invalidate any existing tokens of this type for this user
	if err := s.tokenRepo.InvalidateUserTokens(ctx, userID, string(tokenType)); err != nil {
		s.logger.Warn("Failed to invalidate existing tokens",
			zap.Uint("userID", userID),
			zap.String("tokenType", string(tokenType)),
			zap.Error(err))
	}

	// Generate new token
	tokenValue, err := generateSecureToken()
	if err != nil {
		return "", errors.Wrap(err, errors.ErrInternalError, "failed to generate token")
	}

	// Create token record
	token := &models.VerificationToken{
		UserID:    userID,
		Token:     tokenValue,
		Type:      string(tokenType),
		Email:     email,
		ExpiresAt: time.Now().Add(expiresIn),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return "", errors.Wrap(err, errors.ErrDatabaseError, "failed to save token")
	}

	return tokenValue, nil
}

// GenerateVerificationToken generates and stores a new email verification token
func (s *emailVerificationService) GenerateVerificationToken(ctx context.Context, userID uint) (string, error) {
	// Get user to obtain email
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrNotFound, "user not found")
	}

	// Determine which email to use
	emailToUse := user.Email
	if emailToUse == "" && strings.Contains(user.Username, "@") {
		emailToUse = user.Username
	}
	if emailToUse == "" {
		return "", errors.NewBusinessError(errors.ErrInvalidRequest, "user has no email address")
	}

	return s.generateToken(ctx, userID, emailToUse, models.TokenTypeEmailVerification, 24*time.Hour)
}

// GeneratePasswordResetToken generates and stores a new password reset token
func (s *emailVerificationService) GeneratePasswordResetToken(ctx context.Context, userID uint) (string, error) {
	// Get user to obtain email
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrNotFound, "user not found")
	}

	// Determine which email to use (same logic as in SendPasswordResetEmailToUser)
	emailToUse := user.Email
	if strings.Contains(user.Username, "@") {
		emailToUse = user.Username
	}
	if emailToUse == "" {
		return "", errors.NewBusinessError(errors.ErrInvalidRequest, "user has no email address")
	}

	return s.generateToken(ctx, userID, emailToUse, models.TokenTypePasswordReset, 1*time.Hour)
}

// VerifyEmailToken verifies an email verification token and marks the user's email as verified
func (s *emailVerificationService) VerifyEmailToken(ctx context.Context, tokenValue string) (*models.User, error) {
	// Get token
	token, err := s.tokenRepo.GetByToken(ctx, tokenValue)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrNotFound, "token not found")
	}

	// Validate token
	if !token.IsValid() {
		if token.IsExpired() {
			return nil, errors.NewBusinessError(errors.ErrInvalidToken, "verification token has expired")
		}
		if token.IsUsed() {
			return nil, errors.NewBusinessError(errors.ErrInvalidToken, "verification token has already been used")
		}
		return nil, errors.NewBusinessError(errors.ErrInvalidToken, "invalid verification token")
	}

	// Check token type
	if token.Type != string(models.TokenTypeEmailVerification) {
		return nil, errors.NewBusinessError(errors.ErrInvalidToken, "invalid token type")
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrNotFound, "user not found")
	}

	// Mark email as verified
	user.EmailVerified = true
	now := time.Now()
	user.EmailVerifiedAt = &now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "failed to update user")
	}

	// Mark token as used using atomic operation
	if err := s.tokenRepo.MarkTokenAsUsed(ctx, token.ID); err != nil {
		// Log warning but don't fail since email was already verified
		s.logger.Warn("Failed to mark token as used", zap.Uint("tokenID", token.ID), zap.Error(err))
	}

	return user, nil
}

// VerifyPasswordResetToken verifies a password reset token
func (s *emailVerificationService) VerifyPasswordResetToken(ctx context.Context, tokenValue string) (*models.VerificationToken, error) {
	// Get token
	token, err := s.tokenRepo.GetByToken(ctx, tokenValue)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrNotFound, "token not found")
	}

	// Validate token
	if !token.IsValid() {
		if token.IsExpired() {
			return nil, errors.NewBusinessError(errors.ErrInvalidToken, "reset token has expired")
		}
		if token.IsUsed() {
			return nil, errors.NewBusinessError(errors.ErrInvalidToken, "reset token has already been used")
		}
		return nil, errors.NewBusinessError(errors.ErrInvalidToken, "invalid reset token")
	}

	// Check token type
	if token.Type != string(models.TokenTypePasswordReset) {
		return nil, errors.NewBusinessError(errors.ErrInvalidToken, "invalid token type")
	}

	return token, nil
}

// SendVerificationEmailToUser generates a token and sends verification email
func (s *emailVerificationService) SendVerificationEmailToUser(ctx context.Context, user *models.User) error {
	// Check if email was sent recently (anti-spam protection)
	if user.EmailVerificationSentAt != nil {
		timeSinceLastEmail := time.Since(*user.EmailVerificationSentAt)
		minWaitTime := 1 * time.Minute

		if timeSinceLastEmail < minWaitTime {
			waitTime := minWaitTime - timeSinceLastEmail
			return errors.NewBusinessError(
				errors.ErrRateLimitExceeded,
				fmt.Sprintf("Please wait %d minutes before requesting another verification email", int(waitTime.Minutes())),
			)
		}
	}

	// Generate verification token
	token, err := s.GenerateVerificationToken(ctx, user.ID)
	if err != nil {
		return err
	}

	// Build verification URL
	verificationURL := fmt.Sprintf("%s?token=%s", s.verificationBaseURL, token)

	// Send email
	if err := s.emailNotifier.SendVerificationEmail(ctx, user, verificationURL); err != nil {
		// If email fails, we should probably delete the token
		// but for now we'll just log the error
		s.logger.Error("Failed to send verification email", zap.Uint("userID", user.ID), zap.Error(err))
		return errors.Wrap(err, errors.ErrInternalError, "failed to send verification email")
	}

	// Update EmailVerificationSentAt timestamp
	now := time.Now()
	user.EmailVerificationSentAt = &now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail since email was already sent
		s.logger.Warn("Failed to update EmailVerificationSentAt",
			zap.Uint("userID", user.ID),
			zap.Error(err))
	}

	return nil
}

// SendPasswordResetEmailToUser generates a token and sends reset email
func (s *emailVerificationService) SendPasswordResetEmailToUser(ctx context.Context, user *models.User) error {
	// Check if password reset email was sent recently (anti-spam protection)
	// We can use the same field since both are email verifications
	if user.EmailVerificationSentAt != nil {
		timeSinceLastEmail := time.Since(*user.EmailVerificationSentAt)
		minWaitTime := 5 * time.Minute

		if timeSinceLastEmail < minWaitTime {
			waitTime := minWaitTime - timeSinceLastEmail
			return errors.NewBusinessError(
				errors.ErrRateLimitExceeded,
				fmt.Sprintf("Please wait %d minutes before requesting another password reset email", int(waitTime.Minutes())),
			)
		}
	}

	// Determine which email to use
	// If username is an email, use that; otherwise use the email field
	emailToUse := user.Email
	if strings.Contains(user.Username, "@") {
		emailToUse = user.Username
	} else if emailToUse == "" {
		s.logger.Error("User has no email address", zap.Uint("userID", user.ID))
		return errors.NewBusinessError(errors.ErrInvalidRequest, "user has no email address")
	}

	// Generate reset token
	token, err := s.GeneratePasswordResetToken(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to generate reset token", zap.Uint("userID", user.ID), zap.Error(err))
		return err
	}

	// Build reset URL
	resetURL := fmt.Sprintf("%s?token=%s", s.resetBaseURL, token)

	// Send email

	// Create a copy of the user with the correct email
	userCopy := *user
	userCopy.Email = emailToUse // Use the determined email

	if err := s.emailNotifier.SendPasswordResetEmail(ctx, &userCopy, resetURL); err != nil {
		s.logger.Error("Failed to send password reset email", zap.Uint("userID", user.ID), zap.Error(err))
		return errors.Wrap(err, errors.ErrInternalError, "failed to send password reset email")
	}

	// Update EmailVerificationSentAt timestamp (we use the same field for all verification emails)
	now := time.Now()
	user.EmailVerificationSentAt = &now
	user.UpdatedAt = now

	if err := s.userRepo.Update(ctx, user); err != nil {
		// Log error but don't fail since email was already sent
		s.logger.Warn("Failed to update EmailVerificationSentAt for password reset",
			zap.Uint("userID", user.ID),
			zap.Error(err))
	}

	return nil
}

// CleanupExpiredTokens removes all expired tokens from the database
func (s *emailVerificationService) CleanupExpiredTokens(ctx context.Context) error {
	return s.tokenRepo.DeleteExpired(ctx)
}
