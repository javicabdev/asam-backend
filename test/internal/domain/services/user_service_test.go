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
	"github.com/javicabdev/asam-backend/internal/ports/input"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// Helper function to create user service with mocks
func setupUserServiceTest() (input.UserService, *MockUserRepository, *MockVerificationTokenRepository, *MockEmailService) {
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockVerificationTokenRepository)
	emailService := new(MockEmailService)
	logger := &test.MockLogger{}
	baseURL := "http://test.example.com"

	userService := services.NewUserService(userRepo, tokenRepo, emailService, logger, baseURL)

	return userService, userRepo, tokenRepo, emailService
}

// TestUserService_CreateUser tests
func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange
	userService, userRepo, _, _ := setupUserServiceTest()

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

func TestUserService_CreateUser_WithEmail_Success(t *testing.T) {
	// Arrange
	userService, userRepo, tokenRepo, emailService := setupUserServiceTest()

	ctx := context.Background()
	username := "newuser@example.com"
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
		assert.False(t, user.EmailVerified) // Should not be verified for email username

		// Verify password is hashed
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		assert.NoError(t, err)

		// Set ID to simulate DB behavior
		user.ID = 1
	})

	// Email verification should be sent
	userRepo.On("FindByID", ctx, uint(1)).Return(&models.User{
		Model:         gorm.Model{ID: 1},
		Username:      username,
		EmailVerified: false,
	}, nil)
	tokenRepo.On("InvalidateUserTokens", ctx, uint(1), string(models.TokenTypeEmailVerification)).Return(nil)
	tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil)
	emailService.On("SendVerificationEmail", ctx, username, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

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
	userService, _, _, _ := setupUserServiceTest()

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
			username: "thisisaverylongusernamethatexceedsonehundredcharactersandshouldfailvalidationbecauseitistoolongforusX",
			errMsg:   "Username must not exceed 100 characters",
		},
		{
			name:     "Invalid characters",
			username: "user#$%!",
			errMsg:   "Username can only contain letters, numbers, underscore, hyphen, and dot",
		},
		{
			name:     "Invalid email - no domain",
			username: "user@",
			errMsg:   "Please provide a valid email address",
		},
		{
			name:     "Invalid email - no local part",
			username: "@example.com",
			errMsg:   "Please provide a valid email address",
		},
		{
			name:     "Invalid email - multiple @",
			username: "user@@example.com",
			errMsg:   "Please provide a valid email address",
		},
		{
			name:     "Invalid email - no TLD",
			username: "user@example",
			errMsg:   "Please provide a valid email address",
		},
		{
			name:     "Invalid email - consecutive dots",
			username: "user..name@example.com",
			errMsg:   "Email cannot contain consecutive dots",
		},
		{
			name:     "Invalid email - starts with dot",
			username: ".user@example.com",
			errMsg:   "Email local part cannot start or end with a dot",
		},
		{
			name:     "Invalid email - ends with dot",
			username: "user.@example.com",
			errMsg:   "Email local part cannot start or end with a dot",
		},
		{
			name:     "Invalid email - local part too long",
			username: "verylonglocalpartthathassixtyfivorcharactersandexceedsthelimitset@example.com",
			errMsg:   "Email local part must be between 1 and 64 characters",
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

func TestUserService_CreateUser_ValidEmails(t *testing.T) {
	// Arrange
	userService, userRepo, tokenRepo, emailService := setupUserServiceTest()

	ctx := context.Background()
	password := "ValidPass123!"
	role := models.RoleUser

	validEmails := []string{
		"user@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user_name@example.com",
		"user-name@example.com",
		"user123@example.com",
		"123user@example.com",
		"u@example.com",
		"user@subdomain.example.com",
		"user@example.co.uk",
		"user%test@example.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			// Setup mocks
			userRepo.On("FindByUsername", ctx, email).Return(nil, nil).Once()
			userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Once().Run(func(args mock.Arguments) {
				user := args.Get(1).(*models.User)
				user.ID = 1
			})

			// Email verification mocks
			userRepo.On("FindByID", ctx, uint(1)).Return(&models.User{
				Model:         gorm.Model{ID: 1},
				Username:      email,
				EmailVerified: false,
			}, nil).Once()
			tokenRepo.On("InvalidateUserTokens", ctx, uint(1), string(models.TokenTypeEmailVerification)).Return(nil).Once()
			tokenRepo.On("Create", ctx, mock.AnythingOfType("*models.VerificationToken")).Return(nil).Once()
			emailService.On("SendVerificationEmail", ctx, email, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Once()

			// Act
			result, err := userService.CreateUser(ctx, email, password, role)

			// Assert
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, email, result.Username)
		})
	}

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_CreateUser_InvalidPassword(t *testing.T) {
	// Arrange
	userService, _, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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

func TestUserService_UpdateUser_ToEmail_Success(t *testing.T) {
	// Arrange
	userService, userRepo, _, _ := setupUserServiceTest()

	ctx := context.Background()
	userID := uint(1)
	existingUser := createTestUser(userID, "oldusername", true)
	newEmail := "newemail@example.com"

	updates := map[string]interface{}{
		"username": newEmail,
	}

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(existingUser, nil)
	userRepo.On("FindByUsername", ctx, newEmail).Return(nil, nil) // Email not taken
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
		// Verify the updated user
		user := args.Get(1).(*models.User)
		assert.Equal(t, newEmail, user.Username)
		assert.False(t, user.EmailVerified) // Should be marked as unverified
	})

	// Act
	result, err := userService.UpdateUser(ctx, userID, updates)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newEmail, result.Username)

	userRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_WithPassword(t *testing.T) {
	// Arrange
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, emailService := setupUserServiceTest()

	ctx := context.Background()
	userID := uint(1)
	currentPassword := "password123"
	newPassword := "NewSecurePass123!"

	// Create user with known password
	user := &models.User{
		Model:    gorm.Model{ID: userID},
		Username: "test@example.com",
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
	emailService.On("SendPasswordChangedEmail", ctx, "test@example.com", "test").Return(nil)

	// Act
	err := userService.ChangePassword(ctx, userID, currentPassword, newPassword)

	// Assert
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestUserService_ChangePassword_IncorrectCurrentPassword(t *testing.T) {
	// Arrange
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	assert.Contains(t, appErr.Fields["password"], "Password must be at least 8 characters long")

	userRepo.AssertExpectations(t)
}

// TestUserService_ResetPassword tests
func TestUserService_ResetPassword_Success(t *testing.T) {
	// Arrange
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, userRepo, _, _ := setupUserServiceTest()

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
	userService, _, _, _ := setupUserServiceTest()

	ctx := context.Background()
	userID := uint(1)

	// Act - password too short, no mocks needed since validation happens first
	err := userService.ResetPassword(ctx, userID, "short")

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrValidationFailed, appErr.Code)
	assert.Contains(t, appErr.Fields["password"], "Password must be at least 8 characters long")
}
