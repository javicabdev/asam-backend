package input

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// UserService defines user management operations
type UserService interface {
	// CreateUser creates a new user with the given details
	CreateUser(ctx context.Context, username, password string, role models.UserRole) (*models.User, error)

	// UpdateUser updates an existing user's details
	UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error)

	// DeleteUser deletes a user by ID (soft delete)
	DeleteUser(ctx context.Context, id uint) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id uint) (*models.User, error)

	// ListUsers retrieves a paginated list of users
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error

	// ResetPassword resets a user's password (admin function)
	ResetPassword(ctx context.Context, userID uint, newPassword string) error
}
