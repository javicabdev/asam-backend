package main

import (
	"context"
	"fmt"
	"os"

	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Member struct {
	ID               uint   `gorm:"column:id"`
	MembershipNumber string `gorm:"column:membership_number"`
	Name             string `gorm:"column:name"`
	Surnames         string `gorm:"column:surnames"`
	MembershipType   string `gorm:"column:membership_type"`
	State            string `gorm:"column:state"`
}

func main() {
	dryRun := true
	if len(os.Args) > 1 && os.Args[1] == "--apply" {
		dryRun = false
	}

	// Inicializar logger
	logConfig := logger.Config{
		Level:         logger.ErrorLevel,
		OutputPath:    "stdout",
		Development:   true,
		ConsoleOutput: false,
	}
	_, err := logger.InitLogger(logConfig)
	if err != nil {
		fmt.Printf("Error inicializando logger: %v\n", err)
		os.Exit(1)
	}

	// Cargar configuración
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error cargando configuración: %v\n", err)
		os.Exit(1)
	}

	// Conectar a la base de datos
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error conectando a la base de datos: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if dryRun {
		fmt.Println("🔍 Modo DRY-RUN - Solo identificando problemas (no se harán cambios)")
		fmt.Println("   Para aplicar los cambios, ejecutar con: --apply")
	} else {
		fmt.Println("⚠️  Modo APLICAR - Se harán cambios en la base de datos")
	}
	fmt.Println("==========================================")
	fmt.Println("")

	// Buscar todos los miembros
	var members []Member
	if err := db.WithContext(ctx).Table("members").Find(&members).Error; err != nil {
		fmt.Printf("Error obteniendo miembros: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total de miembros en la base de datos: %d\n", len(members))
	fmt.Println("")

	// Contadores
	var incorrectPrefixes []Member
	var correctCount int

	// Analizar cada miembro
	for _, member := range members {
		prefix := ""
		if len(member.MembershipNumber) > 0 {
			prefix = string(member.MembershipNumber[0])
		}

		expectedPrefix := ""
		if member.MembershipType == "familiar" {
			expectedPrefix = "A"
		} else if member.MembershipType == "individual" {
			expectedPrefix = "B"
		}

		// Verificar si el prefijo es correcto
		if prefix != expectedPrefix && expectedPrefix != "" {
			incorrectPrefixes = append(incorrectPrefixes, member)
		} else {
			correctCount++
		}
	}

	fmt.Printf("✅ Miembros con prefijo correcto: %d\n", correctCount)
	fmt.Printf("❌ Miembros con prefijo incorrecto: %d\n", len(incorrectPrefixes))
	fmt.Println("")

	if len(incorrectPrefixes) == 0 {
		fmt.Println("🎉 ¡Todos los miembros tienen el prefijo correcto!")
		return
	}

	// Mostrar detalles de los miembros con prefijo incorrecto
	fmt.Println("Miembros con prefijo incorrecto:")
	fmt.Println("--------------------------------")
	for _, member := range incorrectPrefixes {
		currentPrefix := string(member.MembershipNumber[0])
		expectedPrefix := "B"
		if member.MembershipType == "familiar" {
			expectedPrefix = "A"
		}

		// Construir nuevo número
		newNumber := expectedPrefix + member.MembershipNumber[1:]

		fmt.Printf("ID: %d\n", member.ID)
		fmt.Printf("  Nombre: %s %s\n", member.Name, member.Surnames)
		fmt.Printf("  Tipo: %s (esperado prefijo: %s)\n", member.MembershipType, expectedPrefix)
		fmt.Printf("  Número actual: %s (prefijo: %s) ❌\n", member.MembershipNumber, currentPrefix)
		fmt.Printf("  Número nuevo:  %s (prefijo: %s) ✅\n", newNumber, expectedPrefix)
		fmt.Println("")
	}

	if dryRun {
		fmt.Println("==========================================")
		fmt.Println("ℹ️  Modo DRY-RUN: No se han realizado cambios")
		fmt.Printf("   Para aplicar estos %d cambios, ejecutar:\n", len(incorrectPrefixes))
		fmt.Println("   go run cmd/fix-member-prefixes/main.go --apply")
		return
	}

	// Aplicar cambios
	fmt.Println("==========================================")
	fmt.Println("🔧 Aplicando correcciones...")
	fmt.Println("")

	successCount := 0
	errorCount := 0

	for _, member := range incorrectPrefixes {
		expectedPrefix := "B"
		if member.MembershipType == "familiar" {
			expectedPrefix = "A"
		}

		newNumber := expectedPrefix + member.MembershipNumber[1:]

		fmt.Printf("Actualizando ID %d: %s -> %s... ", member.ID, member.MembershipNumber, newNumber)

		// Actualizar en la base de datos
		result := db.WithContext(ctx).Table("members").
			Where("id = ?", member.ID).
			Update("membership_number", newNumber)

		if result.Error != nil {
			fmt.Printf("❌ Error: %v\n", result.Error)
			errorCount++
		} else {
			fmt.Printf("✅\n")
			successCount++
		}
	}

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Printf("✅ Actualizaciones exitosas: %d\n", successCount)
	if errorCount > 0 {
		fmt.Printf("❌ Errores: %d\n", errorCount)
	}
	fmt.Println("🎉 Corrección completada!")
}
