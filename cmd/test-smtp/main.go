package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get SMTP configuration
	smtpHost := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM_EMAIL")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		log.Fatal("Missing SMTP configuration. Please check your .env file")
	}

	// Test recipient - change this to your email
	testRecipient := "javierfernandezc@gmail.com"

	fmt.Println("🔧 SMTP Configuration:")
	fmt.Printf("   Server: %s:%s\n", smtpHost, smtpPort)
	fmt.Printf("   User: %s\n", smtpUser)
	fmt.Printf("   From: %s\n", smtpFrom)
	fmt.Printf("   Password: %s\n", strings.Repeat("*", len(smtpPassword)-2)+smtpPassword[len(smtpPassword)-2:])
	fmt.Println()

	// Create message
	subject := "Test ASAM - Configuración SMTP Exitosa"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #28a745;">✅ ¡Configuración SMTP Exitosa!</h2>
        <p>Este es un email de prueba enviado desde el backend de ASAM usando SendGrid.</p>
        <p>Si estás recibiendo este email, significa que tu configuración SMTP está funcionando correctamente.</p>
        <hr style="border: none; border-top: 1px solid #ccc; margin: 20px 0;">
        <p style="font-size: 12px; color: #666;">
            Enviado desde: ASAM Backend<br>
            Servidor SMTP: smtp.sendgrid.net
        </p>
    </div>
</body>
</html>
`

	message := buildMessage(smtpFrom, testRecipient, subject, body)

	// Test connection
	fmt.Println("📡 Testing SMTP connection...")

	// Method 1: Try with TLS first
	if err := sendMailTLS(smtpHost, smtpPort, smtpUser, smtpPassword, smtpFrom, testRecipient, message); err != nil {
		fmt.Printf("❌ TLS method failed: %v\n", err)

		// Method 2: Try with STARTTLS
		fmt.Println("\n🔄 Trying alternative method (STARTTLS)...")
		if err := sendMailSTARTTLS(smtpHost, smtpPort, smtpUser, smtpPassword, smtpFrom, testRecipient, message); err != nil {
			fmt.Printf("❌ STARTTLS method failed: %v\n", err)
			fmt.Println("\n❓ Possible issues:")
			fmt.Println("   - Check if the API key is correct")
			fmt.Println("   - Ensure the API key has 'Mail Send' permission")
			fmt.Println("   - Verify your SendGrid account is active")
			fmt.Println("   - Check if domain authentication is complete")
		} else {
			fmt.Printf("✅ Email sent successfully to %s using STARTTLS!\n", testRecipient)
		}
	} else {
		fmt.Printf("✅ Email sent successfully to %s using TLS!\n", testRecipient)
	}
}

func buildMessage(from, to, subject, body string) []byte {
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return []byte(message)
}

func sendMailTLS(host, port, username, password, from, to string, message []byte) error {
	// Connect with TLS
	addr := fmt.Sprintf("%s:%s", host, port)

	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", username, password, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Send email
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

func sendMailSTARTTLS(host, port, username, password, from, to string, message []byte) error {
	// Standard SMTP with STARTTLS
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", username, password, host)

	// Send using standard SendMail (uses STARTTLS automatically on port 587)
	return smtp.SendMail(addr, auth, from, []string{to}, message)
}
