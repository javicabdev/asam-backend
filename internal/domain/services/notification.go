package services

import (
	"context"
	"regexp"

	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Expresiones regulares compiladas para validación
var (
	// RFC 5322 compliant email regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// Spanish phone number format (including international format)
	// Accepts: +34612345678, 0034612345678, 612345678
	phoneRegex = regexp.MustCompile(`^(?:(?:\+|00)?34)?[6789]\d{8}$`)
)

// emailNotificationService implementa NotificationService para enviar emails
type emailNotificationService struct {
	// Configuraciones necesarias para el servicio de email
	smtpServer   string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	useTLS       bool
	fromAddress  string
}

// NewEmailNotificationService crea una nueva instancia del servicio de notificaciones por email
func NewEmailNotificationService(
	smtpServer string,
	smtpPort int,
	smtpUser string,
	smtpPassword string,
	useTLS bool,
	fromAddress string,
) input.NotificationService {
	// Validación de parámetros obligatorios
	if smtpServer == "" {
		panic("SMTP server cannot be empty")
	}

	if smtpPort <= 0 || smtpPort > 65535 {
		panic("Invalid SMTP port number")
	}

	if smtpUser == "" || smtpPassword == "" {
		panic("SMTP credentials cannot be empty")
	}

	if fromAddress == "" || !emailRegex.MatchString(fromAddress) {
		panic("Invalid sender email address")
	}

	return &emailNotificationService{
		smtpServer:   smtpServer,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		useTLS:       useTLS,
		fromAddress:  fromAddress,
	}
}

// SendEmail envía un correo electrónico al destinatario especificado
func (s *emailNotificationService) SendEmail(_ context.Context, to string, subject string, body string) error {
	// Validar parámetros
	if to == "" {
		return errors.Validation("Email address is required", "to", "required")
	}

	if subject == "" {
		return errors.Validation("Email subject is required", "subject", "required")
	}

	if body == "" {
		return errors.Validation("Email body is required", "body", "required")
	}

	// Validar formato de email con regex
	if !emailRegex.MatchString(to) {
		return errors.Validation("Invalid email format", "to", "invalid_format")
	}

	// Aquí iría la implementación real del envío de email usando el paquete "net/smtp"
	// Por ejemplo:
	/*
		auth := smtp.PlainAuth("", s.smtpUser, s.smtpPassword, s.smtpServer)

		msg := fmt.Sprintf("From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n", s.fromAddress, to, subject, body)

		addr := fmt.Sprintf("%s:%d", s.smtpServer, s.smtpPort)

		err := smtp.SendMail(addr, auth, s.fromAddress, []string{to}, []byte(msg))
		if err != nil {
			// Clasificar el error según su tipo
			if strings.Contains(err.Error(), "connection refused") ||
			   strings.Contains(err.Error(), "no such host") {
				return errors.NetworkError("Failed to connect to SMTP server", err)
			}
			if strings.Contains(err.Error(), "authentication failed") {
				return errors.AuthError(errors.ErrUnauthorized, "SMTP authentication failed", err)
			}
			// Error genérico
			return errors.Wrap(err, errors.ErrInternalError, "Failed to send email")
		}
	*/

	// Por ahora retornamos nil como placeholder
	return nil
}

// SendSMS envía un mensaje SMS al número especificado
func (s *emailNotificationService) SendSMS(_ context.Context, to string, message string) error {
	// Validar parámetros
	if to == "" {
		return errors.Validation("Phone number is required", "to", "required")
	}

	if message == "" {
		return errors.Validation("SMS message is required", "message", "required")
	}

	// Validar longitud del mensaje (típicamente 160 caracteres para SMS estándar)
	if len(message) > 160 {
		return errors.Validation(
			"SMS message exceeds maximum length of 160 characters",
			"message",
			"too_long",
		)
	}

	// Validar formato de teléfono con regex (formato español)
	if !phoneRegex.MatchString(to) {
		return errors.Validation("Invalid phone number format", "to", "invalid_format")
	}

	// Aquí iría la implementación real del envío de SMS a través de una API externa
	// Por ejemplo:
	/*
		smsRequest := &smsGatewayRequest{
			To:      to,
			Message: message,
			APIKey:  s.apiKey,
		}

		jsonData, err := json.Marshal(smsRequest)
		if err != nil {
			return errors.Wrap(err, errors.ErrInternalError, "Failed to serialize SMS request")
		}

		resp, err := http.Post(s.smsGatewayURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return errors.NetworkError("Failed to connect to SMS gateway", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.New(
				errors.ErrInternalError,
				fmt.Sprintf("SMS gateway returned error: %s - %s", resp.Status, string(body)),
			)
		}
	*/

	// Por ahora retornamos nil como placeholder
	return nil
}
