package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// SMTPConfig contiene la configuración para el servicio SMTP
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	UseTLS   bool
}

// SMTPEmailService es una implementación real del servicio de email usando SMTP
type SMTPEmailService struct {
	config SMTPConfig
	logger logger.Logger
}

// NewSMTPEmailService crea una nueva instancia del servicio de email SMTP
func NewSMTPEmailService(config SMTPConfig, logger logger.Logger) output.EmailService {
	return &SMTPEmailService{
		config: config,
		logger: logger,
	}
}

// SendEmail envía un email de texto plano
func (s *SMTPEmailService) SendEmail(_ context.Context, to, subject, body string) error {
	// Construir el mensaje
	message := s.buildMessage(to, subject, body, false)

	// Enviar el email
	return s.sendMail(to, message)
}

// SendHTMLEmail envía un email HTML
func (s *SMTPEmailService) SendHTMLEmail(_ context.Context, to, subject, htmlBody string) error {
	// Construir el mensaje HTML
	message := s.buildMessage(to, subject, htmlBody, true)

	// Enviar el email
	return s.sendMail(to, message)
}

// SendVerificationEmail envía un email de verificación
func (s *SMTPEmailService) SendVerificationEmail(ctx context.Context, to, username, verificationURL string) error {
	subject := "Verifica tu cuenta en ASAM"

	// Versión HTML
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f8f9fa; }
        .button { display: inline-block; padding: 10px 20px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ASAM - Asociación</h1>
        </div>
        <div class="content">
            <h2>Hola %s,</h2>
            <p>Por favor verifica tu cuenta haciendo clic en el siguiente botón:</p>
            <center>
                <a href="%s" class="button">Verificar mi cuenta</a>
            </center>
            <p>O copia y pega este enlace en tu navegador:</p>
            <p style="word-break: break-all;">%s</p>
            <p><strong>Este enlace expirará en 24 horas.</strong></p>
            <p>Si no solicitaste esta verificación, puedes ignorar este email.</p>
        </div>
        <div class="footer">
            <p>© 2024 ASAM. Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`, username, verificationURL, verificationURL)

	return s.SendHTMLEmail(ctx, to, subject, htmlBody)
}

// SendPasswordResetEmail envía un email de recuperación de contraseña
func (s *SMTPEmailService) SendPasswordResetEmail(ctx context.Context, to, username, resetURL string) error {
	subject := "Recuperación de contraseña - ASAM"

	// Versión HTML
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #dc3545; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f8f9fa; }
        .button { display: inline-block; padding: 10px 20px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .warning { background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 10px; margin: 10px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ASAM - Recuperación de Contraseña</h1>
        </div>
        <div class="content">
            <h2>Hola %s,</h2>
            <p>Hemos recibido una solicitud para restablecer tu contraseña.</p>
            <p>Para continuar, haz clic en el siguiente botón:</p>
            <center>
                <a href="%s" class="button">Restablecer mi contraseña</a>
            </center>
            <p>O copia y pega este enlace en tu navegador:</p>
            <p style="word-break: break-all;">%s</p>
            <div class="warning">
                <strong>⚠️ Este enlace expirará en 1 hora.</strong>
            </div>
            <p>Si no solicitaste este cambio, puedes ignorar este email. Tu contraseña no será modificada.</p>
        </div>
        <div class="footer">
            <p>© 2024 ASAM. Todos los derechos reservados.</p>
        </div>
    </div>
</body>
</html>
`, username, resetURL, resetURL)

	return s.SendHTMLEmail(ctx, to, subject, htmlBody)
}

// SendPasswordChangedEmail envía una notificación de cambio de contraseña
func (s *SMTPEmailService) SendPasswordChangedEmail(ctx context.Context, to, username string) error {
	subject := "Tu contraseña ha sido cambiada - ASAM"

	// Versión HTML
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #28a745; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f8f9fa; }
        .footer { padding: 20px; text-align: center; font-size: 12px; color: #666; }
        .alert { background-color: #d4edda; border: 1px solid #c3e6cb; padding: 10px; margin: 10px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ASAM - Notificación de Seguridad</h1>
        </div>
        <div class="content">
            <h2>Hola %s,</h2>
            <div class="alert">
                <strong>✓ Tu contraseña ha sido cambiada exitosamente.</strong>
            </div>
            <p>Te informamos que tu contraseña ha sido modificada recientemente.</p>
            <p><strong>Si realizaste este cambio:</strong> No necesitas hacer nada más.</p>
            <p><strong>Si NO realizaste este cambio:</strong> Por favor contacta con soporte inmediatamente respondiendo a este email.</p>
            <hr>
            <p>Detalles de seguridad:</p>
            <ul>
                <li>Fecha y hora: %s</li>
                <li>Si no reconoces esta actividad, tu cuenta podría estar comprometida.</li>
            </ul>
        </div>
        <div class="footer">
            <p>© 2024 ASAM. Todos los derechos reservados.</p>
            <p>Este es un mensaje automático de seguridad.</p>
        </div>
    </div>
</body>
</html>
`, username, time.Now().Format("02/01/2006 15:04:05"))

	return s.SendHTMLEmail(ctx, to, subject, htmlBody)
}

// buildMessage construye el mensaje de email con headers
func (s *SMTPEmailService) buildMessage(to, subject, body string, isHTML bool) []byte {
	headers := make(map[string]string)
	headers["From"] = s.config.From
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"

	if isHTML {
		headers["Content-Type"] = "text/html; charset=UTF-8"
	} else {
		headers["Content-Type"] = "text/plain; charset=UTF-8"
	}

	// Construir el mensaje
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return []byte(message)
}

// sendMail envía el email usando SMTP
func (s *SMTPEmailService) sendMail(to string, message []byte) error {
	// Configurar autenticación
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Dirección del servidor
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Para servidores que requieren TLS
	if s.config.UseTLS {
		return s.sendMailTLS(addr, auth, s.config.From, []string{to}, message)
	}

	// Enviar email
	err := smtp.SendMail(addr, auth, s.config.From, []string{to}, message)
	if err != nil {
		s.logger.Error("Failed to send email",
			zap.String("to", to),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully",
		zap.String("to", to),
	)

	return nil
}

// sendMailTLS envía email usando TLS
func (s *SMTPEmailService) sendMailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Conectar al servidor
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName: s.config.Host,
		MinVersion: tls.VersionTLS12, // Enforce minimum TLS 1.2
	})
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Warn("Failed to close TLS connection", zap.Error(err))
		}
	}()

	// Crear cliente SMTP
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			s.logger.Warn("Failed to close SMTP client", zap.Error(err))
		}
	}()

	// Autenticar
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Establecer remitente
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Establecer destinatarios
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Enviar el mensaje
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
