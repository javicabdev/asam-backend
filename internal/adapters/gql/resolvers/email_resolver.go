package resolvers

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// emailResolver contains email-related resolvers
type emailResolver struct {
	*Resolver
}

// SendVerificationEmail sends an email verification link to the current user
func (r *emailResolver) SendVerificationEmail(ctx context.Context) (*model.MutationResponse, error) {
	// Debug logging
	r.logger.Info("[DEBUG] SendVerificationEmail called",
		zap.Any("contextKeys", ctx),
		zap.Bool("hasUserKey", ctx.Value(constants.UserContextKey) != nil),
		zap.String("userKeyType", fmt.Sprintf("%T", ctx.Value(constants.UserContextKey))),
	)

	// Get current user from context
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		r.logger.Error("[DEBUG] SendVerificationEmail: User not found in context",
			zap.Bool("contextHasUser", ctx.Value(constants.UserContextKey) != nil),
			zap.String("contextType", fmt.Sprintf("%T", ctx.Value(constants.UserContextKey))),
			zap.Bool("typeAssertionOk", ok),
			zap.Bool("userIsNil", user == nil),
		)
		return nil, errors.NewUnauthorizedError()
	}

	// Check if email is already verified
	if user.EmailVerified {
		msg := "Email already verified"
		return &model.MutationResponse{
			Success: true,
			Message: &msg,
		}, nil
	}

	// Check if user has an email
	if user.Email == "" {
		errMsg := "User does not have an email address"
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	// Send verification email
	if err := r.emailVerificationService.SendVerificationEmailToUser(ctx, user); err != nil {
		errMsg := "Failed to send verification email"
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	msg := "Verification email sent successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &msg,
	}, nil
}

// VerifyEmail verifies a user's email address using a token
func (r *emailResolver) VerifyEmail(ctx context.Context, token string) (*model.MutationResponse, error) {
	// Verify the email token
	user, err := r.emailVerificationService.VerifyEmailToken(ctx, token)
	if err != nil {
		// Check if it's a "token already used" error
		if appErr, ok := errors.AsAppError(err); ok {
			if appErr.Code == errors.ErrInvalidToken && appErr.Message == "verification token has already been used" {
				// Try to get the user associated with this token to check if they're already verified
				// This is a workaround for frontend issues where the token might be verified multiple times
				r.logger.Warn("Verification token already used, checking if user is verified",
					zap.String("token", token[:8]+"..."), // Log only first 8 chars for security
				)

				// For now, return a success message indicating the email is already verified
				msg := "Email is already verified"
				return &model.MutationResponse{
					Success: true,
					Message: &msg,
				}, nil
			}
		}

		// For other errors, return the error message
		errMsg := "Invalid or expired verification token"
		if appErr, ok := errors.AsAppError(err); ok {
			errMsg = appErr.Message
		}

		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	// Send welcome email
	if err := r.emailNotificationService.SendWelcomeEmail(ctx, user); err != nil {
		// Log the error but don't fail the verification
		r.logger.Warn("Failed to send welcome email", zap.Uint("userID", user.ID), zap.Error(err))
	}

	msg := "Email verified successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &msg,
	}, nil
}

// ResendVerificationEmail resends the verification email to a specific email address
func (r *emailResolver) ResendVerificationEmail(ctx context.Context, email string) (*model.MutationResponse, error) {
	// Find user by email
	user, err := r.userService.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		errMsg := "If an account exists with this email, a verification email will be sent"
		//nolint:nilerr // Intentionally returning nil error for security reasons
		return &model.MutationResponse{
			Success: true,
			Message: &errMsg,
		}, nil
	}

	// Check if email is already verified
	if user.EmailVerified {
		msg := "Email already verified"
		return &model.MutationResponse{
			Success: true,
			Message: &msg,
		}, nil
	}

	// Send verification email
	if err := r.emailVerificationService.SendVerificationEmailToUser(ctx, user); err != nil {
		// Log error but return success for security
		r.logger.Error("Failed to send verification email", zap.String("email", email), zap.Error(err))
	}

	msg := "If an account exists with this email, a verification email will be sent"
	return &model.MutationResponse{
		Success: true,
		Message: &msg,
	}, nil
}

// RequestPasswordReset sends a password reset email
func (r *emailResolver) RequestPasswordReset(ctx context.Context, email string) (*model.MutationResponse, error) {
	// Log the start of password reset request
	r.logger.Info("[PASSWORD-RESET] RequestPasswordReset called",
		zap.String("email", email),
		zap.String("clientIP", fmt.Sprintf("%v", ctx.Value(constants.IPContextKey))),
	)

	// Find user by email (now with cascade search: username first, then email field)
	r.logger.Debug("[PASSWORD-RESET] Looking up user by email", zap.String("email", email))
	user, err := r.userService.GetUserByEmail(ctx, email)
	if err != nil {
		r.logger.Warn("[PASSWORD-RESET] User not found or error occurred",
			zap.String("email", email),
			zap.Error(err),
			zap.String("errorType", fmt.Sprintf("%T", err)),
		)
		// Don't reveal if email exists or not for security
		msg := "If an account exists with this email, a password reset link will be sent"
		return &model.MutationResponse{
			Success: true,
			Message: &msg,
		}, nil
	}

	// Log user found
	r.logger.Info("[PASSWORD-RESET] User found, preparing to send reset email",
		zap.Uint("userID", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.Bool("emailVerified", user.EmailVerified),
		zap.Bool("isActive", user.IsActive),
	)

	// Send password reset email
	r.logger.Debug("[PASSWORD-RESET] Calling SendPasswordResetEmailToUser")
	if err := r.emailVerificationService.SendPasswordResetEmailToUser(ctx, user); err != nil {
		// Log error but return success for security
		r.logger.Error("[PASSWORD-RESET] Failed to send password reset email",
			zap.String("email", email),
			zap.Uint("userID", user.ID),
			zap.Error(err),
			zap.String("errorType", fmt.Sprintf("%T", err)),
		)
	} else {
		r.logger.Info("[PASSWORD-RESET] SendPasswordResetEmailToUser completed successfully",
			zap.String("email", email),
			zap.Uint("userID", user.ID),
		)
	}

	msg := "Si existe una cuenta con esta dirección de correo electrónico, se enviará un enlace para restablecer la contraseña."
	return &model.MutationResponse{
		Success: true,
		Message: &msg,
	}, nil
}

// ResetPasswordWithToken resets a user's password using a valid reset token
func (r *emailResolver) ResetPasswordWithToken(ctx context.Context, token string, newPassword string) (*model.MutationResponse, error) {
	// Validate password strength
	if len(newPassword) < 8 {
		errMsg := "Password must be at least 8 characters long"
		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, nil
	}

	// Reset password using token
	if err := r.authService.ResetPasswordWithToken(ctx, token, newPassword); err != nil {
		// Determine error message based on error type
		errMsg := "Invalid or expired reset token"
		if appErr, ok := errors.AsAppError(err); ok {
			errMsg = appErr.Message
		}

		return &model.MutationResponse{
			Success: false,
			Error:   &errMsg,
		}, err
	}

	// Get the user to send notification email
	verificationToken, _ := r.emailVerificationService.VerifyPasswordResetToken(ctx, token)
	if verificationToken != nil {
		user, _ := r.userService.GetUser(ctx, verificationToken.UserID)
		if user != nil {
			// Send password changed notification
			if err := r.emailNotificationService.SendPasswordChangedEmail(ctx, user); err != nil {
				// Log the error but don't fail the reset
				r.logger.Warn("Failed to send password changed email", zap.Uint("userID", user.ID), zap.Error(err))
			}
		}
	}

	msg := "Password reset successfully"
	return &model.MutationResponse{
		Success: true,
		Message: &msg,
	}, nil
}
