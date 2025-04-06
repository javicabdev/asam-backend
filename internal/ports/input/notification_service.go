package input

import (
	"context"
)

// NotificationService define las operaciones necesarias para enviar notificaciones
type NotificationService interface {
	SendEmail(ctx context.Context, to string, subject string, body string) error
	SendSMS(ctx context.Context, to string, message string) error
}
