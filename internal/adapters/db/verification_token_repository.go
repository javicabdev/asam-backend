package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// verificationTokenRepository implementation using verification_tokens table
type verificationTokenRepository struct {
	db *gorm.DB
}

// NewVerificationTokenRepository creates a new verification token repository
func NewVerificationTokenRepository(db *gorm.DB) output.VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

// Create creates a new verification token
func (r *verificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		return appErrors.DB(result.Error, "Error creating verification token")
	}
	return nil
}

// FindByToken finds a token by its value
func (r *verificationTokenRepository) FindByToken(ctx context.Context, tokenValue string) (*models.VerificationToken, error) {
	var token models.VerificationToken
	result := r.db.WithContext(ctx).
		Where("token = ?", tokenValue).
		First(&token)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil without error when not found
		}
		return nil, appErrors.DB(result.Error, "Error finding verification token")
	}

	return &token, nil
}

// Update updates an existing token
func (r *verificationTokenRepository) Update(ctx context.Context, token *models.VerificationToken) error {
	result := r.db.WithContext(ctx).Save(token)
	if result.Error != nil {
		return appErrors.DB(result.Error, "Error updating verification token")
	}
	return nil
}

// DeleteExpiredTokens removes expired tokens from the database
func (r *verificationTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&models.VerificationToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deleting expired tokens")
	}

	return nil
}

// DeleteUserTokensByType deletes all tokens of a specific type for a user
func (r *verificationTokenRepository) DeleteUserTokensByType(ctx context.Context, userID uint, tokenType models.TokenType) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, tokenType).
		Delete(&models.VerificationToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deleting user tokens by type")
	}

	return nil
}

// CountActiveTokensByUser counts active tokens for a user of a specific type
func (r *verificationTokenRepository) CountActiveTokensByUser(ctx context.Context, userID uint, tokenType models.TokenType) (int64, error) {
	var count int64
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.VerificationToken{}).
		Where("user_id = ? AND type = ? AND used = ? AND expires_at > ?", userID, tokenType, false, now).
		Count(&count)

	if result.Error != nil {
		return 0, appErrors.DB(result.Error, "Error counting active tokens")
	}

	return count, nil
}
