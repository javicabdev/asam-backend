package resolvers

import (
	"context"
	"testing"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockAuthService implementa la interfaz input.AuthService para testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*input.TokenDetails, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.TokenDetails), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*input.TokenDetails, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.TokenDetails), args.Error(1)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestLogin(t *testing.T) {
	// Crear mock del servicio de autenticación
	mockAuthService := new(MockAuthService)

	// Crear el resolver
	resolver := &Resolver{
		authService: mockAuthService,
	}

	// Crear contexto de prueba
	ctx := context.Background()

	// Configurar el comportamiento del mock
	mockTokenDetails := &input.TokenDetails{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		AccessUuid:   "test-access-uuid",
		RefreshUuid:  "test-refresh-uuid",
		AtExpires:    1620000000,
		RtExpires:    1620100000,
	}

	// Crear el usuario para prueba
	mockUser := &models.User{
		Model:    gorm.Model{ID: 1},
		Username: "testuser",
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	// Configurar el comportamiento esperado del mock
	mockAuthService.On("Login", ctx, "testuser", "testpassword").Return(mockTokenDetails, nil)
	mockAuthService.On("ValidateToken", ctx, "test-access-token").Return(mockUser, nil)

	// Llamar a la función Login
	authResponse, err := resolver.Login(ctx, model.LoginInput{
		Username: "testuser",
		Password: "testpassword",
	})

	// Verificar que no hay error
	assert.NoError(t, err)
	assert.NotNil(t, authResponse)

	// Verificar los valores devueltos
	assert.Equal(t, uint64(1), uint64(authResponse.User.ID))
	assert.Equal(t, "testuser", authResponse.User.Username)
	assert.Equal(t, models.RoleAdmin, authResponse.User.Role)
	assert.Equal(t, "test-access-token", authResponse.AccessToken)
	assert.Equal(t, "test-refresh-token", authResponse.RefreshToken)

	// Verificar que se llamaron los métodos esperados
	mockAuthService.AssertExpectations(t)
}

func TestLoginValidationError(t *testing.T) {
	// Crear el resolver con un mock vacío
	mockAuthService := new(MockAuthService)
	resolver := &Resolver{
		authService: mockAuthService,
	}

	// Llamar a Login con username vacío
	_, err := resolver.Login(context.Background(), model.LoginInput{
		Username: "",
		Password: "testpassword",
	})

	// Verificar que se devuelve un error de validación
	assert.Error(t, err)
	appErr, ok := errors.AsAppError(err)
	assert.True(t, ok)
	assert.Equal(t, errors.ErrValidationFailed, appErr.Code)
	assert.Contains(t, appErr.Fields, "username")

	// Verificar que no se llamó al servicio
	mockAuthService.AssertNotCalled(t, "Login")
}

func TestLogout(t *testing.T) {
	// Crear mock y resolver
	mockAuthService := new(MockAuthService)
	resolver := &Resolver{
		authService: mockAuthService,
	}

	// Crear contexto con token
	ctx := context.WithValue(context.Background(), "authorization", "Bearer test-token")

	// Configurar mock
	mockAuthService.On("Logout", ctx, "test-token").Return(nil)

	// Llamar a la función
	result, err := resolver.Logout(ctx)

	// Verificar resultado
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mutationResponse, ok := result.(*model.MutationResponse)
	assert.True(t, ok)
	assert.True(t, mutationResponse.Success)
	assert.NotNil(t, mutationResponse.Message)
	assert.Nil(t, mutationResponse.Error)

	// Verificar llamada al mock
	mockAuthService.AssertExpectations(t)
}

func TestRefreshToken(t *testing.T) {
	// Crear mock y resolver
	mockAuthService := new(MockAuthService)
	resolver := &Resolver{
		authService: mockAuthService,
	}

	// Configurar mock
	mockTokenDetails := &input.TokenDetails{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		AccessUuid:   "new-access-uuid",
		RefreshUuid:  "new-refresh-uuid",
		AtExpires:    1630000000,
		RtExpires:    1630100000,
	}

	mockAuthService.On("RefreshToken", mock.Anything, "old-refresh-token").Return(mockTokenDetails, nil)

	// Llamar a la función
	result, err := resolver.RefreshToken(context.Background(), model.RefreshTokenInput{
		RefreshToken: "old-refresh-token",
	})

	// Verificar resultado
	assert.NoError(t, err)
	assert.NotNil(t, result)
	tokenResponse, ok := result.(*model.TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "new-access-token", tokenResponse.AccessToken)
	assert.Equal(t, "new-refresh-token", tokenResponse.RefreshToken)

	// Verificar llamada al mock
	mockAuthService.AssertExpectations(t)
}
