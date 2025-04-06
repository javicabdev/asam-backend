package services

import (
	"context"
	"testing"

	"github.com/javicabdev/asam-backend/pkg/errors" // Biblioteca de errores personalizados
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNotificationService es un mock para pruebas del servicio de notificaciones
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
	tests := []struct {
		name      string
		to        string
		subject   string
		body      string
		setupMock func(*MockNotificationService)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:    "successful email sending",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendEmail", mock.Anything, "test@example.com", "Test Subject", "Test Body").Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "invalid email format",
			to:      "invalid_email",
			subject: "Test Subject",
			body:    "Test Body",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendEmail", mock.Anything, "invalid_email", "Test Subject", "Test Body").
					Return(errors.NewValidationError("invalid email format", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name:    "network error",
			to:      "test@example.com",
			subject: "Test Subject",
			body:    "Test Body",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendEmail", mock.Anything, "test@example.com", "Test Subject", "Test Body").
					Return(errors.NetworkError("failed to connect to email server", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNetworkError(err), "debería ser un error de red")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notificationService := new(MockNotificationService)
			tt.setupMock(notificationService)

			err := notificationService.SendEmail(context.Background(), tt.to, tt.subject, tt.body)

			tt.checkErr(t, err)
			notificationService.AssertExpectations(t)
		})
	}
}

func TestSendSMS(t *testing.T) {
	tests := []struct {
		name      string
		to        string
		message   string
		setupMock func(*MockNotificationService)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:    "successful SMS sending",
			to:      "+34666777888",
			message: "Test message",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendSMS", mock.Anything, "+34666777888", "Test message").Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:    "invalid phone number",
			to:      "invalid_phone",
			message: "Test message",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendSMS", mock.Anything, "invalid_phone", "Test message").
					Return(errors.NewValidationError("invalid phone number", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name:    "network error",
			to:      "+34666777888",
			message: "Test message",
			setupMock: func(ns *MockNotificationService) {
				ns.On("SendSMS", mock.Anything, "+34666777888", "Test message").
					Return(errors.NetworkError("failed to connect to SMS gateway", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNetworkError(err), "debería ser un error de red")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notificationService := new(MockNotificationService)
			tt.setupMock(notificationService)

			err := notificationService.SendSMS(context.Background(), tt.to, tt.message)

			tt.checkErr(t, err)
			notificationService.AssertExpectations(t)
		})
	}
}
