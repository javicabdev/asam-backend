package output

import (
	"context"
)

// EmailService define la interfaz para enviar emails
type EmailService interface {
	// SendEmail envía un email genérico
	SendEmail(ctx context.Context, to, subject, body string) error

	// SendHTMLEmail envía un email con contenido HTML
	SendHTMLEmail(ctx context.Context, to, subject, htmlBody string) error

	// SendVerificationEmail envía un email de verificación
	SendVerificationEmail(ctx context.Context, to, username, verificationURL string) error

	// SendPasswordResetEmail envía un email de recuperación de contraseña
	SendPasswordResetEmail(ctx context.Context, to, username, resetURL string) error

	// SendPasswordChangedEmail envía una notificación de cambio de contraseña
	SendPasswordChangedEmail(ctx context.Context, to, username string) error
}
