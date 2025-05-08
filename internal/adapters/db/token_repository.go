package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// RefreshToken model for storing refresh tokens
type RefreshToken struct {
	UUID      string `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	ExpiresAt int64  `gorm:"not null"`
	CreatedAt time.Time
}

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) output.TokenRepository {
	return &tokenRepository{db: db}
}

// SaveRefreshToken stores a refresh token in the database
func (r *tokenRepository) SaveRefreshToken(ctx context.Context, uuid string, userID uint, expires int64) error {
	token := RefreshToken{
		UUID:      uuid,
		UserID:    userID,
		ExpiresAt: expires,
	}

	result := r.db.WithContext(ctx).Create(&token)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "Token UUID already exists")
		}
		return appErrors.DB(result.Error, "Error saving refresh token")
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	now := time.Now().Unix()

	result := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&RefreshToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error cleaning up expired tokens")
	}

	return nil
}

// ValidateRefreshToken checks if a token exists and is valid
func (r *tokenRepository) ValidateRefreshToken(ctx context.Context, uuid string, userID uint) error {
	// First clean up expired tokens
	if err := r.CleanupExpiredTokens(ctx); err != nil {
		// Log error but continue - not critical
		// Instead of directly returning, we wrap this as a low-severity error
		return appErrors.Wrap(err, appErrors.ErrInternalError, "Warning: Failed to clean up expired tokens")
	}

	// Then validate the specific token
	var token RefreshToken
	result := r.db.WithContext(ctx).
		Where("uuid = ? AND user_id = ?", uuid, userID).
		First(&token)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.New(appErrors.ErrInvalidOperation, "Token not found or does not match user")
		}
		return appErrors.DB(result.Error, "Error validating refresh token")
	}

	// Check for expiration
	if time.Now().Unix() > token.ExpiresAt {
		return appErrors.New(appErrors.ErrInvalidToken, "Token has expired")
	}

	return nil
}

// DeleteRefreshToken removes a token from the database
func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		Delete(&RefreshToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deleting refresh token")
	}

	// Check if any token was actually deleted
	if result.RowsAffected == 0 {
		return appErrors.New(appErrors.ErrNotFound, "Token not found for deletion")
	}

	return nil
}
