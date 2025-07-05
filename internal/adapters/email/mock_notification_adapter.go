package email

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// mockNotificationAdapter implements EmailNotificationService for development/testing
type mockNotificationAdapter struct {
	logger logger.Logger
}

// NewMockNotificationAdapter creates a new mock email notification adapter
func NewMockNotificationAdapter(logger logger.Logger) input.EmailNotificationService {
	return &mockNotificationAdapter{
		logger: logger,
	}
}

// SendVerificationEmail logs the verification email details
func (m *mockNotificationAdapter) SendVerificationEmail(ctx context.Context, user *models.User, verificationURL string) error {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	m.logger.Info("MOCK EMAIL: Verification email",
		zap.String("to", email),
		zap.String("username", user.Username),
		zap.String("verification_url", verificationURL),
	)

	// Log the email content for development
	m.logger.Debug(fmt.Sprintf(`
=== MOCK EMAIL SERVICE ===
TO: %s
SUBJECT: Verifica tu correo electrónico - ASAM
---
Hola %s,

Por favor verifica tu correo electrónico visitando:
%s

Este enlace expirará en 24 horas.
=========================
`, email, user.Username, verificationURL))

	return nil
}

// SendPasswordResetEmail logs the password reset email details
func (m *mockNotificationAdapter) SendPasswordResetEmail(ctx context.Context, user *models.User, resetURL string) error {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	m.logger.Info("MOCK EMAIL: Password reset email",
		zap.String("to", email),
		zap.String("username", user.Username),
		zap.String("reset_url", resetURL),
	)

	// Log the email content for development
	m.logger.Debug(fmt.Sprintf(`
=== MOCK EMAIL SERVICE ===
TO: %s
SUBJECT: Restablecer contraseña - ASAM
---
Hola %s,

Has solicitado restablecer tu contraseña. 
Visita el siguiente enlace:
%s

Este enlace expirará en 1 hora.
=========================
`, email, user.Username, resetURL))

	return nil
}

// SendWelcomeEmail logs the welcome email details
func (m *mockNotificationAdapter) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	m.logger.Info("MOCK EMAIL: Welcome email",
		zap.String("to", email),
		zap.String("username", user.Username),
	)

	// Log the email content for development
	m.logger.Debug(fmt.Sprintf(`
=== MOCK EMAIL SERVICE ===
TO: %s
SUBJECT: ¡Bienvenido a ASAM!
---
Hola %s,

¡Tu cuenta ha sido verificada exitosamente!
Ahora formas parte de la Asociación de Ayuda Mutua (ASAM).

Con tu cuenta puedes:
- Gestionar tu información personal
- Ver el estado de tus pagos
- Acceder a los beneficios de la asociación
- Mantenerte informado sobre las actividades
=========================
`, email, user.Username))

	return nil
}

// SendPasswordChangedEmail logs the password changed notification
func (m *mockNotificationAdapter) SendPasswordChangedEmail(ctx context.Context, user *models.User) error {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	m.logger.Info("MOCK EMAIL: Password changed notification",
		zap.String("to", email),
		zap.String("username", user.Username),
	)

	// Log the email content for development
	m.logger.Debug(fmt.Sprintf(`
=== MOCK EMAIL SERVICE ===
TO: %s
SUBJECT: Tu contraseña ha sido cambiada - ASAM
---
Hola %s,

Te informamos que la contraseña de tu cuenta en ASAM 
ha sido actualizada exitosamente.

Si no has realizado este cambio, por favor contacta 
inmediatamente con el administrador del sistema.
=========================
`, email, user.Username))

	return nil
}
