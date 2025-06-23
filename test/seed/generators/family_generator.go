package generators

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

// FamilyGenerator generates test families
type FamilyGenerator struct {
	db        *sqlx.DB
	rand      *rand.Rand
	lastCount int
}

// Family represents a family record for generation
type Family struct {
	ID                       int        `db:"id"`
	NumeroSocio              string     `db:"numero_socio"`
	MiembroOrigenID          *int       `db:"miembro_origen_id"`
	EsposoNombre             string     `db:"esposo_nombre"`
	EsposoApellidos          string     `db:"esposo_apellidos"`
	EsposaNombre             string     `db:"esposa_nombre"`
	EsposaApellidos          string     `db:"esposa_apellidos"`
	EsposoFechaNacimiento    *time.Time `db:"esposo_fecha_nacimiento"`
	EsposoDocumentoIdentidad string     `db:"esposo_documento_identidad"`
	EsposoCorreoElectronico  string     `db:"esposo_correo_electronico"`
	EsposaFechaNacimiento    *time.Time `db:"esposa_fecha_nacimiento"`
	EsposaDocumentoIdentidad string     `db:"esposa_documento_identidad"`
	EsposaCorreoElectronico  string     `db:"esposa_correo_electronico"`
}

// NewFamilyGenerator creates a new family generator
func NewFamilyGenerator(db *sqlx.DB, seed int64) *FamilyGenerator {
	return &FamilyGenerator{
		db:   db,
		rand: rand.New(rand.NewSource(seed)),
	}
}

// Generate creates n random families
func (g *FamilyGenerator) Generate(ctx context.Context, n int) error {
	// Get current count to start sequence
	var count int
	err := g.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM families")
	if err != nil {
		return fmt.Errorf("failed to get family count: %w", err)
	}

	g.lastCount = count + 1

	// Get potential origin members (with tipo_membresia = 'familiar')
	var originMembers []struct {
		MiembroID int `db:"miembro_id"`
	}

	err = g.db.SelectContext(ctx, &originMembers,
		"SELECT id as miembro_id FROM members WHERE membership_type = 'familiar' AND state = 'active'")
	if err != nil {
		return fmt.Errorf("failed to get potential origin members: %w", err)
	}

	if len(originMembers) == 0 {
		return fmt.Errorf("no family-type members found to associate with families")
	}

	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert families in batches
	batchSize := 10
	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}

		if err := g.generateBatch(ctx, tx, originMembers, i, end); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but continue with original error
				fmt.Printf("rollback error: %v\n", rollbackErr)
			}
			return err
		}
	}

	return tx.Commit()
}

// generateBatch generates a batch of families
func (g *FamilyGenerator) generateBatch(
	ctx context.Context,
	tx *sqlx.Tx,
	originMembers []struct {
		MiembroID int `db:"miembro_id"`
	},
	start,
	end int,
) error {
	// Prepare query
	query := `
		INSERT INTO families (
			numero_socio, miembro_origen_id,
			esposo_nombre, esposo_apellidos, esposa_nombre, esposa_apellidos,
			esposo_fecha_nacimiento, esposo_documento_identidad, esposo_correo_electronico,
			esposa_fecha_nacimiento, esposa_documento_identidad, esposa_correo_electronico,
			created_at, updated_at
		) VALUES (
			:numero_socio, :miembro_origen_id,
			:esposo_nombre, :esposo_apellidos, :esposa_nombre, :esposa_apellidos,
			:esposo_fecha_nacimiento, :esposo_documento_identidad, :esposo_correo_electronico,
			:esposa_fecha_nacimiento, :esposa_documento_identidad, :esposa_correo_electronico,
			NOW(), NOW()
		)
	`

	// Generate families
	families := make([]Family, 0, end-start)
	for i := start; i < end; i++ {
		// Select a random origin member
		originMemberID := originMembers[g.rand.Intn(len(originMembers))].MiembroID

		family := g.generateFamily(originMemberID)
		families = append(families, family)
	}

	// Insert batch
	_, err := tx.NamedExecContext(ctx, query, families)
	if err != nil {
		return fmt.Errorf("failed to insert families: %w", err)
	}

	return nil
}

// generateFamily creates a single random family
func (g *FamilyGenerator) generateFamily(originMemberID int) Family {
	// Generate husband data
	husbandFirstName, husbandLastName := GenerateRandomName(g.rand, "male")

	// Generate wife data
	wifeFirstName, wifeLastName := GenerateRandomName(g.rand, "female")

	// Generate birth dates - between 25 and 75 years ago
	now := time.Now()

	husbandBirthYear := now.Year() - g.rand.Intn(50) - 25
	husbandBirthMonth := time.Month(g.rand.Intn(12) + 1)
	husbandBirthDay := g.rand.Intn(28) + 1
	husbandBirthDate := time.Date(husbandBirthYear, husbandBirthMonth, husbandBirthDay, 0, 0, 0, 0, time.UTC)

	wifeBirthYear := now.Year() - g.rand.Intn(50) - 25
	wifeBirthMonth := time.Month(g.rand.Intn(12) + 1)
	wifeBirthDay := g.rand.Intn(28) + 1
	wifeBirthDate := time.Date(wifeBirthYear, wifeBirthMonth, wifeBirthDay, 0, 0, 0, 0, time.UTC)

	// Generate IDs - mix of DNI and NIE
	var husbandID, wifeID string

	if g.rand.Float64() < 0.6 { // 60% Spanish
		husbandID = GenerateRandomDNI(g.rand)
	} else {
		husbandID = GenerateRandomNIE(g.rand)
	}

	if g.rand.Float64() < 0.6 { // 60% Spanish
		wifeID = GenerateRandomDNI(g.rand)
	} else {
		wifeID = GenerateRandomNIE(g.rand)
	}

	// Generate emails
	husbandEmail := GenerateRandomEmail(g.rand, husbandFirstName, husbandLastName)
	wifeEmail := GenerateRandomEmail(g.rand, wifeFirstName, wifeLastName)

	// Generate membership number (format: FAM-XXXX)
	memberNumber := GenerateRandomMembershipNumber(g.rand, "FAM", g.lastCount)
	g.lastCount++

	return Family{
		NumeroSocio:              memberNumber,
		MiembroOrigenID:          &originMemberID,
		EsposoNombre:             husbandFirstName,
		EsposoApellidos:          husbandLastName,
		EsposaNombre:             wifeFirstName,
		EsposaApellidos:          wifeLastName,
		EsposoFechaNacimiento:    &husbandBirthDate,
		EsposoDocumentoIdentidad: husbandID,
		EsposoCorreoElectronico:  husbandEmail,
		EsposaFechaNacimiento:    &wifeBirthDate,
		EsposaDocumentoIdentidad: wifeID,
		EsposaCorreoElectronico:  wifeEmail,
	}
}

// GetAllFamilies retrieves all families from the database
func (g *FamilyGenerator) GetAllFamilies(ctx context.Context) ([]Family, error) {
	var families []Family

	query := `
		SELECT 
			id, numero_socio, miembro_origen_id,
			esposo_nombre, esposo_apellidos, esposa_nombre, esposa_apellidos,
			esposo_fecha_nacimiento, esposo_documento_identidad, esposo_correo_electronico,
			esposa_fecha_nacimiento, esposa_documento_identidad, esposa_correo_electronico
		FROM families
	`

	err := g.db.SelectContext(ctx, &families, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get families: %w", err)
	}

	return families, nil
}

// GetLastInsertedFamilies retrieves the last n families inserted into the database
func (g *FamilyGenerator) GetLastInsertedFamilies(ctx context.Context, n int) ([]Family, error) {
	var families []Family

	query := `
		SELECT 
			id, numero_socio, miembro_origen_id,
			esposo_nombre, esposo_apellidos, esposa_nombre, esposa_apellidos,
			esposo_fecha_nacimiento, esposo_documento_identidad, esposo_correo_electronico,
			esposa_fecha_nacimiento, esposa_documento_identidad, esposa_correo_electronico
		FROM families
		ORDER BY id DESC
		LIMIT $1
	`

	err := g.db.SelectContext(ctx, &families, query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get last inserted families: %w", err)
	}

	return families, nil
}
