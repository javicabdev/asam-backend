package output

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// internal/ports/output/auth_repository.go
type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, uuid string, userId uint, expires int64) error
	ValidateRefreshToken(ctx context.Context, uuid string, userId uint) error
	DeleteRefreshToken(ctx context.Context, uuid string) error
	CleanupExpiredTokens(ctx context.Context) error // Añadimos este método
}
