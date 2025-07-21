package seeds

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// SeedUserWithMember crea un usuario de ejemplo con socio asociado
func SeedUserWithMember(db *gorm.DB) error {
	// 1. Verificar si ya existe el socio
	var existingMember models.Member
	if err := db.Where("membership_number = ?", "M001").First(&existingMember).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("error checking existing member: %w", err)
		}

		// 2. Crear el socio si no existe
		member := models.Member{
			MembershipNumber: "M001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "Pérez García",
			Address:          "Calle Mayor 123",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			Nationality:      "Española",
			RegistrationDate: time.Now(),
		}

		if err := db.Create(&member).Error; err != nil {
			return fmt.Errorf("error creating member: %w", err)
		}
		existingMember = member
		log.Printf("✓ Created member: %s %s (ID: %d)", member.Name, member.Surnames, member.ID)
	}

	// 3. Verificar si ya existe el usuario
	var existingUser models.User
	if err := db.Where("username = ?", "juan.perez").First(&existingUser).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("error checking existing user: %w", err)
		}

		// 4. Crear el usuario asociado al socio
		user := models.User{
			Username:      "juan.perez",
			Email:         "juan.perez@example.com",
			Role:          models.RoleUser,
			MemberID:      &existingMember.ID,
			IsActive:      true,
			EmailVerified: true,
		}

		// Establecer contraseña
		if err := user.SetPassword("password123"); err != nil {
			return fmt.Errorf("error setting password: %w", err)
		}

		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("error creating user: %w", err)
		}

		log.Printf("✓ Created user: %s (Role: %s, MemberID: %d)", user.Username, user.Role, *user.MemberID)
	} else if existingUser.MemberID == nil {
		// Actualizar el usuario existente para asociarlo al socio
		existingUser.MemberID = &existingMember.ID
		if err := db.Save(&existingUser).Error; err != nil {
			return fmt.Errorf("error updating user: %w", err)
		}
		log.Printf("✓ Updated user %s with MemberID: %d", existingUser.Username, *existingUser.MemberID)
	}

	// 5. Crear un segundo ejemplo: socio sin usuario (para futura asociación)
	var member2 models.Member
	if err := db.Where("membership_number = ?", "M002").First(&member2).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			member2 = models.Member{
				MembershipNumber: "M002",
				MembershipType:   models.TipoMembresiaPIndividual,
				Name:             "María",
				Surnames:         "González López",
				Address:          "Calle Nueva 456",
				Postcode:         "08002",
				City:             "Barcelona",
				Province:         "Barcelona",
				Country:          "España",
				State:            models.EstadoActivo,
				Nationality:      "Española",
				RegistrationDate: time.Now(),
			}

			if err := db.Create(&member2).Error; err != nil {
				return fmt.Errorf("error creating member2: %w", err)
			}
			log.Printf("✓ Created member without user: %s %s (ID: %d) - Available for association",
				member2.Name, member2.Surnames, member2.ID)
		}
	}

	log.Println("✓ User-Member seed completed successfully")
	log.Println("  Test credentials:")
	log.Println("  - Admin: admin / admin123")
	log.Println("  - User with member: juan.perez / password123")

	return nil
}

// SeedUserMemberAssociations actualiza usuarios existentes con socios
func SeedUserMemberAssociations(db *gorm.DB) error {
	// Este seed es útil para migrar usuarios existentes
	// y asociarlos con socios basándose en algún criterio

	// Ejemplo: asociar usuarios por email coincidente
	var users []models.User
	if err := db.Where("role = ? AND member_id IS NULL", models.RoleUser).Find(&users).Error; err != nil {
		return fmt.Errorf("error finding users without members: %w", err)
	}

	for _, user := range users {
		// Buscar socio por email
		var member models.Member
		if err := db.Where("email = ?", user.Email).First(&member).Error; err == nil {
			// Asociar usuario con socio
			user.MemberID = &member.ID
			if err := db.Save(&user).Error; err != nil {
				log.Printf("✗ Error associating user %s with member %d: %v",
					user.Username, member.ID, err)
			} else {
				log.Printf("✓ Associated user %s with member %s %s (ID: %d)",
					user.Username, member.Name, member.Surnames, member.ID)
			}
		}
	}

	return nil
}
