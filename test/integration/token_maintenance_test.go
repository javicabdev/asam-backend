//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/test/fixtures"
	"github.com/javicabdev/asam-backend/test/helpers"
)

func TestTokenRepository_TokenMaintenance(t *testing.T) {
	database := helpers.SetupTestDB(t)
	tokenRepo := db.NewTokenRepository(database)
	userRepo := db.NewUserRepository(database)

	// Create test user
	user := fixtures.CreateTestUser(t, userRepo, "test@example.com", "password123")

	ctx := context.Background()

	t.Run("CleanupExpiredTokens", func(t *testing.T) {
		// Create expired token
		expiredToken := &models.RefreshToken{
			UUID:       "expired-token-uuid",
			UserID:     user.ID,
			ExpiresAt:  time.Now().Add(-24 * time.Hour).Unix(), // Expired 24 hours ago
			CreatedAt:  time.Now().Add(-48 * time.Hour),
			LastUsedAt: time.Now().Add(-24 * time.Hour),
		}
		err := database.Create(expiredToken).Error
		require.NoError(t, err)

		// Create active token
		activeToken := &models.RefreshToken{
			UUID:       "active-token-uuid",
			UserID:     user.ID,
			ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(), // Expires in 24 hours
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
		}
		err = database.Create(activeToken).Error
		require.NoError(t, err)

		// Run cleanup
		err = tokenRepo.CleanupExpiredTokens(ctx)
		assert.NoError(t, err)

		// Verify expired token is deleted
		var count int64
		database.Model(&models.RefreshToken{}).Where("uuid = ?", "expired-token-uuid").Count(&count)
		assert.Equal(t, int64(0), count, "Expired token should be deleted")

		// Verify active token still exists
		database.Model(&models.RefreshToken{}).Where("uuid = ?", "active-token-uuid").Count(&count)
		assert.Equal(t, int64(1), count, "Active token should still exist")

		// Cleanup
		database.Where("uuid = ?", "active-token-uuid").Delete(&models.RefreshToken{})
	})

	t.Run("EnforceTokenLimitPerUser", func(t *testing.T) {
		// Create multiple tokens for the same user
		for i := 0; i < 10; i++ {
			token := &models.RefreshToken{
				UUID:       fmt.Sprintf("token-%d", i),
				UserID:     user.ID,
				ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(),
				CreatedAt:  time.Now().Add(time.Duration(-i) * time.Hour), // Older tokens have earlier creation time
				LastUsedAt: time.Now(),
			}
			err := database.Create(token).Error
			require.NoError(t, err)
		}

		// Enforce limit of 5 tokens
		err := tokenRepo.EnforceTokenLimitPerUser(ctx, 5)
		assert.NoError(t, err)

		// Count remaining tokens
		var count int64
		database.Model(&models.RefreshToken{}).Where("user_id = ?", user.ID).Count(&count)
		assert.Equal(t, int64(5), count, "Should have exactly 5 tokens after enforcement")

		// Verify that the newest tokens are kept
		var remainingTokens []models.RefreshToken
		database.Where("user_id = ?", user.ID).Order("created_at DESC").Find(&remainingTokens)

		assert.Len(t, remainingTokens, 5)
		for i, token := range remainingTokens {
			assert.Equal(t, fmt.Sprintf("token-%d", i), token.UUID, "Newest tokens should be kept")
		}

		// Cleanup
		database.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})
	})

	t.Run("SaveRefreshToken_WithContextInfo", func(t *testing.T) {
		// Create context with client information
		ctx := context.Background()
		ctx = context.WithValue(ctx, constants.DeviceNameContextKey, "iPhone 15")
		ctx = context.WithValue(ctx, constants.IPAddressContextKey, "192.168.1.100")
		ctx = context.WithValue(ctx, constants.UserAgentContextKey, "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)")

		// Save token
		err := tokenRepo.SaveRefreshToken(ctx, "context-test-uuid", user.ID, time.Now().Add(24*time.Hour).Unix())
		assert.NoError(t, err)

		// Retrieve and verify
		var savedToken models.RefreshToken
		err = database.Where("uuid = ?", "context-test-uuid").First(&savedToken).Error
		require.NoError(t, err)

		assert.Equal(t, "iPhone 15", savedToken.DeviceName)
		assert.Equal(t, "192.168.1.100", savedToken.IPAddress)
		assert.Equal(t, "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X)", savedToken.UserAgent)

		// Cleanup
		database.Where("uuid = ?", "context-test-uuid").Delete(&models.RefreshToken{})
	})

	t.Run("GetUserActiveSessions", func(t *testing.T) {
		// Create mixed tokens (active and expired)
		tokens := []models.RefreshToken{
			{
				UUID:       "active-1",
				UserID:     user.ID,
				ExpiresAt:  time.Now().Add(24 * time.Hour).Unix(),
				CreatedAt:  time.Now(),
				DeviceName: "Windows Desktop",
				IPAddress:  "10.0.0.1",
			},
			{
				UUID:       "active-2",
				UserID:     user.ID,
				ExpiresAt:  time.Now().Add(48 * time.Hour).Unix(),
				CreatedAt:  time.Now(),
				DeviceName: "Android Mobile",
				IPAddress:  "10.0.0.2",
			},
			{
				UUID:       "expired-1",
				UserID:     user.ID,
				ExpiresAt:  time.Now().Add(-24 * time.Hour).Unix(),
				CreatedAt:  time.Now().Add(-48 * time.Hour),
				DeviceName: "Old Device",
				IPAddress:  "10.0.0.3",
			},
		}

		for _, token := range tokens {
			err := database.Create(&token).Error
			require.NoError(t, err)
		}

		// Get active sessions
		sessions, err := tokenRepo.GetUserActiveSessions(ctx, user.ID)
		assert.NoError(t, err)
		assert.Len(t, sessions, 2, "Should return only active sessions")

		// Verify sessions are ordered by creation date (newest first)
		if len(sessions) >= 2 {
			assert.True(t, sessions[0].CreatedAt.After(sessions[1].CreatedAt) || sessions[0].CreatedAt.Equal(sessions[1].CreatedAt))
		}

		// Cleanup
		database.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{})
	})
}
