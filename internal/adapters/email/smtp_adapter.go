package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// smtpAdapter implements EmailNotificationService using SMTP
type smtpAdapter struct {
	config SMTPConfig
	logger logger.Logger
}

// NewSMTPAdapter creates a new SMTP email adapter
func NewSMTPAdapter(config SMTPConfig, logger logger.Logger) input.EmailNotificationService {
	return &smtpAdapter{
		config: config,
		logger: logger,
	}
}

// sendEmail sends an email using SMTP
func (a *smtpAdapter) sendEmail(to, subject, body string) error {
	// Validate inputs to prevent SMTP injection
	if strings.ContainsAny(to, "\r\n") {
		a.logger.Error("Invalid recipient email", zap.String("to", to))
		return errors.NewValidationError("recipient email contains invalid characters", nil)
	}
	if strings.ContainsAny(subject, "\r\n") {
		a.logger.Error("Invalid subject", zap.String("subject", subject))
		return errors.NewValidationError("subject contains invalid characters", nil)
	}
	if strings.ContainsAny(a.config.From, "\r\n") {
		a.logger.Error("Invalid sender email", zap.String("from", a.config.From))
		return errors.NewValidationError("sender email contains invalid characters", nil)
	}

	auth := smtp.PlainAuth("", a.config.Username, a.config.Password, a.config.Host)

	// Compose the email
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
			"\r\n"+
			"%s",
		a.config.From,
		to,
		subject,
		body,
	))

	// Send the email
	addr := fmt.Sprintf("%s:%d", a.config.Host, a.config.Port)

	if err := smtp.SendMail(addr, auth, a.config.From, []string{to}, msg); err != nil {
		a.logger.Error("Failed to send email via SMTP", zap.String("to", to), zap.Error(err))
		return errors.Wrap(err, errors.ErrNetworkError, "failed to send email")
	}

	return nil
}

// SendVerificationEmail sends an email verification link to the user
func (a *smtpAdapter) SendVerificationEmail(ctx context.Context, user *models.User, verificationURL string) error {
	if user.Email == "" {
		return errors.NewValidationError("user email is empty", nil)
	}

	subject := "Verifica tu correo electrónico - ASAM"

	tmpl := template.Must(template.New("verification").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Asociación de Ayuda Mutua (ASAM)</h1>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>
            <p>Gracias por registrarte en ASAM. Para completar tu registro, por favor verifica tu dirección de correo electrónico haciendo clic en el siguiente enlace:</p>
            <center>
                <a href="{{.VerificationURL}}" class="button">Verificar mi correo</a>
            </center>
            <p>O copia y pega este enlace en tu navegador:</p>
            <p style="word-break: break-all;">{{.VerificationURL}}</p>
            <p>Este enlace expirará en 24 horas.</p>
            <p>Si no has solicitado este registro, puedes ignorar este correo.</p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]string{
		"Username":        user.Username,
		"VerificationURL": verificationURL,
	})
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "failed to generate email template")
	}

	return a.sendEmail(user.Email, subject, buf.String())
}

// SendPasswordResetEmail sends a password reset link to the user
func (a *smtpAdapter) SendPasswordResetEmail(ctx context.Context, user *models.User, resetURL string) error {
	if user.Email == "" {
		a.logger.Error("User email is empty", zap.Uint("userID", user.ID))
		return errors.NewValidationError("user email is empty", nil)
	}

	subject := "Restablecer contraseña - ASAM"

	tmpl := template.Must(template.New("reset").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { text-align: center; padding: 20px; font-size: 0.9em; color: #666; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 12px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Asociación de Ayuda Mutua (ASAM)</h1>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>
            <p>Hemos recibido una solicitud para restablecer la contraseña de tu cuenta.</p>
            <p>Si has solicitado este cambio, haz clic en el siguiente enlace para crear una nueva contraseña:</p>
            <center>
                <a href="{{.ResetURL}}" class="button">Restablecer mi contraseña</a>
            </center>
            <p>O copia y pega este enlace en tu navegador:</p>
            <p style="word-break: break-all;">{{.ResetURL}}</p>
            <div class="warning">
                <strong>⚠️ Importante:</strong> Este enlace expirará en 1 hora por razones de seguridad.
            </div>
            <p>Si no has solicitado restablecer tu contraseña, ignora este correo. Tu contraseña actual seguirá siendo válida.</p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]string{
		"Username": user.Username,
		"ResetURL": resetURL,
	})
	if err != nil {
		a.logger.Error("Failed to generate email template", zap.Error(err))
		return errors.Wrap(err, errors.ErrInternalError, "failed to generate email template")
	}

	return a.sendEmail(user.Email, subject, buf.String())
}

// SendWelcomeEmail sends a welcome email to a new user
func (a *smtpAdapter) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	if user.Email == "" {
		return errors.NewValidationError("user email is empty", nil)
	}

	subject := "¡Bienvenido a ASAM!"

	tmpl := template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { text-align: center; padding: 20px; font-size: 0.9em; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>¡Bienvenido a ASAM!</h1>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>
            <p>¡Tu cuenta ha sido verificada exitosamente!</p>
            <p>Ahora formas parte de la Asociación de Ayuda Mutua (ASAM). Estamos encantados de tenerte con nosotros.</p>
            <p>Con tu cuenta puedes:</p>
            <ul>
                <li>Gestionar tu información personal</li>
                <li>Ver el estado de tus pagos</li>
                <li>Acceder a los beneficios de la asociación</li>
                <li>Mantenerte informado sobre las actividades</li>
            </ul>
            <p>Si tienes alguna pregunta, no dudes en contactarnos.</p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]string{
		"Username": user.Username,
	})
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "failed to generate email template")
	}

	return a.sendEmail(user.Email, subject, buf.String())
}

// SendPasswordChangedEmail sends a notification that password was changed
func (a *smtpAdapter) SendPasswordChangedEmail(ctx context.Context, user *models.User) error {
	if user.Email == "" {
		return errors.NewValidationError("user email is empty", nil)
	}

	subject := "Tu contraseña ha sido cambiada - ASAM"

	tmpl := template.Must(template.New("password-changed").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #ffc107; color: #333; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .footer { text-align: center; padding: 20px; font-size: 0.9em; color: #666; }
        .alert { background-color: #f8d7da; border: 1px solid #f5c6cb; padding: 12px; margin: 20px 0; border-radius: 4px; color: #721c24; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Contraseña Actualizada</h1>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>
            <p>Te informamos que la contraseña de tu cuenta en ASAM ha sido actualizada exitosamente.</p>
            <div class="alert">
                <strong>⚠️ ¿No reconoces este cambio?</strong><br>
                Si no has realizado este cambio, por favor contacta inmediatamente con el administrador del sistema.
            </div>
            <p>Por razones de seguridad, te recomendamos:</p>
            <ul>
                <li>No compartir tu contraseña con nadie</li>
                <li>Usar una contraseña única para ASAM</li>
                <li>Cambiar tu contraseña regularmente</li>
            </ul>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]string{
		"Username": user.Username,
	})
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "failed to generate email template")
	}

	return a.sendEmail(user.Email, subject, buf.String())
}
