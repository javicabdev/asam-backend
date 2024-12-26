package services

import (
	"context"
)

// NotificationService define las operaciones necesarias para enviar notificaciones
type NotificationService interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
	SendSMS(ctx context.Context, to string, message string) error
}

// emailNotificationService implementa NotificationService para enviar emails
type emailNotificationService struct {
	// Aquí irían las configuraciones necesarias para el servicio de email
	smtpServer   string
	smtpPort     int
	smtpUser     string
	smtpPassword string
}

// NewEmailNotificationService crea una nueva instancia del servicio de notificaciones por email
func NewEmailNotificationService(
	smtpServer string,
	smtpPort int,
	smtpUser string,
	smtpPassword string,
) NotificationService {
	return &emailNotificationService{
		smtpServer:   smtpServer,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
	}
}

func (s *emailNotificationService) SendEmail(ctx context.Context, to string, subject string, body string) error {
	// Aquí iría la implementación real del envío de email
	// Por ahora retornamos nil como placeholder
	return nil
}

func (s *emailNotificationService) SendSMS(ctx context.Context, to string, message string) error {
	// Aquí iría la implementación real del envío de SMS
	// Por ahora retornamos nil como placeholder
	return nil
}
