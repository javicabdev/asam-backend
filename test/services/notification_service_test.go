package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService es un mock simple para pruebas
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendEmail(ctx context.Context, to string, subject string, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *MockNotificationService) SendSMS(ctx context.Context, to string, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

func TestSendEmail(t *testing.T) {
	notificationService := new(MockNotificationService)

	notificationService.On("SendEmail",
		mock.Anything,
		"test@example.com",
		"Test Subject",
		"Test Body").Return(nil)

	err := notificationService.SendEmail(
		context.Background(),
		"test@example.com",
		"Test Subject",
		"Test Body")

	assert.NoError(t, err)
	notificationService.AssertExpectations(t)
}

func TestSendSMS(t *testing.T) {
	notificationService := new(MockNotificationService)

	notificationService.On("SendSMS",
		mock.Anything,
		"+34666777888",
		"Test message").Return(nil)

	err := notificationService.SendSMS(
		context.Background(),
		"+34666777888",
		"Test message")

	assert.NoError(t, err)
	notificationService.AssertExpectations(t)
}
