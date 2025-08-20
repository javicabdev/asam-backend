package output

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// VerificationTokenRepository defines the interface for verification token persistence
type VerificationTokenRepository interface {
	// Create creates a new verification token
	Create(ctx context.Context, token *models.VerificationToken) error

	// GetByToken retrieves a token by its value
	GetByToken(ctx context.Context, token string) (*models.VerificationToken, error)

	// GetByUserIDAndType retrieves tokens by user ID and type
	GetByUserIDAndType(ctx context.Context, userID uint, tokenType string) ([]*models.VerificationToken, error)

	// Update updates a verification token
	Update(ctx context.Context, token *models.VerificationToken) error

	// Delete deletes a verification token
	Delete(ctx context.Context, tokenID uint) error

	// DeleteExpired deletes all expired tokens
	DeleteExpired(ctx context.Context) error

	// InvalidateUserTokens invalidates all tokens of a specific type for a user
	InvalidateUserTokens(ctx context.Context, userID uint, tokenType string) error

	// CountActiveTokensByUser counts the number of active (non-expired, non-used) tokens for a user and type
	CountActiveTokensByUser(ctx context.Context, userID uint, tokenType string) (int64, error)
}
