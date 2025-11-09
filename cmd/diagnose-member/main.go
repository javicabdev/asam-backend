package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

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

type Family struct {
	ID              uint   `gorm:"column:id"`
	NumeroSocio     string `gorm:"column:numero_socio"`
	Direccion       string `gorm:"column:direccion"`
	MiembroOrigenID *uint  `gorm:"column:miembro_origen_id"`
}

type Familiar struct {
	ID         uint   `gorm:"column:id"`
	Nombre     string `gorm:"column:nombre"`
	Parentesco string `gorm:"column:parentesco"`
	FamiliaID  uint   `gorm:"column:familia_id"`
}

type Payment struct {
	ID          uint    `gorm:"column:id"`
	Concept     string  `gorm:"column:concept"`
	Amount      float64 `gorm:"column:amount"`
	PaymentDate string  `gorm:"column:payment_date"`
	State       string  `gorm:"column:state"`
	MemberID    *uint   `gorm:"column:member_id"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run cmd/diagnose-member/main.go <member_id>")
		os.Exit(1)
	}

	memberID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: member_id debe ser un número: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	logConfig := logger.Config{
		Level:         logger.ErrorLevel, // Solo errores para no ensuciar output
		OutputPath:    "stdout",
		Development:   true,
		ConsoleOutput: false,
	}
	_, err = logger.InitLogger(logConfig)
	if err != nil {
		fmt.Printf("Error inicializando logger: %v\n", err)
		os.Exit(1)
	}

	// Cargar configuración de DB
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

	fmt.Printf("🔍 Diagnosticando socio ID: %d\n", memberID)
	fmt.Println("==========================================")
	fmt.Println("")

	// 1. Obtener información del socio
	fmt.Println("1. Información del socio:")
	fmt.Println("-------------------------")
	var member Member
	if err := db.WithContext(ctx).Table("members").Where("id = ?", memberID).First(&member).Error; err != nil {
		fmt.Printf("❌ Error obteniendo socio: %v\n", err)
		if err == gorm.ErrRecordNotFound {
			fmt.Println("⚠️  El socio no existe en la base de datos")
		}
		os.Exit(1)
	}

	fmt.Printf("ID: %d\n", member.ID)
	fmt.Printf("Número socio: %s\n", member.MembershipNumber)
	fmt.Printf("Nombre: %s %s\n", member.Name, member.Surnames)
	fmt.Printf("Tipo: %s\n", member.MembershipType)
	fmt.Printf("Estado: %s\n", member.State)
	fmt.Println("")

	// 2. Si es familiar, buscar la familia
	if member.MembershipType == "familiar" {
		fmt.Println("2. Buscando familia asociada:")
		fmt.Println("------------------------------")

		var family Family
		err := db.WithContext(ctx).Table("families").
			Where("numero_socio = ?", member.MembershipNumber).
			First(&family).Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("❌ PROBLEMA ENCONTRADO: El socio es de tipo 'familiar' pero NO tiene familia asociada\n")
				fmt.Printf("   - Número de socio: %s\n", member.MembershipNumber)
				fmt.Printf("   - Este es el origen del error 'familia no encontrada'\n")
				fmt.Println("")
				fmt.Println("Posibles soluciones:")
				fmt.Println("1. Crear una familia con numero_socio = %s", member.MembershipNumber)
				fmt.Println("2. Cambiar el tipo de membresía a 'individual'")
				fmt.Println("3. Verificar si el número de socio es correcto")
			} else {
				fmt.Printf("❌ Error buscando familia: %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("✅ Familia encontrada:\n")
		fmt.Printf("   ID: %d\n", family.ID)
		fmt.Printf("   Número socio: %s\n", family.NumeroSocio)
		fmt.Printf("   Dirección: %s\n", family.Direccion)
		if family.MiembroOrigenID != nil {
			fmt.Printf("   Miembro origen ID: %d\n", *family.MiembroOrigenID)

			// Verificar que el miembro origen existe
			var originMember Member
			if err := db.WithContext(ctx).Table("members").Where("id = ?", *family.MiembroOrigenID).First(&originMember).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					fmt.Printf("   ⚠️  ADVERTENCIA: El miembro origen (ID=%d) no existe\n", *family.MiembroOrigenID)
				}
			} else {
				fmt.Printf("   Miembro origen: %s %s (%s)\n", originMember.Name, originMember.Surnames, originMember.MembershipNumber)
			}
		} else {
			fmt.Println("   ⚠️  Miembro origen ID: NULL")
		}
		fmt.Println("")

		// 3. Buscar familiares
		fmt.Println("3. Familiares asociados:")
		fmt.Println("------------------------")
		var familiares []Familiar
		if err := db.WithContext(ctx).Table("familiares").Where("familia_id = ?", family.ID).Find(&familiares).Error; err != nil {
			fmt.Printf("Error buscando familiares: %v\n", err)
		} else {
			if len(familiares) == 0 {
				fmt.Println("   ℹ️  No hay familiares registrados")
			} else {
				for _, f := range familiares {
					fmt.Printf("   - %s (%s)\n", f.Nombre, f.Parentesco)
				}
			}
		}
		fmt.Println("")
	} else {
		fmt.Printf("ℹ️  Este socio es de tipo '%s', no debería tener familia asociada\n", member.MembershipType)
		fmt.Println("")
	}

	// 4. Ver últimos pagos
	fmt.Println("4. Últimos 5 pagos del socio:")
	fmt.Println("-----------------------------")
	var payments []Payment
	if err := db.WithContext(ctx).Table("payments").
		Where("member_id = ?", memberID).
		Order("payment_date DESC").
		Limit(5).
		Find(&payments).Error; err != nil {
		fmt.Printf("Error obteniendo pagos: %v\n", err)
	} else {
		if len(payments) == 0 {
			fmt.Println("   ℹ️  No hay pagos registrados")
		} else {
			for _, p := range payments {
				fmt.Printf("   - %s: %.2f€ (%s) - Estado: %s\n", p.PaymentDate, p.Amount, p.Concept, p.State)
			}
		}
	}

	fmt.Println("")
	fmt.Println("==========================================")
	fmt.Println("✅ Diagnóstico completado")
}
