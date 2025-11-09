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
	Address          string `gorm:"column:address"`
	Postcode         string `gorm:"column:postcode"`
	City             string `gorm:"column:city"`
	Province         string `gorm:"column:province"`
	Country          string `gorm:"column:country"`
}

type Family struct {
	ID              uint   `gorm:"column:id"`
	NumeroSocio     string `gorm:"column:numero_socio"`
	MiembroOrigenID *uint  `gorm:"column:miembro_origen_id"`
	EsposoNombre    string `gorm:"column:esposo_nombre"`
	EsposoApellidos string `gorm:"column:esposo_apellidos"`
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
		fmt.Println("🔍 Modo DRY-RUN - Solo identificando miembros familiares sin familia")
		fmt.Println("   Para aplicar los cambios, ejecutar con: --apply")
	} else {
		fmt.Println("⚠️  Modo APLICAR - Se crearán familias en la base de datos")
	}
	fmt.Println("==========================================")
	fmt.Println("")

	// Buscar miembros de tipo familiar
	var familiarMembers []Member
	if err := db.WithContext(ctx).Table("members").
		Where("membership_type = ?", "familiar").
		Find(&familiarMembers).Error; err != nil {
		fmt.Printf("Error obteniendo miembros familiares: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Total de miembros familiares: %d\n", len(familiarMembers))
	fmt.Println("")

	// Verificar cuáles no tienen familia
	var missingFamilies []Member

	for _, member := range familiarMembers {
		var family Family
		err := db.WithContext(ctx).Table("families").
			Where("numero_socio = ?", member.MembershipNumber).
			First(&family).Error

		if err == gorm.ErrRecordNotFound {
			missingFamilies = append(missingFamilies, member)
		} else if err != nil {
			fmt.Printf("Error verificando familia para %s: %v\n", member.MembershipNumber, err)
		}
	}

	if len(missingFamilies) == 0 {
		fmt.Println("✅ Todos los miembros familiares tienen su familia asociada")
		return
	}

	fmt.Printf("❌ Miembros familiares sin familia: %d\n", len(missingFamilies))
	fmt.Println("")

	// Mostrar detalles
	fmt.Println("Miembros familiares sin familia:")
	fmt.Println("--------------------------------")
	for _, member := range missingFamilies {
		fmt.Printf("ID: %d\n", member.ID)
		fmt.Printf("  Nombre: %s %s\n", member.Name, member.Surnames)
		fmt.Printf("  Número socio: %s\n", member.MembershipNumber)
		fmt.Printf("  Dirección: %s, %s %s\n", member.Address, member.Postcode, member.City)
		fmt.Println("")
	}

	if dryRun {
		fmt.Println("==========================================")
		fmt.Println("ℹ️  Modo DRY-RUN: No se han realizado cambios")
		fmt.Printf("   Para crear estas %d familias, ejecutar:\n", len(missingFamilies))
		fmt.Println("   go run cmd/create-missing-families/main.go --apply")
		return
	}

	// Aplicar cambios
	fmt.Println("==========================================")
	fmt.Println("🔧 Creando familias...")
	fmt.Println("")

	successCount := 0
	errorCount := 0

	for _, member := range missingFamilies {
		fmt.Printf("Creando familia para %s (%s %s)... ", member.MembershipNumber, member.Name, member.Surnames)

		family := Family{
			NumeroSocio:     member.MembershipNumber,
			MiembroOrigenID: &member.ID,
			EsposoNombre:    member.Name,
			EsposoApellidos: member.Surnames,
		}

		result := db.WithContext(ctx).Table("families").Create(&family)

		if result.Error != nil {
			fmt.Printf("❌ Error: %v\n", result.Error)
			errorCount++
		} else {
			fmt.Printf("✅ (ID: %d)\n", family.ID)
			successCount++
		}
	}

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Printf("✅ Familias creadas: %d\n", successCount)
	if errorCount > 0 {
		fmt.Printf("❌ Errores: %d\n", errorCount)
	}
	fmt.Println("🎉 Proceso completado!")
}
