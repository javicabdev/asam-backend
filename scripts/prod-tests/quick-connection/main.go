// Package main provides quick database connection testing
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func main() {
	fmt.Println("=== Prueba rápida de conexión a base de datos ===")
	fmt.Println()

	// Load production environment
	if err := godotenv.Load(".env.production"); err != nil {
		log.Printf("Advertencia: No se pudo cargar .env.production: %v\n", err)
	}

	// Force production environment
	if err := os.Setenv("APP_ENV", "production"); err != nil {
		log.Printf("Advertencia: No se pudo configurar APP_ENV: %v\n", err)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ Error cargando configuración: %v", err)
	}

	fmt.Printf("📍 Conectando a: %s:%s/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
	fmt.Printf("👤 Usuario: %s\n", cfg.DBUser)
	fmt.Printf("🔒 SSL Mode: %s\n", cfg.DBSSLMode)
	fmt.Println()

	// Initialize database connection
	start := time.Now()
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("❌ Error conectando a la base de datos: %v", err)
	}
	connectionTime := time.Since(start)

	fmt.Printf("✅ Conexión exitosa en %v\n", connectionTime)

	// Test database with a simple query
	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("❌ Error obteniendo conexión SQL: %v", err)
	}

	// Ping the database
	start = time.Now()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Error haciendo ping a la base de datos: %v", err)
	}
	pingTime := time.Since(start)

	fmt.Printf("✅ Ping exitoso en %v\n", pingTime)

	// Get some statistics
	fmt.Println()
	fmt.Println("📊 Estadísticas de la base de datos:")

	// Count members
	var memberCount int64
	if err := database.Model(&models.Member{}).Count(&memberCount).Error; err != nil {
		fmt.Printf("   ⚠️  No se pudo contar miembros: %v\n", err)
	} else {
		fmt.Printf("   • Miembros totales: %d\n", memberCount)
	}

	// Count active members
	var activeMemberCount int64
	if err := database.Model(&models.Member{}).Where("state = ?", "active").Count(&activeMemberCount).Error; err != nil {
		fmt.Printf("   ⚠️  No se pudo contar miembros activos: %v\n", err)
	} else {
		fmt.Printf("   • Miembros activos: %d\n", activeMemberCount)
	}

	// Count payments
	var paymentCount int64
	if err := database.Model(&models.Payment{}).Count(&paymentCount).Error; err != nil {
		fmt.Printf("   ⚠️  No se pudo contar pagos: %v\n", err)
	} else {
		fmt.Printf("   • Pagos registrados: %d\n", paymentCount)
	}

	// Get database stats
	stats := sqlDB.Stats()
	fmt.Println()
	fmt.Println("🔧 Estado del pool de conexiones:")
	fmt.Printf("   • Conexiones abiertas: %d\n", stats.OpenConnections)
	fmt.Printf("   • Conexiones en uso: %d\n", stats.InUse)
	fmt.Printf("   • Conexiones idle: %d\n", stats.Idle)
	fmt.Printf("   • Máximo de conexiones: %d\n", cfg.DBMaxOpenConns)

	fmt.Println()
	fmt.Println("✅ ¡Todas las verificaciones pasaron exitosamente!")
	fmt.Println("✅ La base de datos de producción está operativa y accesible.")
}
