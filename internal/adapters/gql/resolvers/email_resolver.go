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
	// Log all context values at the very beginning
	r.logger.Info("SendVerificationEmail: START - Checking all context values",
		zap.Bool("hasUserContext", ctx.Value(constants.UserContextKey) != nil),
		zap.Bool("isAuthorized", ctx.Value(constants.AuthorizedContextKey) != nil),
		zap.Bool("hasAuthToken", ctx.Value(constants.AuthTokenContextKey) != nil),
		zap.Bool("hasUserID", ctx.Value(constants.UserIDContextKey) != nil),
		zap.Bool("hasUserRole", ctx.Value(constants.UserRoleContextKey) != nil),
		zap.Bool("hasClientIP", ctx.Value(constants.IPContextKey) != nil),
	)

	// Try to extract and log user details if present
	if userVal := ctx.Value(constants.UserContextKey); userVal != nil {
		if user, ok := userVal.(*models.User); ok && user != nil {
			r.logger.Info("SendVerificationEmail: User found in context!",
				zap.Uint("userID", user.ID),
				zap.String("username", user.Username),
				zap.String("role", string(user.Role)),
				zap.Bool("emailVerified", user.EmailVerified),
				zap.Bool("hasEmail", user.Email != ""),
			)
		} else {
			r.logger.Error("SendVerificationEmail: User context exists but wrong type",
				zap.String("actualType", fmt.Sprintf("%T", userVal)),
			)
		}
	} else {
		r.logger.Error("SendVerificationEmail: User context is completely nil")
	}

	// Get current user from context
	user, ok := ctx.Value(constants.UserContextKey).(*models.User)
	if !ok || user == nil {
		r.logger.Warn("SendVerificationEmail: User not found in context",
			zap.Bool("contextHasUser", ctx.Value(constants.UserContextKey) != nil),
			zap.String("contextType", fmt.Sprintf("%T", ctx.Value(constants.UserContextKey))),
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
	// Find user by email
	user, err := r.userService.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		msg := "If an account exists with this email, a password reset link will be sent"
		//nolint:nilerr // Intentionally returning nil error for security reasons
		return &model.MutationResponse{
			Success: true,
			Message: &msg,
		}, nil
	}

	// Send password reset email
	if err := r.emailVerificationService.SendPasswordResetEmailToUser(ctx, user); err != nil {
		// Log error but return success for security
		r.logger.Error("Failed to send password reset email", zap.String("email", email), zap.Error(err))
	}

	msg := "If an account exists with this email, a password reset link will be sent"
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
