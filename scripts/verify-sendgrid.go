package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Script para verificar la API Key de SendGrid
func main() {
	// Cargar .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No se pudo cargar .env, usando variables del sistema")
	}

	apiKey := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM_EMAIL")

	fmt.Println("=== Verificación de SendGrid ===")
	fmt.Printf("API Key: %s\n", maskKey(apiKey))
	fmt.Printf("From Email: %s\n", fromEmail)
	fmt.Println()

	if apiKey == "" || apiKey == "REEMPLAZAR-CON-TU-API-KEY-DE-SENDGRID" {
		fmt.Println("❌ ERROR: No hay API Key configurada en SMTP_PASSWORD")
		os.Exit(1)
	}

	// 1. Verificar API Key
	fmt.Println("1. Verificando API Key...")
	if !verifyAPIKey(apiKey) {
		fmt.Println("❌ API Key inválida o sin permisos")
		fmt.Println("\nSolución:")
		fmt.Println("1. Ve a https://app.sendgrid.com/settings/api_keys")
		fmt.Println("2. Crea una nueva API Key con permisos de 'Mail Send'")
		fmt.Println("3. Actualiza SMTP_PASSWORD en el archivo .env")
		os.Exit(1)
	}
	fmt.Println("✅ API Key válida")

	// 2. Verificar sender
	fmt.Println("\n2. Verificando sender autenticado...")
	if !verifySender(apiKey, fromEmail) {
		fmt.Println("⚠️  El sender podría no estar verificado")
		fmt.Println("\nSolución:")
		fmt.Println("1. Ve a https://app.sendgrid.com/settings/sender_auth")
		fmt.Println("2. Verifica el email:", fromEmail)
	} else {
		fmt.Println("✅ Sender verificado")
	}

	// 3. Obtener estadísticas
	fmt.Println("\n3. Estadísticas de la cuenta...")
	getStats(apiKey)

	fmt.Println("\n✅ Configuración lista para usar!")
}

func maskKey(key string) string {
	if len(key) < 20 {
		return "***"
	}
	return key[:15] + "..."
}

func verifyAPIKey(apiKey string) bool {
	req, _ := http.NewRequest("GET", "https://api.sendgrid.com/v3/scopes", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if scopes, ok := result["scopes"].([]interface{}); ok {
			fmt.Printf("   Permisos: ")
			for i, scope := range scopes {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%v", scope)
			}
			fmt.Println()
		}
		return true
	}

	return false
}

func verifySender(apiKey, fromEmail string) bool {
	req, _ := http.NewRequest("GET", "https://api.sendgrid.com/v3/verified_senders", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if results, ok := result["results"].([]interface{}); ok {
			for _, sender := range results {
				if s, ok := sender.(map[string]interface{}); ok {
					if email, ok := s["from_email"].(string); ok && strings.EqualFold(email, fromEmail) {
						if verified, ok := s["verified"].(bool); ok && verified {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func getStats(apiKey string) {
	// Obtener límites
	req, _ := http.NewRequest("GET", "https://api.sendgrid.com/v3/user/credits", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if remain, ok := result["remain"].(float64); ok {
			fmt.Printf("   Emails restantes este mes: %.0f\n", remain)
		}
		if total, ok := result["total"].(float64); ok {
			fmt.Printf("   Límite mensual: %.0f\n", total)
		}
	}
	resp.Body.Close()
}
