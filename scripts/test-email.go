package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

// Script para probar la configuración de SendGrid
// Uso: go run scripts/test-email.go

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar .env, usando variables de entorno del sistema")
	}

	// Obtener configuración
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")

	// Verificar configuración
	fmt.Println("=== Configuración SMTP ===")
	fmt.Printf("Server: %s\n", smtpServer)
	fmt.Printf("Port: %s\n", smtpPort)
	fmt.Printf("User: %s\n", smtpUser)
	fmt.Printf("Password: %s\n", maskPassword(smtpPassword))
	fmt.Printf("From: %s\n", fromEmail)
	fmt.Println()

	if smtpServer == "" || smtpPassword == "" {
		log.Fatal("ERROR: Configuración SMTP incompleta. Verifica tu archivo .env")
	}

	// Email de prueba
	to := "javierfernandezc@gmail.com" // Cambia esto a tu email
	subject := "Test Email - ASAM Backend"
	body := `
Hola,

Este es un email de prueba desde el backend de ASAM.

Si recibes este email, la configuración de SendGrid está funcionando correctamente.

Saludos,
Sistema ASAM
`

	// Formato del mensaje
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s", fromEmail, to, subject, body)

	// Configurar autenticación
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)

	// Enviar email
	fmt.Printf("Enviando email de prueba a %s...\n", to)

	addr := fmt.Sprintf("%s:%s", smtpServer, smtpPort)
	err := smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(message))

	if err != nil {
		log.Fatalf("ERROR al enviar email: %v", err)
	}

	fmt.Println("✅ Email enviado exitosamente!")
	fmt.Println("Revisa tu bandeja de entrada (y la carpeta de SPAM)")
}

func maskPassword(password string) string {
	if len(password) < 10 {
		return "***"
	}
	return password[:10] + "..."
}
