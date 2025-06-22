// Package main provides user management functionality
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/term"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func main() {
	fmt.Println("=== Gestión de Usuarios ASAM ===")
	fmt.Println()

	// Load environment
	envFile := selectEnvironment()
	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Advertencia: No se pudo cargar %s: %v\n", envFile, err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error cargando configuración: %v", err)
	}

	// Connect to database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}

	// Show menu
	for {
		showMenu()
		choice := readInput("Seleccione una opción: ")

		switch choice {
		case "1":
			createUser(database)
		case "2":
			listUsers(database)
		case "3":
			changeUserStatus(database)
		case "4":
			changeUserPassword(database)
		case "5":
			fmt.Println("¡Hasta luego!")
			return
		default:
			fmt.Println("Opción no válida")
		}
		fmt.Println()
	}
}

func selectEnvironment() string {
	fmt.Println("Seleccione el entorno:")
	fmt.Println("1. Desarrollo (.env.development)")
	fmt.Println("2. Producción (.env.production.test)")

	choice := readInput("Opción (1-2): ")

	switch choice {
	case "1":
		return ".env.development"
	case "2":
		return ".env.production.test"
	default:
		fmt.Println("Opción no válida, usando desarrollo por defecto")
		return ".env.development"
	}
}

func showMenu() {
	fmt.Println("\n--- MENÚ PRINCIPAL ---")
	fmt.Println("1. Crear nuevo usuario")
	fmt.Println("2. Listar usuarios")
	fmt.Println("3. Activar/Desactivar usuario")
	fmt.Println("4. Cambiar contraseña")
	fmt.Println("5. Salir")
	fmt.Println()
}

func createUser(database *gorm.DB) {
	fmt.Println("\n--- CREAR NUEVO USUARIO ---")

	// Get username
	username := readInput("Nombre de usuario: ")

	// Check if user already exists
	var existingUser models.User
	if err := database.Where("username = ?", username).First(&existingUser).Error; err == nil {
		fmt.Printf("Error: El usuario '%s' ya existe\n", username)
		return
	}

	// Get password
	password := readPassword("Contraseña: ")
	confirmPassword := readPassword("Confirmar contraseña: ")

	if password != confirmPassword {
		fmt.Println("Error: Las contraseñas no coinciden")
		return
	}

	// Get role
	fmt.Println("Seleccione el rol:")
	fmt.Println("1. Usuario normal (user)")
	fmt.Println("2. Administrador (admin)")
	roleChoice := readInput("Opción (1-2): ")

	var role models.Role
	switch roleChoice {
	case "2":
		role = models.RoleAdmin
	default:
		role = models.RoleUser
	}

	// Create user
	user := &models.User{
		Username: username,
		Role:     role,
		IsActive: true,
	}

	// Hash password
	if err := user.SetPassword(password); err != nil {
		fmt.Printf("Error hasheando contraseña: %v\n", err)
		return
	}

	// Save to database
	if err := database.Create(user).Error; err != nil {
		fmt.Printf("Error creando usuario: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Usuario '%s' creado exitosamente con rol '%s'\n", username, role)
}

func listUsers(database *gorm.DB) {
	fmt.Println("\n--- LISTA DE USUARIOS ---")

	var users []models.User
	if err := database.Order("created_at DESC").Find(&users).Error; err != nil {
		fmt.Printf("Error obteniendo usuarios: %v\n", err)
		return
	}

	if len(users) == 0 {
		fmt.Println("No hay usuarios registrados")
		return
	}

	fmt.Printf("\n%-5s %-20s %-10s %-10s %-20s\n", "ID", "Usuario", "Rol", "Estado", "Último Login")
	fmt.Println(strings.Repeat("-", 70))

	for _, user := range users {
		status := "Activo"
		if !user.IsActive {
			status = "Inactivo"
		}

		lastLogin := "Nunca"
		if !user.LastLogin.IsZero() {
			lastLogin = user.LastLogin.Format("02/01/2006 15:04")
		}

		fmt.Printf("%-5d %-20s %-10s %-10s %-20s\n",
			user.ID, user.Username, user.Role, status, lastLogin)
	}
}

func changeUserStatus(database *gorm.DB) {
	fmt.Println("\n--- ACTIVAR/DESACTIVAR USUARIO ---")

	username := readInput("Nombre de usuario: ")

	var user models.User
	if err := database.Where("username = ?", username).First(&user).Error; err != nil {
		fmt.Printf("Error: Usuario '%s' no encontrado\n", username)
		return
	}

	currentStatus := "activo"
	if !user.IsActive {
		currentStatus = "inactivo"
	}

	fmt.Printf("Estado actual: %s\n", currentStatus)
	fmt.Println("1. Activar usuario")
	fmt.Println("2. Desactivar usuario")
	choice := readInput("Opción (1-2): ")

	switch choice {
	case "1":
		user.IsActive = true
	case "2":
		user.IsActive = false
	default:
		fmt.Println("Opción no válida")
		return
	}

	if err := database.Save(&user).Error; err != nil {
		fmt.Printf("Error actualizando usuario: %v\n", err)
		return
	}

	newStatus := "activo"
	if !user.IsActive {
		newStatus = "inactivo"
	}

	fmt.Printf("\n✓ Usuario '%s' ahora está %s\n", username, newStatus)
}

func changeUserPassword(database *gorm.DB) {
	fmt.Println("\n--- CAMBIAR CONTRASEÑA ---")

	username := readInput("Nombre de usuario: ")

	var user models.User
	if err := database.Where("username = ?", username).First(&user).Error; err != nil {
		fmt.Printf("Error: Usuario '%s' no encontrado\n", username)
		return
	}

	// Get new password
	newPassword := readPassword("Nueva contraseña: ")
	confirmPassword := readPassword("Confirmar nueva contraseña: ")

	if newPassword != confirmPassword {
		fmt.Println("Error: Las contraseñas no coinciden")
		return
	}

	// Hash and update password
	if err := user.SetPassword(newPassword); err != nil {
		fmt.Printf("Error hasheando contraseña: %v\n", err)
		return
	}

	if err := database.Save(&user).Error; err != nil {
		fmt.Printf("Error actualizando contraseña: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Contraseña actualizada exitosamente para '%s'\n", username)
}

// Helper functions
func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // New line after password input
	if err != nil {
		log.Fatalf("Error leyendo contraseña: %v", err)
	}
	return string(password)
}
