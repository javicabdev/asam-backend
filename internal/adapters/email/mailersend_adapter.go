package email

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/mailersend/mailersend-go"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// MailerSendConfig holds MailerSend configuration
type MailerSendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

// mailersendAdapter implements EmailNotificationService using MailerSend
type mailersendAdapter struct {
	client *mailersend.Mailersend
	config MailerSendConfig
	logger logger.Logger
}

// NewMailerSendAdapter creates a new MailerSend email adapter
func NewMailerSendAdapter(config MailerSendConfig, logger logger.Logger) input.EmailNotificationService {
	client := mailersend.NewMailersend(config.APIKey)

	return &mailersendAdapter{
		client: client,
		config: config,
		logger: logger,
	}
}

// sendEmail sends an email using MailerSend API
func (a *mailersendAdapter) sendEmail(to, toName, subject, htmlContent string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	from := mailersend.From{
		Name:  a.config.FromName,
		Email: a.config.FromEmail,
	}

	recipients := []mailersend.Recipient{
		{
			Name:  toName,
			Email: to,
		},
	}

	// Create message
	message := a.client.Email.NewMessage()
	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(htmlContent)

	// Add tags for better tracking
	message.SetTags([]string{"asam", "transactional"})

	// Send the email
	response, err := a.client.Email.Send(ctx, message)
	if err != nil {
		a.logger.Error("Failed to send email via MailerSend",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err))
		return errors.Wrap(err, errors.ErrNetworkError, "failed to send email via MailerSend")
	}

	// Log success with message ID for tracking
	if response != nil && response.Header != nil {
		messageID := response.Header.Get("X-Message-Id")
		a.logger.Info("Email sent successfully via MailerSend",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.String("message_id", messageID))
	}

	return nil
}

// SendVerificationEmail sends an email verification link to the user
func (a *mailersendAdapter) SendVerificationEmail(_ context.Context, user *models.User, verificationURL string) error {
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
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; border: 1px solid #e0e0e0; border-top: none; }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #007bff;
            color: #ffffff !important;
            text-decoration: none !important;
            border-radius: 5px;
            margin: 20px 0;
            font-weight: bold;
            font-size: 16px;
        }
        .button:hover {
            background-color: #0056b3;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 0.9em;
            color: #666;
            background-color: #f5f5f5;
            border: 1px solid #e0e0e0;
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        .url-container {
            background-color: #f0f0f0;
            padding: 10px;
            border-radius: 4px;
            word-break: break-all;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Asociación de Ayuda Mutua (ASAM)</h1>
        </div>
        <div class="content">
            <h2>¡Hola {{.Username}}!</h2>
            <p>Gracias por registrarte en ASAM. Para completar tu registro y activar tu cuenta, por favor verifica tu dirección de correo electrónico.</p>

            <div style="text-align: center; margin: 30px 0;">
                <a href="{{.VerificationURL}}" class="button">Verificar mi correo electrónico</a>
            </div>

            <p>Si el botón no funciona, puedes copiar y pegar este enlace en tu navegador:</p>
            <div class="url-container">
                {{.VerificationURL}}
            </div>

            <p><strong>⏰ Este enlace expirará en 24 horas.</strong></p>

            <p>Si no has solicitado este registro, puedes ignorar este correo de forma segura.</p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
            <p>Este es un correo automático, por favor no respondas a este mensaje.</p>
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

	return a.sendEmail(user.Email, user.Username, subject, buf.String())
}

// SendPasswordResetEmail sends a password reset link to the user
func (a *mailersendAdapter) SendPasswordResetEmail(_ context.Context, user *models.User, resetURL string) error {
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
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { padding: 20px; background-color: #f9f9f9; border: 1px solid #e0e0e0; border-top: none; }
        .button {
            display: inline-block;
            padding: 12px 30px;
            background-color: #dc3545;
            color: #ffffff !important;
            text-decoration: none !important;
            border-radius: 5px;
            margin: 20px 0;
            font-weight: bold;
            font-size: 16px;
        }
        .button:hover {
            background-color: #bd2130;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 0.9em;
            color: #666;
            background-color: #f5f5f5;
            border: 1px solid #e0e0e0;
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        .warning {
            background-color: #fff3cd;
            border: 1px solid #ffeaa7;
            padding: 12px;
            margin: 20px 0;
            border-radius: 4px;
            color: #856404;
        }
        .url-container {
            background-color: #f0f0f0;
            padding: 10px;
            border-radius: 4px;
            word-break: break-all;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Restablecer Contraseña</h1>
            <p style="margin: 0;">Asociación de Ayuda Mutua (ASAM)</p>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>
            <p>Hemos recibido una solicitud para restablecer la contraseña de tu cuenta en ASAM.</p>

            <p>Si has solicitado este cambio, haz clic en el siguiente botón para crear una nueva contraseña:</p>

            <div style="text-align: center; margin: 30px 0;">
                <a href="{{.ResetURL}}" class="button">Restablecer mi contraseña</a>
            </div>

            <p>Si el botón no funciona, puedes copiar y pegar este enlace en tu navegador:</p>
            <div class="url-container">
                {{.ResetURL}}
            </div>

            <div class="warning">
                <strong>⚠️ Importante:</strong>
                <ul style="margin: 5px 0; padding-left: 20px;">
                    <li>Este enlace expirará en <strong>1 hora</strong> por razones de seguridad.</li>
                    <li>Solo puedes usar este enlace una vez.</li>
                    <li>Si no has solicitado este cambio, ignora este correo y tu contraseña permanecerá sin cambios.</li>
                </ul>
            </div>

            <p>Por tu seguridad, te recomendamos:</p>
            <ul>
                <li>Usar una contraseña única y segura</li>
                <li>No compartir tu contraseña con nadie</li>
                <li>Cambiar tu contraseña regularmente</li>
            </ul>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
            <p>Este es un correo automático, por favor no respondas a este mensaje.</p>
            <p>Si necesitas ayuda, contacta con el administrador del sistema.</p>
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

	return a.sendEmail(user.Email, user.Username, subject, buf.String())
}

// SendWelcomeEmail sends a welcome email to a new user
func (a *mailersendAdapter) SendWelcomeEmail(_ context.Context, user *models.User) error {
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
        .header {
            background: linear-gradient(135deg, #28a745 0%, #20c997 100%);
            color: white;
            padding: 30px 20px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            padding: 30px 20px;
            background-color: #ffffff;
            border: 1px solid #e0e0e0;
            border-top: none;
        }
        .features {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .feature-item {
            padding: 10px 0;
            border-bottom: 1px solid #e0e0e0;
        }
        .feature-item:last-child {
            border-bottom: none;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 0.9em;
            color: #666;
            background-color: #f5f5f5;
            border: 1px solid #e0e0e0;
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        .success-icon {
            font-size: 48px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="success-icon">✅</div>
            <h1 style="margin: 10px 0;">¡Bienvenido a ASAM!</h1>
            <p style="margin: 0; opacity: 0.9;">Tu cuenta ha sido verificada exitosamente</p>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>

            <p>¡Felicidades! Tu cuenta ha sido verificada y ahora formas parte oficial de la <strong>Asociación de Ayuda Mutua (ASAM)</strong>.</p>

            <p>Estamos encantados de tenerte con nosotros en esta comunidad solidaria.</p>

            <div class="features">
                <h3 style="margin-top: 0;">Con tu cuenta puedes:</h3>
                <div class="feature-item">
                    <strong>📋 Gestionar tu información personal</strong>
                    <p style="margin: 5px 0; color: #666;">Mantén actualizados tus datos de contacto y preferencias.</p>
                </div>
                <div class="feature-item">
                    <strong>💳 Ver el estado de tus pagos</strong>
                    <p style="margin: 5px 0; color: #666;">Consulta tu historial de pagos y próximas cuotas.</p>
                </div>
                <div class="feature-item">
                    <strong>🎁 Acceder a los beneficios de la asociación</strong>
                    <p style="margin: 5px 0; color: #666;">Disfruta de todos los servicios y ayudas disponibles.</p>
                </div>
                <div class="feature-item">
                    <strong>📢 Mantenerte informado</strong>
                    <p style="margin: 5px 0; color: #666;">Recibe noticias y actualizaciones importantes de la asociación.</p>
                </div>
            </div>

            <p><strong>¿Necesitas ayuda?</strong></p>
            <p>Si tienes alguna pregunta o necesitas asistencia, no dudes en contactar con el administrador del sistema. Estamos aquí para ayudarte.</p>

            <p style="margin-top: 30px;">Gracias por unirte a nuestra comunidad solidaria.</p>

            <p><em>- El equipo de ASAM</em></p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
            <p>Este es un correo automático de confirmación.</p>
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

	return a.sendEmail(user.Email, user.Username, subject, buf.String())
}

// SendPasswordChangedEmail sends a notification that password was changed
func (a *mailersendAdapter) SendPasswordChangedEmail(_ context.Context, user *models.User) error {
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
        .header {
            background-color: #ffc107;
            color: #333;
            padding: 20px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            padding: 20px;
            background-color: #ffffff;
            border: 1px solid #e0e0e0;
            border-top: none;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 0.9em;
            color: #666;
            background-color: #f5f5f5;
            border: 1px solid #e0e0e0;
            border-top: none;
            border-radius: 0 0 8px 8px;
        }
        .alert {
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            padding: 15px;
            margin: 20px 0;
            border-radius: 4px;
            color: #721c24;
        }
        .security-tips {
            background-color: #d1ecf1;
            border: 1px solid #bee5eb;
            padding: 15px;
            border-radius: 4px;
            margin: 20px 0;
        }
        .icon {
            font-size: 48px;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="icon">🔐</div>
            <h1 style="margin: 10px 0;">Contraseña Actualizada</h1>
            <p style="margin: 0;">Asociación de Ayuda Mutua (ASAM)</p>
        </div>
        <div class="content">
            <h2>Hola {{.Username}},</h2>

            <p>Te informamos que la contraseña de tu cuenta en ASAM ha sido <strong>actualizada exitosamente</strong>.</p>

            <p><strong>📅 Fecha y hora del cambio:</strong> {{.Timestamp}}</p>

            <div class="alert">
                <strong>⚠️ ¿No reconoces este cambio?</strong>
                <p style="margin: 10px 0;">Si no has realizado este cambio, tu cuenta podría estar comprometida. Por favor:</p>
                <ol style="margin: 5px 0; padding-left: 20px;">
                    <li>Contacta inmediatamente con el administrador del sistema</li>
                    <li>Intenta recuperar tu cuenta usando la opción "Olvidé mi contraseña"</li>
                    <li>Revisa la seguridad de tu correo electrónico</li>
                </ol>
            </div>

            <div class="security-tips">
                <strong>🛡️ Consejos de seguridad:</strong>
                <ul style="margin: 10px 0; padding-left: 20px;">
                    <li>Usa una contraseña única para ASAM (no la uses en otros sitios)</li>
                    <li>Crea contraseñas con al menos 8 caracteres, incluyendo números y símbolos</li>
                    <li>No compartas tu contraseña con nadie, ni siquiera con el personal de ASAM</li>
                    <li>Cambia tu contraseña regularmente (cada 3-6 meses)</li>
                    <li>Ten cuidado con los correos de phishing que soliciten tu contraseña</li>
                </ul>
            </div>

            <p>Si realizaste este cambio, puedes ignorar este mensaje de forma segura.</p>

            <p style="margin-top: 30px;">Gracias por mantener tu cuenta segura.</p>
        </div>
        <div class="footer">
            <p>© 2024 Asociación de Ayuda Mutua (ASAM). Todos los derechos reservados.</p>
            <p>Este es un correo automático de seguridad. Por favor, no respondas a este mensaje.</p>
            <p>Si necesitas ayuda, contacta con el administrador.</p>
        </div>
    </div>
</body>
</html>
`))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]string{
		"Username":  user.Username,
		"Timestamp": time.Now().Format("02/01/2006 15:04:05 MST"),
	})
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "failed to generate email template")
	}

	return a.sendEmail(user.Email, user.Username, subject, buf.String())
}
