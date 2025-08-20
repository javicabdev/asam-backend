package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error cargando .env:", err)
		return
	}

	fmt.Println("=== Configuración Actual ===")
	fmt.Println()

	// Mostrar todas las variables SMTP
	envVars := []string{
		"SMTP_SERVER",
		"SMTP_PORT",
		"SMTP_USER",
		"SMTP_PASSWORD",
		"SMTP_USE_TLS",
		"SMTP_FROM_EMAIL",
		"SMTP_FROM_NAME",
	}

	for _, key := range envVars {
		value := os.Getenv(key)
		if key == "SMTP_PASSWORD" && value != "" {
			// Ocultar parte del password
			if len(value) > 20 {
				value = value[:20] + "..."
			}
		}
		fmt.Printf("%-20s: %s\n", key, value)
	}

	fmt.Println()
	fmt.Println("=== Verificaciones ===")

	// Verificar que no sea el valor por defecto
	if os.Getenv("SMTP_PASSWORD") == "REEMPLAZAR-CON-TU-API-KEY-DE-SENDGRID" {
		fmt.Println("❌ ERROR: La API Key no ha sido configurada")
		return
	}

	// Verificar formato de API key
	apiKey := os.Getenv("SMTP_PASSWORD")
	if !strings.HasPrefix(apiKey, "SG.") {
		fmt.Println("⚠️  ADVERTENCIA: La API Key no parece tener el formato correcto de SendGrid (debe empezar con 'SG.')")
	}

	// Verificar que el FROM email no sea el problemático
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")
	if fromEmail == "noreply@asam.org" {
		fmt.Println("❌ ERROR: El FROM email sigue siendo 'noreply@asam.org' que NO está verificado")
		fmt.Println("   Debes cambiarlo a 'admin@mutuaasam.org' o verificar el sender")
		return
	}

	fmt.Println("✅ Configuración parece correcta")
	fmt.Println()
	fmt.Println("Si sigues teniendo errores, verifica:")
	fmt.Println("1. Que reiniciaste el backend después de cambiar el .env")
	fmt.Println("2. Que el email", fromEmail, "esté verificado en SendGrid")
	fmt.Println("3. Los logs detallados del backend")
}
