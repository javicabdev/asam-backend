package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// Helper function to create user service with all dependencies
func createUserServiceWithMocks() (input.UserService, *MockUserRepository, *MockVerificationTokenRepository, *MockEmailService) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockVerificationTokenRepository)
	emailService := new(MockEmailService)
	logger := &test.MockLogger{}
	baseURL := "http://test.example.com"

	userService := services.NewUserService(userRepo, tokenRepo, emailService, logger, baseURL)

	return userService, userRepo, tokenRepo, emailService
}

// TestSendVerificationEmail tests
func TestUserService_SendVerificationEmail_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uint(1)
	email := "test@example.com"

	_, userRepo, tokenRepo, emailService := createUserServiceWithMocks()

	user := &models.User{
		Model:         gorm.Model{ID: userID},
		Username:      email,
		EmailVerified: false,
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	tokenRepo.On("DeleteUserTokensByType", ctx, userID, models.TokenTypeEmailVerification).Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil)
	emailService.On("SendVerificationEmail", ctx, email, "test", mock.AnythingOfType("string")).Return(nil)

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, emailService, &test.MockLogger{}, "http://test.example.com")
	err := userService.SendVerificationEmail(ctx, userID)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_SendVerificationEmail_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uint(999)

	_, userRepo, _, _ := createUserServiceWithMocks()

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	userService := services.NewUserService(userRepo, new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.SendVerificationEmail(ctx, userID)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	userRepo.AssertExpectations(t)
}

func TestUserService_SendVerificationEmail_NotEmailUsername(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uint(1)

	_, userRepo, _, _ := createUserServiceWithMocks()

	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: "regularusername",
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act
	userService := services.NewUserService(userRepo, new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.SendVerificationEmail(ctx, userID)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidRequest, appErr.Code)
	assert.Contains(t, appErr.Message, "does not have an email username")

	userRepo.AssertExpectations(t)
}

func TestUserService_SendVerificationEmail_AlreadyVerified(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := uint(1)

	_, userRepo, _, _ := createUserServiceWithMocks()

	user := &models.User{
		Model:         gorm.Model{ID: userID},
		Username:      "test@example.com",
		EmailVerified: true,
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act
	userService := services.NewUserService(userRepo, new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.SendVerificationEmail(ctx, userID)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidRequest, appErr.Code)
	assert.Contains(t, appErr.Message, "already verified")

	userRepo.AssertExpectations(t)
}

// TestVerifyEmail tests
func TestUserService_VerifyEmail_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "valid-token"
	userID := uint(1)
	email := "test@example.com"

	_, userRepo, tokenRepo, _ := createUserServiceWithMocks()

	user := &models.User{
		Model:         gorm.Model{ID: userID},
		Username:      email,
		EmailVerified: false,
	}

	verificationToken := &models.VerificationToken{
		Model:     gorm.Model{ID: 1},
		Token:     tokenValue,
		UserID:    userID,
		Type:      models.TokenTypeEmailVerification,
		Email:     email,
		Used:      false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Setup mocks
	tokenRepo.On("FindByToken", ctx, tokenValue).Return(verificationToken, nil)
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify user is marked as verified
		updatedUser := args.Get(1).(*models.User)
		assert.True(t, updatedUser.EmailVerified)
		assert.NotNil(t, updatedUser.EmailVerifiedAt)
	})
	tokenRepo.On("Update", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil).Run(func(args mock.Arguments) {
		// Verify token is marked as used
		updatedToken := args.Get(1).(*models.VerificationToken)
		assert.True(t, updatedToken.Used)
		assert.NotNil(t, updatedToken.UsedAt)
	})

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.VerifyEmail(ctx, tokenValue)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestUserService_VerifyEmail_InvalidToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "invalid-token"

	_, _, tokenRepo, _ := createUserServiceWithMocks()

	// Setup mocks
	tokenRepo.On("FindByToken", ctx, tokenValue).Return(nil, nil)

	// Act
	userService := services.NewUserService(new(MockUserRepository), tokenRepo, new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.VerifyEmail(ctx, tokenValue)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	tokenRepo.AssertExpectations(t)
}

func TestUserService_VerifyEmail_ExpiredToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "expired-token"
	userID := uint(1)

	_, _, tokenRepo, _ := createUserServiceWithMocks()

	verificationToken := &models.VerificationToken{
		Model:     gorm.Model{ID: 1},
		Token:     tokenValue,
		UserID:    userID,
		Type:      models.TokenTypeEmailVerification,
		Used:      false,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
	}

	// Setup mocks
	tokenRepo.On("FindByToken", ctx, tokenValue).Return(verificationToken, nil)

	// Act
	userService := services.NewUserService(new(MockUserRepository), tokenRepo, new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.VerifyEmail(ctx, tokenValue)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidRequest, appErr.Code)
	assert.Contains(t, appErr.Message, "expired")

	tokenRepo.AssertExpectations(t)
}

// TestRequestPasswordReset tests
func TestUserService_RequestPasswordReset_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	userID := uint(1)

	_, userRepo, tokenRepo, emailService := createUserServiceWithMocks()

	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: email,
		IsActive: true,
	}

	// Setup mocks
	userRepo.On("FindByUsername", ctx, email).Return(user, nil)
	tokenRepo.On("CountActiveTokensByUser", ctx, userID, models.TokenTypePasswordReset).Return(int64(0), nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil)
	emailService.On("SendPasswordResetEmail", ctx, email, "test", mock.AnythingOfType("string")).Return(nil)

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, emailService, &test.MockLogger{}, "http://test.example.com")
	err := userService.RequestPasswordReset(ctx, email)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_RequestPasswordReset_NonExistentEmail(t *testing.T) {
	// Arrange
	ctx := context.Background()
	email := "nonexistent@example.com"

	_, userRepo, _, _ := createUserServiceWithMocks()

	// Setup mocks
	userRepo.On("FindByUsername", ctx, email).Return(nil, nil)

	// Act
	userService := services.NewUserService(userRepo, new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.RequestPasswordReset(ctx, email)

	// Assert
	// Should not reveal that email doesn't exist
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestUserService_RequestPasswordReset_RateLimitExceeded(t *testing.T) {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	userID := uint(1)

	_, userRepo, tokenRepo, _ := createUserServiceWithMocks()

	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: email,
		IsActive: true,
	}

	// Setup mocks
	userRepo.On("FindByUsername", ctx, email).Return(user, nil)
	tokenRepo.On("CountActiveTokensByUser", ctx, userID, models.TokenTypePasswordReset).Return(int64(3), nil) // Max reached

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.RequestPasswordReset(ctx, email)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrRateLimitExceeded, appErr.Code)

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// TestResetPasswordWithToken tests
func TestUserService_ResetPasswordWithToken_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "valid-reset-token"
	newPassword := "NewSecurePass123!"
	userID := uint(1)
	email := "test@example.com"

	_, userRepo, tokenRepo, emailService := createUserServiceWithMocks()

	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: email,
		Password: "old-hashed-password",
	}

	resetToken := &models.VerificationToken{
		Model:     gorm.Model{ID: 1},
		Token:     tokenValue,
		UserID:    userID,
		Type:      models.TokenTypePasswordReset,
		Email:     email,
		Used:      false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	// Setup mocks
	tokenRepo.On("FindByToken", ctx, tokenValue).Return(resetToken, nil)
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify password is updated
		updatedUser := args.Get(1).(*models.User)
		assert.NotEqual(t, "old-hashed-password", updatedUser.Password)
	})
	tokenRepo.On("Update", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil)
	tokenRepo.On("DeleteUserTokensByType", ctx, userID, models.TokenTypePasswordReset).Return(nil)
	emailService.On("SendPasswordChangedEmail", ctx, email, "test").Return(nil)

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, emailService, &test.MockLogger{}, "http://test.example.com")
	err := userService.ResetPasswordWithToken(ctx, tokenValue, newPassword)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_ResetPasswordWithToken_InvalidToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "invalid-token"
	newPassword := "NewSecurePass123!"

	_, _, tokenRepo, _ := createUserServiceWithMocks()

	// Setup mocks
	tokenRepo.On("FindByToken", ctx, tokenValue).Return(nil, nil)

	// Act
	userService := services.NewUserService(new(MockUserRepository), tokenRepo, new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.ResetPasswordWithToken(ctx, tokenValue, newPassword)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	tokenRepo.AssertExpectations(t)
}

func TestUserService_ResetPasswordWithToken_InvalidPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	tokenValue := "valid-token"
	newPassword := "weak" // Too short

	// Act - No mocks needed since password validation happens first
	userService := services.NewUserService(new(MockUserRepository), new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.ResetPasswordWithToken(ctx, tokenValue, newPassword)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
	assert.Contains(t, appErr.Fields["password"], "Password must be at least 8 characters long")
}

// TestResendVerificationEmail tests
func TestUserService_ResendVerificationEmail_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	userID := uint(1)

	_, userRepo, tokenRepo, emailService := createUserServiceWithMocks()

	user := &models.User{
		Model:         gorm.Model{ID: userID},
		Username:      email,
		EmailVerified: false,
	}

	// Setup mocks for finding user
	userRepo.On("FindByUsername", ctx, email).Return(user, nil)

	// Setup mocks for SendVerificationEmail
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	tokenRepo.On("DeleteUserTokensByType", ctx, userID, models.TokenTypeEmailVerification).Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil)
	emailService.On("SendVerificationEmail", ctx, email, "test", mock.AnythingOfType("string")).Return(nil)

	// Act
	userService := services.NewUserService(userRepo, tokenRepo, emailService, &test.MockLogger{}, "http://test.example.com")
	err := userService.ResendVerificationEmail(ctx, email)

	// Assert
	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_ResendVerificationEmail_AlreadyVerified(t *testing.T) {
	// Arrange
	ctx := context.Background()
	email := "test@example.com"
	userID := uint(1)

	_, userRepo, _, _ := createUserServiceWithMocks()

	user := &models.User{
		Model:         gorm.Model{ID: userID},
		Username:      email,
		EmailVerified: true, // Already verified
	}

	// Setup mocks
	userRepo.On("FindByUsername", ctx, email).Return(user, nil)

	// Act
	userService := services.NewUserService(userRepo, new(MockVerificationTokenRepository), new(MockEmailService), &test.MockLogger{}, "http://test.example.com")
	err := userService.ResendVerificationEmail(ctx, email)

	// Assert
	assert.Error(t, err)
	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidRequest, appErr.Code)
	assert.Contains(t, appErr.Message, "already verified")

	userRepo.AssertExpectations(t)
}
