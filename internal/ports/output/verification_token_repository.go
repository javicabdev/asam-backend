package output

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// VerificationTokenRepository define las operaciones de persistencia para tokens de verificación
type VerificationTokenRepository interface {
	// Create crea un nuevo token de verificación
	Create(ctx context.Context, token *models.VerificationToken) error

	// FindByToken busca un token por su valor
	FindByToken(ctx context.Context, token string) (*models.VerificationToken, error)

	// Update actualiza un token existente
	Update(ctx context.Context, token *models.VerificationToken) error

	// DeleteExpiredTokens elimina tokens expirados
	DeleteExpiredTokens(ctx context.Context) error

	// DeleteUserTokensByType elimina todos los tokens de un usuario de un tipo específico
	DeleteUserTokensByType(ctx context.Context, userID uint, tokenType models.TokenType) error

	// CountActiveTokensByUser cuenta los tokens activos de un usuario
	CountActiveTokensByUser(ctx context.Context, userID uint, tokenType models.TokenType) (int64, error)
}
