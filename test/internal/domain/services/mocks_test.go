package services_test

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// MockUserRepository es un mock del repositorio de usuarios
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// MockVerificationTokenRepository es un mock del repositorio de tokens de verificación
type MockVerificationTokenRepository struct {
	mock.Mock
}

func (m *MockVerificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) GetByToken(ctx context.Context, token string) (*models.VerificationToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VerificationToken), args.Error(1)
}

func (m *MockVerificationTokenRepository) Update(ctx context.Context, token *models.VerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) InvalidateUserTokens(ctx context.Context, userID uint, tokenType string) error {
	args := m.Called(ctx, userID, tokenType)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) CountActiveTokensByUser(ctx context.Context, userID uint, tokenType string) (int64, error) {
	args := m.Called(ctx, userID, tokenType)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVerificationTokenRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVerificationTokenRepository) GetByUserIDAndType(ctx context.Context, userID uint, tokenType string) ([]*models.VerificationToken, error) {
	args := m.Called(ctx, userID, tokenType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.VerificationToken), args.Error(1)
}

// MockEmailService es un mock del servicio de email
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendHTMLEmail(ctx context.Context, to, subject, htmlBody string) error {
	args := m.Called(ctx, to, subject, htmlBody)
	return args.Error(0)
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, to, username, verificationURL string) error {
	args := m.Called(ctx, to, username, verificationURL)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, to, username, resetURL string) error {
	args := m.Called(ctx, to, username, resetURL)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordChangedEmail(ctx context.Context, to, username string) error {
	args := m.Called(ctx, to, username)
	return args.Error(0)
}

// createTestUser crea un usuario de prueba
func createTestUser(id uint, username string, active bool) *models.User {
	user := &models.User{
		Model:     gorm.Model{ID: id},
		Username:  username,
		Role:      models.RoleUser,
		IsActive:  active,
		LastLogin: time.Now().Add(-24 * time.Hour),
	}
	// Set password using the model's method to ensure proper hashing
	user.SetPassword("password123")
	return user
}
