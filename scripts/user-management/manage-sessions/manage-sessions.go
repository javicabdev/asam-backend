package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func main() {
	fmt.Println("=== Gestión de Sesiones ASAM ===")

	// Load environment
	envFile := ".env.development"
	if _, err := os.Stat(envFile); err != nil {
		envFile = ".env"
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: Could not load %s: %v\n", envFile, err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Connect to database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		showMenu()
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			viewUserSessions(database, reader)
		case "2":
			revokeSession(database, reader)
		case "3":
			revokeAllUserSessions(database, reader)
		case "4":
			cleanupExpiredTokens(database)
		case "5":
			viewSessionStats(database)
		case "6":
			fmt.Println("¡Hasta luego!")
			return
		default:
			fmt.Println("Opción no válida")
		}
	}
}

func showMenu() {
	fmt.Println("\n¿Qué deseas hacer?")
	fmt.Println("1. Ver sesiones activas de un usuario")
	fmt.Println("2. Revocar una sesión específica")
	fmt.Println("3. Revocar todas las sesiones de un usuario")
	fmt.Println("4. Limpiar tokens expirados")
	fmt.Println("5. Ver estadísticas de sesiones")
	fmt.Println("6. Salir")
	fmt.Print("Selección: ")
}

func viewUserSessions(db *gorm.DB, reader *bufio.Reader) {
	// List users first
	var users []models.User
	db.Find(&users)

	fmt.Println("\n=== Usuarios disponibles ===")
	for _, user := range users {
		fmt.Printf("ID: %d - %s\n", user.ID, user.Username)
	}

	fmt.Print("\nID del usuario: ")
	var userID uint
	_, _ = fmt.Scanf("%d", &userID)
	_, _ = reader.ReadString('\n') // Clear buffer

	// Get active sessions
	var tokens []models.RefreshToken
	now := time.Now().Unix()

	result := db.Where("user_id = ? AND expires_at > ?", userID, now).
		Order("created_at DESC").
		Find(&tokens)

	if result.Error != nil {
		fmt.Printf("❌ Error al obtener sesiones: %v\n", result.Error)
		return
	}

	if len(tokens) == 0 {
		fmt.Println("No hay sesiones activas para este usuario")
		return
	}

	fmt.Printf("\n=== Sesiones activas (%d) ===\n", len(tokens))
	fmt.Printf("%-5s %-20s %-20s %-15s %-20s\n", "ID", "Creada", "Última vez usada", "Dispositivo", "IP")
	fmt.Println(strings.Repeat("-", 80))

	for _, token := range tokens {
		lastUsed := "Nunca"
		if !token.LastUsedAt.IsZero() {
			lastUsed = token.LastUsedAt.Format("2006-01-02 15:04")
		}

		device := token.DeviceName
		if device == "" {
			device = "Desconocido"
		}

		ip := token.IPAddress
		if ip == "" {
			ip = "N/A"
		}

		fmt.Printf("%-5d %-20s %-20s %-15s %-20s\n",
			token.ID,
			token.CreatedAt.Format("2006-01-02 15:04"),
			lastUsed,
			device,
			ip,
		)
	}
}

func revokeSession(db *gorm.DB, reader *bufio.Reader) {
	fmt.Print("\nUUID del token a revocar: ")
	uuid, _ := reader.ReadString('\n')
	uuid = strings.TrimSpace(uuid)

	// Confirm
	fmt.Print("¿Estás seguro? (s/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "s" {
		fmt.Println("Operación cancelada")
		return
	}

	result := db.Where("uuid = ?", uuid).Delete(&models.RefreshToken{})

	if result.Error != nil {
		fmt.Printf("❌ Error al revocar sesión: %v\n", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		fmt.Println("❌ No se encontró la sesión")
		return
	}

	fmt.Println("✅ Sesión revocada exitosamente")
}

func revokeAllUserSessions(db *gorm.DB, reader *bufio.Reader) {
	// List users first
	var users []models.User
	db.Find(&users)

	fmt.Println("\n=== Usuarios disponibles ===")
	for _, user := range users {
		fmt.Printf("ID: %d - %s\n", user.ID, user.Username)
	}

	fmt.Print("\nID del usuario: ")
	var userID uint
	_, _ = fmt.Scanf("%d", &userID)
	_, _ = reader.ReadString('\n') // Clear buffer

	// Count active sessions
	var count int64
	now := time.Now().Unix()
	db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND expires_at > ?", userID, now).
		Count(&count)

	if count == 0 {
		fmt.Println("Este usuario no tiene sesiones activas")
		return
	}

	fmt.Printf("Se revocarán %d sesiones activas\n", count)
	fmt.Print("¿Estás seguro? (s/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "s" {
		fmt.Println("Operación cancelada")
		return
	}

	result := db.Where("user_id = ?", userID).Delete(&models.RefreshToken{})

	if result.Error != nil {
		fmt.Printf("❌ Error al revocar sesiones: %v\n", result.Error)
		return
	}

	fmt.Printf("✅ %d sesiones revocadas exitosamente\n", result.RowsAffected)
}

func cleanupExpiredTokens(db *gorm.DB) {
	now := time.Now().Unix()

	// Count expired tokens first
	var count int64
	db.Model(&models.RefreshToken{}).
		Where("expires_at < ?", now).
		Count(&count)

	if count == 0 {
		fmt.Println("No hay tokens expirados para limpiar")
		return
	}

	fmt.Printf("Se encontraron %d tokens expirados\n", count)

	result := db.Where("expires_at < ?", now).Delete(&models.RefreshToken{})

	if result.Error != nil {
		fmt.Printf("❌ Error al limpiar tokens: %v\n", result.Error)
		return
	}

	fmt.Printf("✅ %d tokens expirados eliminados\n", result.RowsAffected)
}

func viewSessionStats(db *gorm.DB) {
	// Total tokens
	var totalTokens int64
	db.Model(&models.RefreshToken{}).Count(&totalTokens)

	// Active tokens
	var activeTokens int64
	now := time.Now().Unix()
	db.Model(&models.RefreshToken{}).
		Where("expires_at > ?", now).
		Count(&activeTokens)

	// Expired tokens
	expiredTokens := totalTokens - activeTokens

	// Users with active sessions
	var usersWithSessions int64
	db.Model(&models.RefreshToken{}).
		Where("expires_at > ?", now).
		Distinct("user_id").
		Count(&usersWithSessions)

	// Average sessions per user
	avgSessions := float64(0)
	if usersWithSessions > 0 {
		avgSessions = float64(activeTokens) / float64(usersWithSessions)
	}

	fmt.Println("\n=== Estadísticas de Sesiones ===")
	fmt.Printf("Total de tokens en BD: %d\n", totalTokens)
	fmt.Printf("Tokens activos: %d\n", activeTokens)
	fmt.Printf("Tokens expirados: %d\n", expiredTokens)
	fmt.Printf("Usuarios con sesiones activas: %d\n", usersWithSessions)
	fmt.Printf("Promedio de sesiones por usuario: %.2f\n", avgSessions)
}
