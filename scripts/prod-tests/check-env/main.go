package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=== Verificación de Variables de Entorno ===")
	fmt.Println()

	// Try to load .env.production
	err := godotenv.Load(".env.production")
	if err != nil {
		fmt.Printf("❌ Error cargando .env.production: %v\n", err)
		fmt.Println("Intentando con variables de entorno existentes...")
	} else {
		fmt.Println("✅ Archivo .env.production cargado exitosamente")
	}

	// Force production environment
	if err := os.Setenv("APP_ENV", "production"); err != nil {
		fmt.Printf("❌ Error configurando APP_ENV: %v\n", err)
	}
	if err := os.Setenv("ENVIRONMENT", "production"); err != nil {
		fmt.Printf("❌ Error configurando ENVIRONMENT: %v\n", err)
	}

	fmt.Println()
	fmt.Println("Variables de entorno actuales:")
	fmt.Println("==============================")

	envVars := []string{
		"APP_ENV",
		"ENVIRONMENT",
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_NAME",
		"DB_SSL_MODE",
		"PORT",
	}

	for _, envVar := range envVars {
		value := os.Getenv(envVar)
		if envVar == "DB_PASSWORD" && value != "" {
			fmt.Printf("%s = ****** (oculto)\n", envVar)
		} else if value != "" {
			fmt.Printf("%s = %s\n", envVar, value)
		} else {
			fmt.Printf("%s = (no definido)\n", envVar)
		}
	}

	fmt.Println()

	// Check which environment we're actually using
	dbHost := os.Getenv("DB_HOST")
	switch dbHost {
	case "localhost", "127.0.0.1":
		fmt.Println("⚠️  ADVERTENCIA: Estás usando la base de datos LOCAL")
		fmt.Println("   Para usar producción, asegúrate de que .env.production se cargue correctamente")
	case "pg-asam-asam-backend-db.l.aivencloud.com":
		fmt.Println("✅ Usando la base de datos de PRODUCCIÓN en Aiven")
	default:
		fmt.Printf("❓ Usando base de datos en: %s\n", dbHost)
	}
}
