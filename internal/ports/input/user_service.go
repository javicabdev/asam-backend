package input

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// UserService defines user management operations
type UserService interface {
	// CreateUser creates a new user with the given details
	CreateUser(ctx context.Context, username, password string, role models.Role) (*models.User, error)

	// UpdateUser updates an existing user's details
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error)

	// DeleteUser deletes a user by ID (soft delete)
	DeleteUser(ctx context.Context, id uint) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id uint) (*models.User, error)

	// GetUserByEmail retrieves a user by email address
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// ListUsers retrieves a paginated list of users
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error

	// ResetPassword resets a user's password (admin function)
	ResetPassword(ctx context.Context, userID uint, newPassword string) error

	// SendVerificationEmail sends a verification email to the user
	SendVerificationEmail(ctx context.Context, userID uint) error

	// VerifyEmail verifies a user's email with the provided token
	VerifyEmail(ctx context.Context, token string) error

	// RequestPasswordReset initiates a password reset for the given email
	RequestPasswordReset(ctx context.Context, email string) error

	// ResetPasswordWithToken resets password using a valid token
	ResetPasswordWithToken(ctx context.Context, token, newPassword string) error

	// ResendVerificationEmail resends the verification email
	ResendVerificationEmail(ctx context.Context, email string) error
}
