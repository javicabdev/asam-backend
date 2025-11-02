// Package main provides automated test user creation functionality
package main

import (
	"context"
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
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/logger"
)

// This script automatically creates test users without user interaction
// It's used by start-docker.ps1 to initialize the system

// Constants for test users
const (
	adminEmail    = "javierfernandezc@gmail.com"
	userEmail     = "javi_nov20@hotmail.com"
	adminUsername = "admin"
	userUsername  = "user"
	adminPassword = "AsamAdmin2025!"
	userPassword  = "AsamUser2025!"
)

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
	if err := createOrUpdateUserWithMember(database, cfg); err != nil {
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
		fmt.Println("  User:  user / AsamUser2025! (member: B99001)")
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
	err := db.Where("username = ?", adminUsername).First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Create new admin user (sin miembro asociado)
		user = models.User{
			Username:      adminUsername,
			Email:         adminEmail,
			Role:          models.RoleAdmin,
			IsActive:      true,
			EmailVerified: false,
			MemberID:      nil, // Admin no tiene miembro asociado
		}

		// Set password using the model method
		if err := user.SetPassword(adminPassword); err != nil {
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

		if err := user.SetPassword(adminPassword); err != nil {
			return fmt.Errorf("failed to update admin password: %w", err)
		}

		// Ensure admin configuration
		user.IsActive = true
		user.Role = models.RoleAdmin
		user.EmailVerified = true
		user.MemberID = nil // Admin no debe tener miembro

		if user.Email != adminEmail {
			user.Email = adminEmail
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
func createOrUpdateUserWithMember(db *gorm.DB, cfg *config.Config) error {
	// Primero crear o buscar el miembro
	member, err := getOrCreateTestMember(db)
	if err != nil {
		return err
	}

	// Luego crear o actualizar el usuario
	if err := createOrUpdateTestUser(db, member); err != nil {
		return err
	}

	// Crear pago pendiente para el primer usuario (B99001)
	currentYear := time.Now().Year()
	var membershipFee models.MembershipFee
	if err := db.Where("year = ?", currentYear).First(&membershipFee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Crear cuota anual si no existe
			membershipFee = models.MembershipFee{
				Year:           currentYear,
				BaseFeeAmount:  30.0, // Cuota base
				FamilyFeeExtra: 10.0, // Extra para familias
				DueDate:        time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC),
			}
			if err := db.Create(&membershipFee).Error; err != nil {
				log.Printf("⚠️  Could not create membership fee for first user: %v", err)
			} else {
				fmt.Printf("   ✓ Created membership fee for year %d\n", currentYear)
			}
		}
	}

	// Verificar si el miembro B99001 ya tiene un pago pendiente
	var existingPayment models.Payment
	if err := db.Where("member_id = ? AND membership_fee_id = ?", member.ID, membershipFee.ID).First(&existingPayment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Crear pago pendiente para B99001
			payment := models.Payment{
				MemberID:        &member.ID,
				Amount:          membershipFee.BaseFeeAmount,
				Status:          models.PaymentStatusPending,
				PaymentDate:     nil,
				MembershipFeeID: &membershipFee.ID,
				Notes:           "Pago de prueba generado automáticamente para usuario test",
			}
			if err := db.Create(&payment).Error; err != nil {
				log.Printf("⚠️  Could not create pending payment for first user: %v", err)
			} else {
				fmt.Printf("   ✓ Created pending payment (%.2f€) for member %s\n", payment.Amount, member.MembershipNumber)
			}
		}
	}

	// Finalmente crear miembros adicionales para testing
	createAdditionalTestMembers(db, cfg)

	return nil
}

// getOrCreateTestMember obtiene o crea el miembro de prueba
func getOrCreateTestMember(db *gorm.DB) (*models.Member, error) {
	// Usar números de membresía válidos según el formato requerido
	// Formato: [A|B] seguido de al menos 5 dígitos
	// Convención ASAM:
	// - A: Miembros FAMILIARES (asociados a una familia)
	// - B: Miembros INDIVIDUALES
	// Usamos B99xxx para miembros individuales de prueba
	memberNumber := "B99001"

	var member models.Member
	err := db.Where("membership_number = ?", memberNumber).First(&member).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Crear nuevo miembro para el usuario de prueba
		member = models.Member{
			MembershipNumber: memberNumber,
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Usuario",
			Surnames:         "Prueba García",
			Address:          "Calle Test 123",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			Email:            stringPtr(userEmail),
			State:            models.EstadoActivo,
			Nationality:      "Española",
			RegistrationDate: time.Now(),
		}

		if err := db.Create(&member).Error; err != nil {
			return nil, fmt.Errorf("failed to create member: %w", err)
		}

		fmt.Printf("✅ Created new member: %s %s (ID: %d, Number: %s)\n",
			member.Name, member.Surnames, member.ID, member.MembershipNumber)

	case err == nil:
		fmt.Printf("ℹ️  Member already exists: %s %s (ID: %d)\n",
			member.Name, member.Surnames, member.ID)

		// Actualizar email del miembro si es necesario
		if member.Email == nil || *member.Email != userEmail {
			member.Email = stringPtr(userEmail)
			if err := db.Save(&member).Error; err != nil {
				log.Printf("Warning: Could not update member email: %v", err)
			} else {
				fmt.Printf("   Updated member email to %s\n", userEmail)
			}
		}

	default:
		return nil, fmt.Errorf("database error checking member: %w", err)
	}

	return &member, nil
}

// createOrUpdateTestUser crea o actualiza el usuario de prueba
func createOrUpdateTestUser(db *gorm.DB, member *models.Member) error {
	var user models.User
	err := db.Where("username = ?", userUsername).First(&user).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Crear nuevo usuario con miembro asociado
		user = models.User{
			Username:      userUsername,
			Email:         userEmail,
			Role:          models.RoleUser,
			MemberID:      &member.ID,
			IsActive:      true,
			EmailVerified: false, // Usuario regular debe verificar email
		}

		// Set password using the model method
		if err := user.SetPassword(userPassword); err != nil {
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

		if err := user.SetPassword(userPassword); err != nil {
			return fmt.Errorf("failed to update user password: %w", err)
		}

		// Ensure user configuration
		user.IsActive = true
		user.Role = models.RoleUser
		user.MemberID = &member.ID

		if user.Email != userEmail {
			user.Email = userEmail
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

	return nil
}

// createAdditionalTestMembers crea miembros adicionales sin usuarios para testing
// Incluye miembros individuales y familiares, todos con pagos pendientes
func createAdditionalTestMembers(database *gorm.DB, _ *config.Config) {
	fmt.Println("\n📝 Creating additional test members with pending payments...")

	ctx := context.Background()
	currentYear := time.Now().Year()

	// Inicializar logger para el servicio
	_, err := logger.InitLogger(logger.DefaultConfig())
	if err != nil {
		log.Printf("⚠️  Could not initialize logger: %v", err)
		return
	}

	// Inicializar repositorios
	memberRepo := db.NewMemberRepository(database)
	familyRepo := db.NewFamilyRepository(database)
	paymentRepo := db.NewPaymentRepository(database)
	membershipFeeRepo := db.NewMembershipFeeRepository(database)

	// Inicializar servicio de familias
	familyService := services.NewFamilyService(familyRepo, memberRepo, paymentRepo, membershipFeeRepo)

	// Crear o obtener cuota anual del año actual
	var membershipFee models.MembershipFee
	if err := database.Where("year = ?", currentYear).First(&membershipFee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Crear cuota anual si no existe
			membershipFee = models.MembershipFee{
				Year:           currentYear,
				BaseFeeAmount:  30.0, // Cuota base
				FamilyFeeExtra: 10.0, // Extra para familias
				DueDate:        time.Date(currentYear, 12, 31, 23, 59, 59, 0, time.UTC),
			}
			if err := database.Create(&membershipFee).Error; err != nil {
				log.Printf("⚠️  Could not create membership fee: %v", err)
				return
			}
			fmt.Printf("   ✓ Created membership fee for year %d\n", currentYear)
		} else {
			log.Printf("⚠️  Error checking membership fee: %v", err)
			return
		}
	}

	// MIEMBROS INDIVIDUALES (prefijo B)
	individualMembers := []struct {
		number   string
		name     string
		surnames string
		email    string
	}{
		{"B99002", "María", "González López", "maria.gonzalez@example.com"},
		{"B99003", "Carlos", "Rodríguez Martín", "carlos.rodriguez@example.com"},
		{"B99004", "Ana", "Martínez Sánchez", "ana.martinez@example.com"},
	}

	fmt.Println("\n   Creating individual members:")
	for _, tm := range individualMembers {
		member := createOrGetMember(database, tm.number, tm.name, tm.surnames, tm.email, models.TipoMembresiaPIndividual)
		if member != nil {
			// Crear pago pendiente para el miembro
			createPendingPayment(database, &member.ID, nil, membershipFee.BaseFeeAmount, membershipFee.ID)
		}
	}

	// FAMILIAS (prefijo A) - Usar CreateFamilyAtomic
	// Usando DNIs válidos españoles (con letra de control correcta)
	familyData := []struct {
		number          string
		esposoNombre    string
		esposoApellidos string
		esposoEmail     string
		esposoDNI       string
		esposaNombre    string
		esposaApellidos string
		esposaEmail     string
	}{
		{"A99001", "Pedro", "López Fernández", "pedro.lopez@example.com", "12345678Z", "Laura", "García Ruiz", "laura.garcia@example.com"},
		{"A99002", "Miguel", "Sánchez Torres", "miguel.sanchez@example.com", "87654321X", "Carmen", "Díaz Moreno", "carmen.diaz@example.com"},
	}

	fmt.Println("\n   Creating family members:")
	for _, fm := range familyData {
		// Verificar si ya existe la familia
		existingFamily, _ := familyService.GetByNumeroSocio(ctx, fm.number)
		if existingFamily != nil {
			fmt.Printf("      ℹ️  Family %s already exists\n", fm.number)
			// Verificar si tiene miembro origen
			if existingFamily.MiembroOrigenID != nil {
				// Verificar si ya tiene pago pendiente
				var existingPayment models.Payment
				if err := database.Where("member_id = ? AND membership_fee_id = ?",
					*existingFamily.MiembroOrigenID, membershipFee.ID).First(&existingPayment).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						// Crear pago pendiente para la familia existente
						familyAmount := membershipFee.BaseFeeAmount + membershipFee.FamilyFeeExtra
						createPendingPayment(database, existingFamily.MiembroOrigenID, nil, familyAmount, membershipFee.ID)
					}
				}
			}
			continue
		}

		// Crear familia con miembro origen usando CreateFamilyAtomic
		req := &input.CreateFamilyAtomicRequest{
			Family: &models.Family{
				NumeroSocio:              fm.number,
				EsposoNombre:             fm.esposoNombre,
				EsposoApellidos:          fm.esposoApellidos,
				EsposoCorreoElectronico:  fm.esposoEmail,
				EsposoDocumentoIdentidad: fm.esposoDNI, // DNI obligatorio cuando CreateMemberIfNotExists=true
				EsposaNombre:             fm.esposaNombre,
				EsposaApellidos:          fm.esposaApellidos,
				EsposaCorreoElectronico:  fm.esposaEmail,
			},
			CreateMemberIfNotExists: true,
			MemberData: &input.CreateMemberData{
				Address:  "Calle Ejemplo 100",
				Postcode: "08001",
				City:     "Barcelona",
				Province: "Barcelona",
				Country:  "España",
			},
		}

		family, err := familyService.CreateFamilyAtomic(ctx, req)
		if err != nil {
			log.Printf("      ⚠️  Could not create family %s: %v", fm.number, err)
			continue
		}

		fmt.Printf("      ✓ Created family: %s %s & %s %s (Number: %s)\n",
			fm.esposoNombre, fm.esposoApellidos, fm.esposaNombre, fm.esposaApellidos, fm.number)

		// Obtener el miembro origen creado
		if family.MiembroOrigenID != nil {
			// Crear pago pendiente para el miembro familiar (incluye extra familiar)
			familyAmount := membershipFee.BaseFeeAmount + membershipFee.FamilyFeeExtra
			createPendingPayment(database, family.MiembroOrigenID, nil, familyAmount, membershipFee.ID)
		}
	}

	fmt.Println("\n✅ Test members with pending payments created successfully!")
}

// createOrGetMember crea o obtiene un miembro individual
func createOrGetMember(db *gorm.DB, number, name, surnames, email string, memberType string) *models.Member {
	var member models.Member
	if err := db.Where("membership_number = ?", number).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			member = models.Member{
				MembershipNumber: number,
				MembershipType:   memberType,
				Name:             name,
				Surnames:         surnames,
				Email:            stringPtr(email),
				Address:          "Calle Ejemplo 100",
				Postcode:         "08001",
				City:             "Barcelona",
				Province:         "Barcelona",
				Country:          "España",
				State:            models.EstadoActivo,
				Nationality:      "Española",
				RegistrationDate: time.Now(),
			}

			if err := db.Create(&member).Error; err != nil {
				log.Printf("      ⚠️  Could not create member %s: %v", number, err)
				return nil
			}
			fmt.Printf("      ✓ Created individual member: %s %s (Number: %s)\n",
				name, surnames, number)
			return &member
		}
		log.Printf("      ⚠️  Error checking member %s: %v", number, err)
		return nil
	}
	fmt.Printf("      ℹ️  Individual member %s already exists\n", number)
	return &member
}

// createPendingPayment crea un pago pendiente para un miembro o familia
func createPendingPayment(db *gorm.DB, memberID, familyID *uint, amount float64, feeID uint) {
	// Verificar si ya existe un pago para esta entidad y cuota
	var existingPayment models.Payment
	query := db.Where("membership_fee_id = ?", feeID)
	if memberID != nil {
		query = query.Where("member_id = ?", *memberID)
	} else if familyID != nil {
		query = query.Where("family_id = ?", *familyID)
	}

	if err := query.First(&existingPayment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Crear nuevo pago pendiente
			payment := models.Payment{
				MemberID:        memberID,
				FamilyID:        familyID,
				Amount:          amount,
				Status:          models.PaymentStatusPending,
				PaymentDate:     nil, // Sin fecha de pago porque está pendiente
				MembershipFeeID: &feeID,
				Notes:           "Pago de prueba generado automáticamente",
			}

			if err := db.Create(&payment).Error; err != nil {
				entityType := "member"
				entityID := uint(0)
				if memberID != nil {
					entityID = *memberID
				} else if familyID != nil {
					entityType = "family"
					entityID = *familyID
				}
				log.Printf("         ⚠️  Could not create pending payment for %s %d: %v", entityType, entityID, err)
			} else {
				entityInfo := ""
				if memberID != nil {
					entityInfo = fmt.Sprintf("member ID %d", *memberID)
				} else if familyID != nil {
					entityInfo = fmt.Sprintf("family ID %d", *familyID)
				}
				fmt.Printf("         ✓ Created pending payment (%.2f€) for %s\n", amount, entityInfo)
			}
		}
	}
}

// stringPtr helper para crear puntero a string
func stringPtr(s string) *string {
	return &s
}
