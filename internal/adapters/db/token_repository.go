package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// tokenRepository implementation using refresh_tokens table
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) output.TokenRepository {
	return &tokenRepository{db: db}
}

// SaveRefreshToken stores a refresh token in the refresh_tokens table
func (r *tokenRepository) SaveRefreshToken(ctx context.Context, uuid string, userID uint, expires int64) error {
	now := time.Now()
	token := &models.RefreshToken{
		UUID:       uuid,
		UserID:     userID,
		ExpiresAt:  expires,
		CreatedAt:  now,
		LastUsedAt: now, // Initialize with current time since it's being used now
	}

	// Extract additional context if available
	if deviceName, ok := ctx.Value(constants.DeviceNameContextKey).(string); ok {
		token.DeviceName = deviceName
	}
	if ipAddress, ok := ctx.Value(constants.IPAddressContextKey).(string); ok {
		token.IPAddress = ipAddress
	}
	if userAgent, ok := ctx.Value(constants.UserAgentContextKey).(string); ok {
		token.UserAgent = userAgent
	}

	result := r.db.WithContext(ctx).Create(token)

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error saving refresh token")
	}

	return nil
}

// ValidateRefreshToken checks if a token exists and is valid
func (r *tokenRepository) ValidateRefreshToken(ctx context.Context, uuid string, userID uint) error {
	var token models.RefreshToken
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
	if token.IsExpired() {
		return appErrors.New(appErrors.ErrInvalidToken, "Token has expired")
	}

	// Update last used timestamp
	r.db.WithContext(ctx).
		Model(&token).
		Update("last_used_at", time.Now())

	return nil
}

// DeleteRefreshToken removes a specific token
func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deleting refresh token")
	}

	if result.RowsAffected == 0 {
		return appErrors.New(appErrors.ErrNotFound, "Token not found for deletion")
	}

	return nil
}

// DeleteAllUserTokens removes all tokens for a specific user
func (r *tokenRepository) DeleteAllUserTokens(ctx context.Context, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error deleting user tokens")
	}

	return nil
}

// GetUserActiveSessions returns all active sessions for a user
func (r *tokenRepository) GetUserActiveSessions(ctx context.Context, userID uint) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	now := time.Now().Unix()

	result := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&tokens)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "Error fetching user sessions")
	}

	return tokens, nil
}

// CleanupExpiredTokens removes expired tokens from the database
func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	now := time.Now().Unix()

	result := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		return appErrors.DB(result.Error, "Error cleaning up expired tokens")
	}

	return nil
}

// EnforceTokenLimitPerUser removes oldest tokens if user has more than maxTokens
func (r *tokenRepository) EnforceTokenLimitPerUser(ctx context.Context, maxTokens int) error {
	// Get all users with their token counts
	type userTokenCount struct {
		UserID uint
		Count  int64
	}

	var userCounts []userTokenCount
	if err := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Select("user_id, COUNT(*) as count").
		Group("user_id").
		Having("COUNT(*) > ?", maxTokens).
		Scan(&userCounts).Error; err != nil {
		return appErrors.DB(err, "Error getting user token counts")
	}

	// For each user with too many tokens
	for _, uc := range userCounts {
		// Calculate how many tokens to delete
		tokensToDelete := int(uc.Count) - maxTokens

		// Get the oldest tokens for this user
		var oldestTokens []models.RefreshToken
		if err := r.db.WithContext(ctx).
			Where("user_id = ?", uc.UserID).
			Order("created_at ASC").
			Limit(tokensToDelete).
			Find(&oldestTokens).Error; err != nil {
			return appErrors.DB(err, "Error getting oldest tokens")
		}

		// Delete the oldest tokens
		for _, token := range oldestTokens {
			if err := r.db.WithContext(ctx).
				Delete(&token).Error; err != nil {
				return appErrors.DB(err, "Error deleting old token")
			}
		}
	}

	return nil
}
