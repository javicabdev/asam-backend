package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// TestUserService_CreateUser tests
func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	username := "newuser"
	password := "SecurePass123!"
	role := models.RoleUser

	// Setup mocks - user doesn't exist
	userRepo.On("FindByUsername", ctx, username).Return(nil, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify the user object passed to Create
		user := args.Get(1).(*models.User)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, role, user.Role)
		assert.True(t, user.IsActive)

		// Verify password is hashed
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		assert.NoError(t, err)

		// Set ID to simulate DB behavior
		user.ID = 1
	})

	// Act
	result, err := userService.CreateUser(ctx, username, password, role)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, role, result.Role)
	assert.True(t, result.IsActive)
	assert.Empty(t, result.Password) // Password should be cleared

	userRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_InvalidUsername(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()

	testCases := []struct {
		name     string
		username string
		errMsg   string
	}{
		{
			name:     "Empty username",
			username: "",
			errMsg:   "Username cannot be empty",
		},
		{
			name:     "Too short username",
			username: "ab",
			errMsg:   "Username must be at least 3 characters long",
		},
		{
			name:     "Too long username",
			username: "thisisaverylongusernamethatexceedsfiftycharactersandshouldfail",
			errMsg:   "Username must not exceed 50 characters",
		},
		{
			name:     "Invalid characters",
			username: "user@#$%",
			errMsg:   "Username can only contain letters, numbers, underscore, hyphen, and dot",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := userService.CreateUser(ctx, tc.username, "ValidPass123!", models.RoleUser)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)

			appErr, ok := appErrors.AsAppError(err)
			require.True(t, ok)
			assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
			assert.Contains(t, appErr.Fields["username"], tc.errMsg)
		})
	}
}

func TestUserService_CreateUser_InvalidPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	username := "validuser"

	testCases := []struct {
		name     string
		password string
		errMsg   string
	}{
		{
			name:     "Empty password",
			password: "",
			errMsg:   "Password cannot be empty",
		},
		{
			name:     "Too short password",
			password: "Pass1!",
			errMsg:   "Password must be at least 8 characters long",
		},
		{
			name:     "Too long password",
			password: "ThisIsAnExtremelyLongPasswordThatExceedsOneHundredCharactersAndShouldFailValidationBecauseItIsTooLongForOurSystem",
			errMsg:   "Password must not exceed 100 characters",
		},
		{
			name:     "No uppercase",
			password: "password123!",
			errMsg:   "Password must contain at least one uppercase letter, one lowercase letter, and one number",
		},
		{
			name:     "No lowercase",
			password: "PASSWORD123!",
			errMsg:   "Password must contain at least one uppercase letter, one lowercase letter, and one number",
		},
		{
			name:     "No number",
			password: "Password!",
			errMsg:   "Password must contain at least one uppercase letter, one lowercase letter, and one number",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := userService.CreateUser(ctx, username, tc.password, models.RoleUser)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)

			appErr, ok := appErrors.AsAppError(err)
			require.True(t, ok)
			assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
			assert.Contains(t, appErr.Fields["password"], tc.errMsg)
		})
	}
}

func TestUserService_CreateUser_UsernameAlreadyExists(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	username := "existinguser"
	password := "ValidPass123!"
	role := models.RoleUser

	existingUser := createTestUser(1, username, true)

	// Setup mocks - user already exists
	userRepo.On("FindByUsername", ctx, username).Return(existingUser, nil)

	// Act
	result, err := userService.CreateUser(ctx, username, password, role)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
	assert.Equal(t, "Username already exists", appErr.Message)
	assert.Contains(t, appErr.Fields["username"], "This username is already taken")

	userRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_DatabaseError(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	username := "newuser"
	password := "ValidPass123!"
	role := models.RoleUser

	dbError := errors.New("database connection error")

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(nil, dbError)

	// Act
	result, err := userService.CreateUser(ctx, username, password, role)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrDatabaseError, appErr.Code)
	assert.Contains(t, appErr.Message, "error checking existing username")

	userRepo.AssertExpectations(t)
}

// TestUserService_UpdateUser tests
func TestUserService_UpdateUser_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	existingUser := createTestUser(userID, "oldusername", true)

	updates := map[string]interface{}{
		"username": "newusername",
		"role":     models.RoleAdmin,
		"isActive": false,
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	userRepo.On("FindByUsername", ctx, "newusername").Return(nil, nil) // Username not taken
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify the updated user
		user := args.Get(1).(*models.User)
		assert.Equal(t, "newusername", user.Username)
		assert.Equal(t, models.RoleAdmin, user.Role)
		assert.False(t, user.IsActive)
	})

	// Act
	result, err := userService.UpdateUser(ctx, userID, updates)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newusername", result.Username)
	assert.Equal(t, models.RoleAdmin, result.Role)
	assert.False(t, result.IsActive)
	assert.Empty(t, result.Password) // Password should be cleared

	userRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_WithPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	existingUser := createTestUser(userID, "testuser", true)
	newPassword := "NewSecurePass123!"

	updates := map[string]interface{}{
		"password": newPassword,
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify password is hashed
		user := args.Get(1).(*models.User)
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword))
		assert.NoError(t, err)
	})

	// Act
	result, err := userService.UpdateUser(ctx, userID, updates)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Password) // Password should be cleared

	userRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(999)

	updates := map[string]interface{}{
		"username": "newusername",
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	result, err := userService.UpdateUser(ctx, userID, updates)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	userRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_UsernameAlreadyTaken(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	existingUser := createTestUser(userID, "oldusername", true)
	otherUser := createTestUser(2, "newusername", true)

	updates := map[string]interface{}{
		"username": "newusername",
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	userRepo.On("FindByUsername", ctx, "newusername").Return(otherUser, nil) // Username taken by another user

	// Act
	result, err := userService.UpdateUser(ctx, userID, updates)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
	assert.Contains(t, appErr.Fields["username"], "This username is already taken")

	userRepo.AssertExpectations(t)
}

// TestUserService_DeleteUser tests
func TestUserService_DeleteUser_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "userToDelete", true)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify user is deactivated
		updatedUser := args.Get(1).(*models.User)
		assert.False(t, updatedUser.IsActive)
	})

	// Act
	err := userService.DeleteUser(ctx, userID)

	// Assert
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(999)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	err := userService.DeleteUser(ctx, userID)

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	userRepo.AssertExpectations(t)
}

// TestUserService_GetUser tests
func TestUserService_GetUser_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "testuser", true)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act
	result, err := userService.GetUser(ctx, userID)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Username, result.Username)
	assert.Empty(t, result.Password) // Password should be cleared

	userRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(999)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	result, err := userService.GetUser(ctx, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	userRepo.AssertExpectations(t)
}

// TestUserService_ChangePassword tests
func TestUserService_ChangePassword_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	currentPassword := "password123"
	newPassword := "NewSecurePass123!"

	// Create user with known password
	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: "testuser",
		Role:     models.RoleUser,
		IsActive: true,
	}
	// Set password using bcrypt
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify new password is hashed
		updatedUser := args.Get(1).(*models.User)
		err := bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		assert.NoError(t, err)
	})

	// Act
	err := userService.ChangePassword(ctx, userID, currentPassword, newPassword)

	// Assert
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_IncorrectCurrentPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "testuser", true)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act
	err := userService.ChangePassword(ctx, userID, "wrongpassword", "NewPass123!")

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "current password is incorrect")

	userRepo.AssertExpectations(t)
}

func TestUserService_ChangePassword_InvalidNewPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	currentPassword := "password123"

	// Create user with known password
	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: "testuser",
		Role:     models.RoleUser,
		IsActive: true,
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act - new password too short
	err := userService.ChangePassword(ctx, userID, currentPassword, "short")

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)

	userRepo.AssertExpectations(t)
}

// TestUserService_ResetPassword tests
func TestUserService_ResetPassword_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	newPassword := "NewSecurePass123!"
	user := createTestUser(userID, "testuser", true)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify password is hashed
		updatedUser := args.Get(1).(*models.User)
		err := bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		assert.NoError(t, err)
	})

	// Act
	err := userService.ResetPassword(ctx, userID, newPassword)

	// Assert
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(999)
	newPassword := "NewSecurePass123!"

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	err := userService.ResetPassword(ctx, userID, newPassword)

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)

	userRepo.AssertExpectations(t)
}

func TestUserService_ResetPassword_InvalidPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	logger := &test.MockLogger{}

	userService := services.NewUserService(userRepo, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "testuser", true)

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act - password too short
	err := userService.ResetPassword(ctx, userID, "short")

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)

	userRepo.AssertExpectations(t)
}
