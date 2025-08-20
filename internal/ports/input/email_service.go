package input

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// EmailNotificationService defines the interface for sending email notifications
type EmailNotificationService interface {
	// SendVerificationEmail sends an email verification link to the user
	SendVerificationEmail(ctx context.Context, user *models.User, verificationURL string) error

	// SendPasswordResetEmail sends a password reset link to the user
	SendPasswordResetEmail(ctx context.Context, user *models.User, resetURL string) error

	// SendWelcomeEmail sends a welcome email to a new user
	SendWelcomeEmail(ctx context.Context, user *models.User) error

	// SendPasswordChangedEmail sends a notification that password was changed
	SendPasswordChangedEmail(ctx context.Context, user *models.User) error
}

// EmailVerificationService defines the interface for email verification operations
type EmailVerificationService interface {
	// GenerateVerificationToken generates and stores a new email verification token
	GenerateVerificationToken(ctx context.Context, userID uint) (string, error)

	// GeneratePasswordResetToken generates and stores a new password reset token
	GeneratePasswordResetToken(ctx context.Context, userID uint) (string, error)

	// VerifyEmailToken verifies an email verification token and marks the user's email as verified
	VerifyEmailToken(ctx context.Context, token string) (*models.User, error)

	// VerifyPasswordResetToken verifies a password reset token
	VerifyPasswordResetToken(ctx context.Context, token string) (*models.VerificationToken, error)

	// SendVerificationEmailToUser generates a token and sends verification email
	SendVerificationEmailToUser(ctx context.Context, user *models.User) error

	// SendPasswordResetEmailToUser generates a token and sends reset email
	SendPasswordResetEmailToUser(ctx context.Context, user *models.User) error

	// CleanupExpiredTokens removes all expired tokens from the database
	CleanupExpiredTokens(ctx context.Context) error
}
