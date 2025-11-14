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
	FindByMemberID(ctx context.Context, memberID uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, userID uint) error

	// Additional operations
	IsUserActive(ctx context.Context, userID uint) (bool, error)
	DeactivateUser(ctx context.Context, userID uint) error
	ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, int64, error)
	GetUserWithMember(ctx context.Context, userID uint) (*models.User, error)
	CountUsersByRole(ctx context.Context, role models.Role) (int64, error)
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

	// Sliding expiration support
	GetRefreshToken(ctx context.Context, uuid string) (*models.RefreshToken, error)
	ExtendTokenExpiration(ctx context.Context, uuid string, newExpires int64) error

	// Maintenance
	CleanupExpiredTokens(ctx context.Context) error
	EnforceTokenLimitPerUser(ctx context.Context, maxTokens int) error
}
