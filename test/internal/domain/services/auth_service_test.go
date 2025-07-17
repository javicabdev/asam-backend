package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// MockTokenRepository es un mock del repositorio de tokens
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) SaveRefreshToken(ctx context.Context, uuid string, userID uint, expires int64) error {
	args := m.Called(ctx, uuid, userID, expires)
	return args.Error(0)
}

func (m *MockTokenRepository) ValidateRefreshToken(ctx context.Context, uuid string, userID uint) error {
	args := m.Called(ctx, uuid, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, uuid string) error {
	args := m.Called(ctx, uuid)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteAllUserTokens(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) GetUserActiveSessions(ctx context.Context, userID uint) ([]*models.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTokenRepository) EnforceTokenLimitPerUser(ctx context.Context, maxTokens int) error {
	args := m.Called(ctx, maxTokens)
	return args.Error(0)
}

func createTestJWTUtil() *auth.JWTUtil {
	// Crear un JWTUtil real con configuración de prueba
	return auth.NewJWTUtil(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		7*24*time.Hour,
	)
}

// TestAuthService_Login tests
func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	user := createTestUser(1, username, true)

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(user, nil)
	tokenRepo.On("SaveRefreshToken", ctx, mock.AnythingOfType("string"), user.ID, mock.AnythingOfType("int64")).Return(nil)
	tokenRepo.On("EnforceTokenLimitPerUser", ctx, 5).Return(nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEmpty(t, result.AccessUUID)
	assert.NotEmpty(t, result.RefreshUUID)
	assert.Greater(t, result.AtExpires, time.Now().Unix())
	assert.Greater(t, result.RtExpires, result.AtExpires)

	// Verify all mocks were called
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "nonexistent"
	password := "password123"

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(nil, nil)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "credenciales inválidas")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "testuser"
	password := "wrongpassword"
	user := createTestUser(1, username, true)

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(user, nil)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "credenciales inválidas")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	user := createTestUser(1, username, false) // Usuario inactivo

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(user, nil)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInvalidStatus, appErr.Code)
	assert.Contains(t, appErr.Message, "usuario inactivo")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_DatabaseError(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	dbError := errors.New("database connection error")

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(nil, dbError)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrDatabaseError, appErr.Code)
	assert.Contains(t, appErr.Message, "error buscando usuario")

	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_SaveRefreshTokenError(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	username := "testuser"
	password := "password123"
	user := createTestUser(1, username, true)
	saveError := errors.New("failed to save token")

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(user, nil)
	tokenRepo.On("SaveRefreshToken", ctx, mock.AnythingOfType("string"), user.ID, mock.AnythingOfType("int64")).Return(saveError)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInternalError, appErr.Code)
	assert.Contains(t, appErr.Message, "error guardando refresh token")

	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Login_WithContextInfo(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	// Context with additional info
	ctx := context.Background()
	ctx = context.WithValue(ctx, constants.IPContextKey, "192.168.1.100")
	ctx = context.WithValue(ctx, constants.UserAgentContextKey, "Mozilla/5.0")
	ctx = context.WithValue(ctx, constants.DeviceNameContextKey, "Chrome Browser")

	username := "testuser"
	password := "password123"
	user := createTestUser(1, username, true)

	// Setup mocks
	userRepo.On("FindByUsername", ctx, username).Return(user, nil)
	tokenRepo.On("SaveRefreshToken", mock.Anything, mock.AnythingOfType("string"), user.ID, mock.AnythingOfType("int64")).Return(nil)
	tokenRepo.On("EnforceTokenLimitPerUser", ctx, 5).Return(nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	// Act
	result, err := authService.Login(ctx, username, password)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Verify the context was passed with additional info
	tokenRepo.AssertExpectations(t)
}

// TestAuthService_Logout tests
func TestAuthService_Logout_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	// Generar un token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(1, string(models.RoleUser))
	accessToken := tokenDetails.AccessToken

	// Setup mocks
	tokenRepo.On("DeleteRefreshToken", ctx, tokenDetails.AccessUUID).Return(nil)

	// Act
	err := authService.Logout(ctx, accessToken)

	// Assert
	assert.NoError(t, err)

	tokenRepo.AssertExpectations(t)
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	accessToken := "invalid-access-token"

	// Act
	err := authService.Logout(ctx, accessToken)

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "token inválido")
}

func TestAuthService_Logout_DeleteRefreshTokenError(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	// Generar un token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(1, string(models.RoleUser))
	accessToken := tokenDetails.AccessToken
	deleteError := errors.New("failed to delete token")

	// Setup mocks
	tokenRepo.On("DeleteRefreshToken", ctx, tokenDetails.AccessUUID).Return(deleteError)

	// Act
	err := authService.Logout(ctx, accessToken)

	// Assert
	assert.Error(t, err)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrInternalError, appErr.Code)
	assert.Contains(t, appErr.Message, "error eliminando refresh token")

	tokenRepo.AssertExpectations(t)
}

// TestAuthService_RefreshToken tests
func TestAuthService_RefreshToken_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "testuser", true)

	// Generar tokens válidos para el test
	oldTokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(user.Role))
	refreshToken := oldTokenDetails.RefreshToken

	// Setup mocks
	tokenRepo.On("ValidateRefreshToken", ctx, oldTokenDetails.RefreshUUID, userID).Return(nil)
	userRepo.On("FindByID", ctx, userID).Return(user, nil)
	tokenRepo.On("DeleteRefreshToken", ctx, oldTokenDetails.RefreshUUID).Return(nil)
	tokenRepo.On("SaveRefreshToken", ctx, mock.AnythingOfType("string"), user.ID, mock.AnythingOfType("int64")).Return(nil)

	// Act
	result, err := authService.RefreshToken(ctx, refreshToken)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	// Verificar que los nuevos tokens son diferentes a los antiguos
	assert.NotEqual(t, oldTokenDetails.AccessToken, result.AccessToken)
	assert.NotEqual(t, oldTokenDetails.RefreshToken, result.RefreshToken)

	// Verify all mocks were called
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	refreshToken := "invalid-refresh-token"

	// Act
	result, err := authService.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "refresh token inválido")
}

func TestAuthService_RefreshToken_TokenNotInDB(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)

	// Generar token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(models.RoleUser))
	refreshToken := tokenDetails.RefreshToken
	dbError := errors.New("token not found")

	// Setup mocks
	tokenRepo.On("ValidateRefreshToken", ctx, tokenDetails.RefreshUUID, userID).Return(dbError)

	// Act
	result, err := authService.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "refresh token no válido")

	tokenRepo.AssertExpectations(t)
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)

	// Generar token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(models.RoleUser))
	refreshToken := tokenDetails.RefreshToken

	// Setup mocks
	tokenRepo.On("ValidateRefreshToken", ctx, tokenDetails.RefreshUUID, userID).Return(nil)
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	result, err := authService.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)
	assert.Contains(t, appErr.Message, "usuario no encontrado")

	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// TestAuthService_ValidateToken tests
func TestAuthService_ValidateToken_Success(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)
	user := createTestUser(userID, "testuser", true)

	// Generar token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(user.Role))
	accessToken := tokenDetails.AccessToken

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(user, nil)

	// Act
	result, err := authService.ValidateToken(ctx, accessToken)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Username, result.Username)
	assert.Equal(t, user.Role, result.Role)

	userRepo.AssertExpectations(t)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	tokenString := "invalid-access-token"

	// Act
	result, err := authService.ValidateToken(ctx, tokenString)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
	assert.Contains(t, appErr.Message, "token inválido")
}

func TestAuthService_ValidateToken_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	jwtUtil := createTestJWTUtil()
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)

	// Generar token válido para el test
	tokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(models.RoleUser))
	accessToken := tokenDetails.AccessToken

	// Setup mocks
	userRepo.On("FindByID", ctx, userID).Return(nil, nil)

	// Act
	result, err := authService.ValidateToken(ctx, accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrNotFound, appErr.Code)
	assert.Contains(t, appErr.Message, "usuario no encontrado")

	userRepo.AssertExpectations(t)
}

// Test para verificar expiración de tokens
func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	// Arrange
	userRepo := new(MockUserRepository)
	tokenRepo := new(MockTokenRepository)
	verificationTokenRepo := new(test.MockVerificationTokenRepository)
	emailVerificationService := new(test.MockEmailVerificationService)
	// Crear JWTUtil con TTL muy corto para que expire inmediatamente
	jwtUtil := auth.NewJWTUtil(
		"test-access-secret",
		"test-refresh-secret",
		1*time.Nanosecond, // Token expira casi inmediatamente
		7*24*time.Hour,
	)
	logger := &test.MockLogger{}

	authService := services.NewAuthService(userRepo, jwtUtil, tokenRepo, verificationTokenRepo, emailVerificationService, logger)

	ctx := context.Background()
	userID := uint(1)

	// Generar token que expirará
	tokenDetails, _ := jwtUtil.GenerateTokenPair(userID, string(models.RoleUser))

	// Esperar para asegurar que el token expire
	time.Sleep(2 * time.Millisecond)

	// Act
	result, err := authService.ValidateToken(ctx, tokenDetails.AccessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	appErr, ok := appErrors.AsAppError(err)
	require.True(t, ok)
	assert.Equal(t, appErrors.ErrUnauthorized, appErr.Code)
}
