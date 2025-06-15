package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func main() {
	fmt.Println("=== Gestión Manual de Usuarios ASAM ===")

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
			createUser(database, reader)
		case "2":
			listUsers(database)
		case "3":
			deleteUser(database, reader)
		case "4":
			resetPassword(database, reader)
		case "5":
			fmt.Println("¡Hasta luego!")
			return
		default:
			fmt.Println("Opción no válida")
		}
	}
}

func showMenu() {
	fmt.Println("\n¿Qué deseas hacer?")
	fmt.Println("1. Crear nuevo usuario")
	fmt.Println("2. Listar usuarios")
	fmt.Println("3. Eliminar usuario")
	fmt.Println("4. Resetear contraseña")
	fmt.Println("5. Salir")
	fmt.Print("Selección: ")
}

func createUser(db *gorm.DB, reader *bufio.Reader) {
	fmt.Print("\nEmail del usuario: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Contraseña: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	fmt.Print("Rol (admin/user) [user]: ")
	roleStr, _ := reader.ReadString('\n')
	roleStr = strings.TrimSpace(roleStr)
	if roleStr == "" {
		roleStr = "user"
	}

	role := models.RoleUser
	if roleStr == "admin" {
		role = models.RoleAdmin
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("username = ?", email).First(&existingUser).Error; err == nil {
		fmt.Printf("❌ El usuario %s ya existe\n", email)
		return
	}

	// Create new user
	user := models.User{
		Username: email,
		Role:     role,
		IsActive: true,
	}

	if err := user.SetPassword(password); err != nil {
		fmt.Printf("❌ Error al establecer contraseña: %v\n", err)
		return
	}

	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("❌ Error al crear usuario: %v\n", err)
		return
	}

	fmt.Printf("✅ Usuario %s creado exitosamente\n", email)
}

func listUsers(db *gorm.DB) {
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		fmt.Printf("❌ Error al listar usuarios: %v\n", err)
		return
	}

	fmt.Println("\n=== Usuarios registrados ===")
	fmt.Printf("%-5s %-30s %-10s %-10s\n", "ID", "Email", "Rol", "Activo")
	fmt.Println(strings.Repeat("-", 60))

	for _, user := range users {
		activeStr := "No"
		if user.IsActive {
			activeStr = "Sí"
		}
		fmt.Printf("%-5d %-30s %-10s %-10s\n", user.ID, user.Username, user.Role, activeStr)
	}
}

func deleteUser(db *gorm.DB, reader *bufio.Reader) {
	listUsers(db)

	fmt.Print("\nID del usuario a eliminar: ")
	var id uint
	fmt.Scanf("%d", &id)
	reader.ReadString('\n') // Clear buffer

	// Confirm deletion
	fmt.Print("¿Estás seguro? (s/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "s" {
		fmt.Println("Operación cancelada")
		return
	}

	if err := db.Delete(&models.User{}, id).Error; err != nil {
		fmt.Printf("❌ Error al eliminar usuario: %v\n", err)
		return
	}

	fmt.Println("✅ Usuario eliminado exitosamente")
}

func resetPassword(db *gorm.DB, reader *bufio.Reader) {
	fmt.Print("\nEmail del usuario: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Nueva contraseña: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	var user models.User
	if err := db.Where("username = ?", email).First(&user).Error; err != nil {
		fmt.Printf("❌ Usuario no encontrado: %s\n", email)
		return
	}

	if err := user.SetPassword(password); err != nil {
		fmt.Printf("❌ Error al establecer contraseña: %v\n", err)
		return
	}

	if err := db.Save(&user).Error; err != nil {
		fmt.Printf("❌ Error al actualizar usuario: %v\n", err)
		return
	}

	fmt.Printf("✅ Contraseña actualizada para %s\n", email)
}
