package main

import (
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

	fmt.Println("🔧 SendGrid SMTP Configuration:")
	fmt.Printf("   Server: %s:%s\n", smtpHost, smtpPort)
	fmt.Printf("   User: %s\n", smtpUser)
	fmt.Printf("   From: %s\n", smtpFrom)
	fmt.Printf("   API Key: %s\n", strings.Repeat("*", len(smtpPassword)-4)+smtpPassword[len(smtpPassword)-4:])
	fmt.Println()

	// Create message
	subject := "Test ASAM - SendGrid Configuration"
	body := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #28a745;">✅ ¡SendGrid Configurado Exitosamente!</h2>
        <p>Este es un email de prueba enviado desde el backend de ASAM usando SendGrid.</p>
        <p>Si estás recibiendo este email, significa que:</p>
        <ul>
            <li>La API Key de SendGrid está funcionando correctamente</li>
            <li>El dominio mutuaasam.org está verificado</li>
            <li>Tu configuración SMTP está lista para producción</li>
        </ul>
        <hr style="border: none; border-top: 1px solid #ccc; margin: 20px 0;">
        <p style="font-size: 12px; color: #666;">
            Enviado desde: ASAM Backend<br>
            Servidor SMTP: SendGrid<br>
            Dominio: mutuaasam.org
        </p>
    </div>
</body>
</html>
`

	message := buildMessage(smtpFrom, testRecipient, subject, body)

	// Test connection
	fmt.Println("📡 Testing SendGrid SMTP connection...")

	// SendGrid uses STARTTLS on port 587
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)

	err := smtp.SendMail(addr, auth, smtpFrom, []string{testRecipient}, message)
	if err != nil {
		fmt.Printf("❌ Failed to send email: %v\n", err)
		fmt.Println("\n❓ Possible issues:")
		fmt.Println("   - Check if the API key is correct (starts with SG.)")
		fmt.Println("   - Ensure the API key has 'Mail Send' permission")
		fmt.Println("   - Verify your SendGrid account is active")
		fmt.Println("   - Check if domain authentication is complete")
		return
	}

	fmt.Printf("✅ Email sent successfully to %s!\n", testRecipient)
	fmt.Println("\n📧 Next steps:")
	fmt.Println("   1. Check your inbox for the test email")
	fmt.Println("   2. If received, your SMTP configuration is ready")
	fmt.Println("   3. Restart your ASAM backend to use the new settings")
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
