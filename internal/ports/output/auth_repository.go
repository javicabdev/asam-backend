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
	Update(ctx context.Context, user *models.User) error
}

// TokenRepository internal/ports/output/auth_repository.go
type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, uuid string, userID uint, expires int64) error
	ValidateRefreshToken(ctx context.Context, uuid string, userID uint) error
	DeleteRefreshToken(ctx context.Context, uuid string) error
	CleanupExpiredTokens(ctx context.Context) error // Añadimos este método
}
