package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

// Test simple de envío de email con gomail
func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		log.Println("No se pudo cargar .env")
	}

	// Configuración
	host := os.Getenv("SMTP_SERVER")
	port := 587 // SendGrid usa 587
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM_EMAIL")
	to := "javierfernandezc@gmail.com"

	fmt.Println("=== Test de Email con SendGrid ===")
	fmt.Printf("Host: %s:%d\n", host, port)
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Pass: %s\n", maskPassword(pass))
	fmt.Printf("From: %s\n", from)
	fmt.Printf("To: %s\n", to)
	fmt.Println()

	// Crear mensaje
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Test ASAM - Verificación SendGrid")
	m.SetBody("text/html", `
		<h2>Test de SendGrid</h2>
		<p>Si recibes este email, la configuración está correcta.</p>
		<p>Este es un email de prueba del sistema ASAM.</p>
	`)

	// Crear dialer
	d := gomail.NewDialer(host, port, user, pass)

	// Enviar
	fmt.Println("Enviando email...")
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("\n❌ ERROR: %v\n", err)
		fmt.Println("\nPosibles causas:")
		fmt.Println("1. API Key inválida o expirada")
		fmt.Println("2. Sender no verificado en SendGrid")
		fmt.Println("3. Límite de emails alcanzado")
		fmt.Println("\nRevisa en https://app.sendgrid.com/activity")
		return
	}

	fmt.Println("\n✅ Email enviado exitosamente!")
	fmt.Println("Revisa tu bandeja de entrada y SPAM")
}

func maskPassword(password string) string {
	if len(password) < 10 {
		return "***"
	}
	return password[:10] + "..."
}
