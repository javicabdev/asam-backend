// Package output defines the output port interfaces for the ASAM backend.
// These interfaces are implemented by adapters to provide persistence capabilities.
package output

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// UserRepository define las operaciones para persistir usuarios
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// TokenRepository handles refresh token operations
type TokenRepository interface {
	// Basic operations
	SaveRefreshToken(ctx context.Context, uuid string, userID uint, expires int64) error
	ValidateRefreshToken(ctx context.Context, uuid string, userID uint) error
	DeleteRefreshToken(ctx context.Context, uuid string) error

	// Session management
	DeleteAllUserTokens(ctx context.Context, userID uint) error
	GetUserActiveSessions(ctx context.Context, userID uint) ([]*models.RefreshToken, error)

	// Maintenance
	CleanupExpiredTokens(ctx context.Context) error
	EnforceTokenLimitPerUser(ctx context.Context, maxTokens int) error
}
