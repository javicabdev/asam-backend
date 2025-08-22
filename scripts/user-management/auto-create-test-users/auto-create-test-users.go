// Package main provides automated test user creation functionality
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/adapters/db"
	"github.com/javicabdev/asam-backend/internal/config"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// This script automatically creates test users without user interaction
// It's used by start-docker.ps1 to initialize the system

func main() {
	fmt.Println("=== Auto-creating Test Users with Members ===")

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

	// Track success
	adminCreated := false
	userCreated := false

	// Create admin user (sin miembro asociado)
	if err := createOrUpdateAdminUser(database); err != nil {
		log.Printf("❌ Error creating admin user: %v", err)
	} else {
		fmt.Println("✓ Admin user ready")
		adminCreated = true
	}

	// Create regular user with associated member
	if err := createOrUpdateUserWithMember(database); err != nil {
		log.Printf("❌ Error creating user with member: %v", err)
	} else {
		fmt.Println("✓ Regular user with member ready")
		userCreated = true
	}

	// Verify final state
	var userCount int64
	database.Model(&models.User{}).Count(&userCount)
	fmt.Printf("\n📊 Final user count in database: %d\n", userCount)

	// List all users
	var users []models.User
	if err := database.Select("id", "username", "email", "role", "member_id").Find(&users).Error; err == nil {
		fmt.Println("\n📋 Users in database:")
		for _, u := range users {
			memberInfo := "no member"
			if u.MemberID != nil {
				memberInfo = fmt.Sprintf("member_id=%d", *u.MemberID)
			}
			fmt.Printf("   - ID:%d | %s | %s | role=%s | %s\n",
				u.ID, u.Username, u.Email, u.Role, memberInfo)
		}
	}

	if adminCreated && userCreated {
		fmt.Println("\n✅ Test users created successfully!")
		fmt.Println("You can login with:")
		fmt.Println("  Admin: admin / AsamAdmin2025! (no member associated)")
		fmt.Println("  User:  user / AsamUser2025! (member: A99001)")
		os.Exit(0)
	} else {
		fmt.Println("\n⚠️ Some users could not be created")
		fmt.Println("Please check the error messages above")
		os.Exit(1)
	}
}

// createOrUpdateAdminUser crea o actualiza el usuario administrador
func createOrUpdateAdminUser(db *gorm.DB) error {
	var user models.User

	// Check if admin user exists
	err := db.Where("username = ?", "admin").First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create new admin user (sin miembro asociado)
		user = models.User{
			Username:      "admin",
			Email:         "admin@asam.org",
			Role:          models.RoleAdmin,
			IsActive:      true,
			EmailVerified: true, // Admin pre-verificado
			MemberID:      nil,  // Admin no tiene miembro asociado
		}

		// Set password using the model method
		if err := user.SetPassword("AsamAdmin2025!"); err != nil {
			return fmt.Errorf("failed to set admin password: %w", err)
		}

		// Create user
		if err := db.Create(&user).Error; err != nil {
			// Check if it's a constraint violation
			if errors.Is(err, gorm.ErrCheckConstraintViolated) {
				return fmt.Errorf("constraint violation creating admin user: %w", err)
			}
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		fmt.Printf("✅ Created new admin user: %s (email: %s, ID: %d)\n",
			user.Username, user.Email, user.ID)

	case err == nil:
		// Update existing admin user
		originalID := user.ID

		if err := user.SetPassword("AsamAdmin2025!"); err != nil {
			return fmt.Errorf("failed to update admin password: %w", err)
		}

		// Ensure admin configuration
		user.IsActive = true
		user.Role = models.RoleAdmin
		user.EmailVerified = true
		user.MemberID = nil // Admin no debe tener miembro

		if user.Email != "admin@asam.org" {
			user.Email = "admin@asam.org"
		}

		// Save changes
		if err := db.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}

		fmt.Printf("✅ Updated existing admin user: %s (ID: %d)\n",
			user.Username, originalID)

	default:
		return fmt.Errorf("database error checking admin: %w", err)
	}

	return nil
}

// createOrUpdateUserWithMember crea o actualiza un usuario regular con su miembro asociado
func createOrUpdateUserWithMember(db *gorm.DB) error {
	// Helper para crear puntero a string
	stringPtr := func(s string) *string {
		return &s
	}

	// Usar números de membresía válidos según el formato requerido
	// Formato: [A|B] seguido de al menos 5 dígitos
	// Usamos A99xxx para miembros de prueba
	memberNumber := "A99001"

	// Primero, crear o buscar el miembro
	var member models.Member
	err := db.Where("membership_number = ?", memberNumber).First(&member).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Crear nuevo miembro para el usuario de prueba
		userEmail := "user@asam.org"
		member = models.Member{
			MembershipNumber: memberNumber, // Usar número válido
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Usuario",
			Surnames:         "Prueba García",
			Address:          "Calle Test 123",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			Email:            &userEmail, // Usar puntero a string
			State:            models.EstadoActivo,
			Nationality:      "Española",
			RegistrationDate: time.Now(),
		}

		if err := db.Create(&member).Error; err != nil {
			return fmt.Errorf("failed to create member: %w", err)
		}

		fmt.Printf("✅ Created new member: %s %s (ID: %d, Number: %s)\n",
			member.Name, member.Surnames, member.ID, member.MembershipNumber)

	case err == nil:
		fmt.Printf("ℹ️  Member already exists: %s %s (ID: %d)\n",
			member.Name, member.Surnames, member.ID)

		// Actualizar email del miembro si es necesario
		userEmail := "user@asam.org"
		if member.Email == nil || *member.Email != userEmail {
			member.Email = &userEmail
			if err := db.Save(&member).Error; err != nil {
				log.Printf("Warning: Could not update member email: %v", err)
			} else {
				fmt.Printf("   Updated member email to %s\n", userEmail)
			}
		}

	default:
		return fmt.Errorf("database error checking member: %w", err)
	}

	// Ahora crear o actualizar el usuario asociado al miembro
	var user models.User
	err = db.Where("username = ?", "user").First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Crear nuevo usuario con miembro asociado
		user = models.User{
			Username:      "user",
			Email:         "user@asam.org",
			Role:          models.RoleUser,
			MemberID:      &member.ID, // Asociar con el miembro
			IsActive:      true,
			EmailVerified: false, // Usuario regular debe verificar email
		}

		// Set password using the model method
		if err := user.SetPassword("AsamUser2025!"); err != nil {
			return fmt.Errorf("failed to set user password: %w", err)
		}

		// Create user
		if err := db.Create(&user).Error; err != nil {
			// Check if it's a constraint violation
			if errors.Is(err, gorm.ErrCheckConstraintViolated) {
				return fmt.Errorf("constraint violation: user role requires member but member_id=%v: %w",
					user.MemberID, err)
			}
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("✅ Created new user: %s (email: %s, MemberID: %d, ID: %d)\n",
			user.Username, user.Email, *user.MemberID, user.ID)

	case err == nil:
		// Update existing user
		originalID := user.ID

		if err := user.SetPassword("AsamUser2025!"); err != nil {
			return fmt.Errorf("failed to update user password: %w", err)
		}

		// Ensure user configuration
		user.IsActive = true
		user.Role = models.RoleUser
		user.MemberID = &member.ID // Asociar con el miembro

		if user.Email != "user@asam.org" {
			user.Email = "user@asam.org"
		}

		// Save changes
		if err := db.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		fmt.Printf("✅ Updated existing user: %s (MemberID: %d, ID: %d)\n",
			user.Username, *user.MemberID, originalID)

	default:
		return fmt.Errorf("database error checking user: %w", err)
	}

	// Crear algunos miembros adicionales sin usuarios para testing
	fmt.Println("\n📝 Creating additional test members...")
	testMembers := []struct {
		number   string
		name     string
		surnames string
		email    string
	}{
		{"A99002", "María", "González López", "maria.gonzalez@example.com"},
		{"A99003", "Carlos", "Rodríguez Martín", "carlos.rodriguez@example.com"},
		{"B99001", "Ana", "Martínez Sánchez", "ana.martinez@example.com"},
		{"B99002", "Pedro", "López Fernández", "pedro.lopez@example.com"},
	}

	for _, tm := range testMembers {
		var existingMember models.Member
		if err := db.Where("membership_number = ?", tm.number).First(&existingMember).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newMember := models.Member{
					MembershipNumber: tm.number,
					MembershipType:   models.TipoMembresiaPIndividual,
					Name:             tm.name,
					Surnames:         tm.surnames,
					Email:            stringPtr(tm.email), // Usar helper para crear puntero
					Address:          "Calle Ejemplo 100",
					Postcode:         "08001",
					City:             "Barcelona",
					Province:         "Barcelona",
					Country:          "España",
					State:            models.EstadoActivo,
					Nationality:      "Española",
					RegistrationDate: time.Now(),
				}

				if err := db.Create(&newMember).Error; err != nil {
					log.Printf("⚠️  Could not create additional member %s: %v", tm.number, err)
				} else {
					fmt.Printf("   ✓ Created member: %s %s (Number: %s) - Available for association\n",
						tm.name, tm.surnames, tm.number)
				}
			} else {
				fmt.Printf("   ℹ️  Member %s already exists\n", tm.number)
			}
		}
	}

	return nil
}
