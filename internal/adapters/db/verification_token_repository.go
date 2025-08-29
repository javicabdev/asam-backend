package db

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// verificationTokenRepository implements the VerificationTokenRepository interface using GORM
type verificationTokenRepository struct {
	db *gorm.DB
}

// NewVerificationTokenRepository creates a new instance of verification token repository
func NewVerificationTokenRepository(db *gorm.DB) output.VerificationTokenRepository {
	return &verificationTokenRepository{db: db}
}

// Create creates a new verification token
func (r *verificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "failed to create verification token")
	}
	return nil
}

// GetByToken retrieves a token by its value
func (r *verificationTokenRepository) GetByToken(ctx context.Context, token string) (*models.VerificationToken, error) {
	var verificationToken models.VerificationToken
	err := r.db.WithContext(ctx).
		Where("token = ?", token).
		First(&verificationToken).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewNotFoundError("verification token")
		}
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "failed to get verification token")
	}

	return &verificationToken, nil
}

// GetByUserIDAndType retrieves tokens by user ID and type
func (r *verificationTokenRepository) GetByUserIDAndType(ctx context.Context, userID uint, tokenType string) ([]*models.VerificationToken, error) {
	var tokens []*models.VerificationToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND type = ?", userID, tokenType).
		Find(&tokens).Error

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "failed to get verification tokens")
	}

	return tokens, nil
}

// Update updates a verification token
func (r *verificationTokenRepository) Update(ctx context.Context, token *models.VerificationToken) error {
	token.UpdatedAt = time.Now()

	// Use Updates with a map to ensure boolean fields are properly updated
	updates := map[string]interface{}{
		"token":      token.Token,
		"user_id":    token.UserID,
		"type":       token.Type,
		"email":      token.Email,
		"used":       token.Used,
		"used_at":    token.UsedAt,
		"expires_at": token.ExpiresAt,
		"updated_at": token.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).Model(token).Updates(updates).Error; err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "failed to update verification token")
	}
	return nil
}

// Delete deletes a verification token
func (r *verificationTokenRepository) Delete(ctx context.Context, tokenID uint) error {
	result := r.db.WithContext(ctx).Delete(&models.VerificationToken{}, tokenID)
	if result.Error != nil {
		return errors.Wrap(result.Error, errors.ErrDatabaseError, "failed to delete verification token")
	}
	if result.RowsAffected == 0 {
		return errors.NewNotFoundError("verification token")
	}
	return nil
}

// DeleteExpired deletes all expired tokens
func (r *verificationTokenRepository) DeleteExpired(ctx context.Context) error {
	err := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.VerificationToken{}).Error

	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "failed to delete expired tokens")
	}

	return nil
}

// InvalidateUserTokens invalidates all tokens of a specific type for a user
func (r *verificationTokenRepository) InvalidateUserTokens(ctx context.Context, userID uint, tokenType string) error {
	now := time.Now()

	// Update both 'used' and 'used_at' fields for consistency
	err := r.db.WithContext(ctx).
		Model(&models.VerificationToken{}).
		Where("user_id = ? AND type = ? AND used = false", userID, tokenType).
		Updates(map[string]interface{}{
			"used":       true,
			"used_at":    &now,
			"updated_at": now,
		}).Error

	if err != nil {
		return errors.Wrap(err, errors.ErrDatabaseError, "failed to invalidate user tokens")
	}

	return nil
}

// CountActiveTokensByUser counts the number of active (non-expired, non-used) tokens for a user and type
func (r *verificationTokenRepository) CountActiveTokensByUser(ctx context.Context, userID uint, tokenType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.VerificationToken{}).
		Where("user_id = ? AND type = ? AND used = false AND expires_at > ?", userID, tokenType, time.Now()).
		Count(&count).Error

	if err != nil {
		return 0, errors.Wrap(err, errors.ErrDatabaseError, "failed to count active tokens")
	}

	return count, nil
}

// MarkTokenAsUsed marks a specific token as used atomically
func (r *verificationTokenRepository) MarkTokenAsUsed(ctx context.Context, tokenID uint) error {
	now := time.Now()

	// Atomic update to mark token as used
	result := r.db.WithContext(ctx).
		Model(&models.VerificationToken{}).
		Where("id = ? AND used = false", tokenID).
		Updates(map[string]interface{}{
			"used":       true,
			"used_at":    &now,
			"updated_at": now,
		})

	if result.Error != nil {
		return errors.Wrap(result.Error, errors.ErrDatabaseError, "failed to mark token as used")
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		return errors.NewBusinessError(errors.ErrInvalidToken, "token already used or not found")
	}

	return nil
}
