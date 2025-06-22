package generators

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

// FamilyMemberGenerator generates test family members (familiares)
type FamilyMemberGenerator struct {
	db   *sqlx.DB
	rand *rand.Rand
}

// FamiliarMember represents a familiar record for generation
type FamiliarMember struct {
	ID                uint       `db:"id"`
	FamiliaID         uint       `db:"familia_id"`
	Nombre            string     `db:"nombre"`
	Apellidos         string     `db:"apellidos"`
	DniNie            string     `db:"dni_nie"`
	FechaNacimiento   *time.Time `db:"fecha_nacimiento"`
	CorreoElectronico string     `db:"correo_electronico"`
	Parentesco        string     `db:"parentesco"`
}

// NewFamilyMemberGenerator creates a new family member generator
func NewFamilyMemberGenerator(db *sqlx.DB, seed int64) *FamilyMemberGenerator {
	return &FamilyMemberGenerator{
		db:   db,
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Generate creates n random family members
func (g *FamilyMemberGenerator) Generate(ctx context.Context, n int) error {
	// Get families to associate with
	var families []struct {
		FamiliaID   int    `db:"familia_id"`
		NumeroSocio string `db:"numero_socio"`
	}

	err := g.db.SelectContext(ctx, &families, "SELECT id as familia_id, numero_socio FROM families")
	if err != nil {
		return fmt.Errorf("failed to get families: %w", err)
	}

	if len(families) == 0 {
		return fmt.Errorf("no families found to associate with family members")
	}

	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert family members in batches
	batchSize := 10
	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}

		if err := g.generateBatch(ctx, tx, families, i, end); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but continue with original error
				fmt.Printf("rollback error: %v\n", rollbackErr)
			}
			return err
		}
	}

	return tx.Commit()
}

// generateBatch generates a batch of family members
func (g *FamilyMemberGenerator) generateBatch(
	ctx context.Context,
	tx *sqlx.Tx,
	families []struct {
		FamiliaID   int    `db:"familia_id"`
		NumeroSocio string `db:"numero_socio"`
	},
	start,
	end int,
) error {
	// Prepare query
	query := `
		INSERT INTO familiars (
			familia_id, nombre, apellidos, dni_nie, fecha_nacimiento, correo_electronico, parentesco,
			created_at, updated_at
		) VALUES (
			:familia_id, :nombre, :apellidos, :dni_nie, :fecha_nacimiento, :correo_electronico, :parentesco,
			NOW(), NOW()
		)
	`

	// Generate family members
	familyMembers := make([]FamiliarMember, 0, end-start)
	for i := start; i < end; i++ {
		// Select a random family
		family := families[g.rand.Intn(len(families))]

		// Generate 1-3 family members per family
		numMembers := g.rand.Intn(3) + 1
		for j := 0; j < numMembers; j++ {
			if len(familyMembers) >= end-start {
				break // Ensure we don't exceed requested count
			}

			familyMember := g.generateFamilyMember(family.FamiliaID)
			familyMembers = append(familyMembers, familyMember)
		}

		if len(familyMembers) >= end-start {
			break
		}
	}

	// Insert batch
	_, err := tx.NamedExecContext(ctx, query, familyMembers)
	if err != nil {
		return fmt.Errorf("failed to insert family members: %w", err)
	}

	return nil
}

// generateFamilyMember creates a single random family member
func (g *FamilyMemberGenerator) generateFamilyMember(familiaID int) FamiliarMember {
	// 50% male, 50% female
	gender := []string{"male", "female"}[g.rand.Intn(2)]
	firstName, lastName := GenerateRandomName(g.rand, gender)

	// Generate random ID (75% chance for DNI if child is older, 75% chance for NIE if child is younger)
	var documentID string

	// Generate birthday - typically children are between 0-25 years old
	now := time.Now()
	childAge := g.rand.Intn(26) // 0-25 years

	birthYear := now.Year() - childAge
	birthMonth := time.Month(g.rand.Intn(12) + 1)
	birthDay := g.rand.Intn(28) + 1
	birthDate := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)

	// Children under 14 might not have ID
	var hasID bool
	if childAge < 14 {
		hasID = g.rand.Float64() < 0.3 // 30% chance younger kids have ID
	} else {
		hasID = g.rand.Float64() < 0.9 // 90% chance older kids have ID
	}

	if hasID {
		if g.rand.Float64() < 0.7 { // 70% Spanish
			documentID = GenerateRandomDNI(g.rand)
		} else {
			documentID = GenerateRandomNIE(g.rand)
		}
	}

	// Children under 14 typically don't have email
	var email string
	if childAge >= 14 && g.rand.Float64() < 0.8 { // 80% chance for email if 14+
		email = GenerateRandomEmail(g.rand, firstName, "familiar")
	}

	// Determine parentesco (relationship)
	var parentesco string
	if gender == "male" {
		parentesco = "Hijo"
	} else {
		parentesco = "Hija"
	}
	// 10% chance of being "Otro" (other relationship)
	if g.rand.Float64() < 0.1 {
		parentesco = "Otro"
	}

	return FamiliarMember{
		FamiliaID:         uint(familiaID),
		Nombre:            firstName,
		Apellidos:         lastName,
		DniNie:            documentID,
		FechaNacimiento:   &birthDate,
		CorreoElectronico: email,
		Parentesco:        parentesco,
	}
}

// GetFamilyMembersByFamily retrieves all family members for a given family
func (g *FamilyMemberGenerator) GetFamilyMembersByFamily(ctx context.Context, familiaID int) ([]FamiliarMember, error) {
	var familyMembers []FamiliarMember

	query := `
		SELECT id, familia_id, nombre, apellidos, dni_nie, fecha_nacimiento, correo_electronico, parentesco
		FROM familiars
		WHERE familia_id = $1
	`

	err := g.db.SelectContext(ctx, &familyMembers, query, familiaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family members: %w", err)
	}

	return familyMembers, nil
}
