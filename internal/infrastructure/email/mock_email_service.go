package infrastructure

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"go.uber.org/zap"
)

// MockEmailService es una implementación mock del servicio de email para desarrollo
type MockEmailService struct {
	logger logger.Logger
}

// NewMockEmailService crea una nueva instancia del servicio de email mock
func NewMockEmailService(logger logger.Logger) output.EmailService {
	return &MockEmailService{
		logger: logger,
	}
}

// SendEmail simula el envío de un email de texto plano
func (s *MockEmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	s.logger.Info("Mock: Sending email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.Int("body_length", len(body)),
	)
	return nil
}

// SendHTMLEmail simula el envío de un email HTML
func (s *MockEmailService) SendHTMLEmail(ctx context.Context, to, subject, htmlBody string) error {
	s.logger.Info("Mock: Sending HTML email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.Int("html_length", len(htmlBody)),
	)
	return nil
}

// SendVerificationEmail simula el envío de un email de verificación
func (s *MockEmailService) SendVerificationEmail(ctx context.Context, to, username, verificationURL string) error {
	s.logger.Info("Mock: Sending verification email",
		zap.String("to", to),
		zap.String("username", username),
		zap.String("verification_url", verificationURL),
	)
	return nil
}

// SendPasswordResetEmail simula el envío de un email de recuperación de contraseña
func (s *MockEmailService) SendPasswordResetEmail(ctx context.Context, to, username, resetURL string) error {
	s.logger.Info("Mock: Sending password reset email",
		zap.String("to", to),
		zap.String("username", username),
		zap.String("reset_url", resetURL),
	)
	return nil
}

// SendPasswordChangedEmail simula el envío de una notificación de cambio de contraseña
func (s *MockEmailService) SendPasswordChangedEmail(ctx context.Context, to, username string) error {
	s.logger.Info("Mock: Sending password changed notification",
		zap.String("to", to),
		zap.String("username", username),
	)
	return nil
}
