package generators

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
)

// MemberGenerator generates test members
type MemberGenerator struct {
	db        *sqlx.DB
	rand      *rand.Rand
	lastCount int
}

// Member represents a member record for generation
type Member struct {
	ID               uint       `db:"id"`
	MembershipNumber string     `db:"membership_number"`
	MembershipType   string     `db:"membership_type"`
	Name             string     `db:"name"`
	Surnames         string     `db:"surnames"`
	Address          string     `db:"address"`
	Postcode         string     `db:"postcode"`
	City             string     `db:"city"`
	Province         string     `db:"province"`
	Country          string     `db:"country"`
	State            string     `db:"state"`
	RegistrationDate time.Time  `db:"registration_date"`
	LeavingDate      *time.Time `db:"leaving_date"`
	BirthDate        *time.Time `db:"birth_date"`
	IdentityCard     string     `db:"identity_card"`
	Email            string     `db:"email"`
	Profession       string     `db:"profession"`
	Nationality      string     `db:"nationality"`
	Remarks          string     `db:"remarks"`
}

// NewMemberGenerator creates a new member generator
func NewMemberGenerator(db *sqlx.DB, seed int64) *MemberGenerator {
	return &MemberGenerator{
		db:   db,
		rand: rand.New(rand.NewSource(seed)), //nolint:gosec // G404: math/rand is acceptable for test data
	}
}

// Generate creates n random members
func (g *MemberGenerator) Generate(ctx context.Context, n int) error {
	// Get current count to start sequence
	var count int
	err := g.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM members")
	if err != nil {
		return fmt.Errorf("failed to get member count: %w", err)
	}

	g.lastCount = count + 1

	tx, err := g.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Insert members in batches
	batchSize := 10
	for i := 0; i < n; i += batchSize {
		end := i + batchSize
		if end > n {
			end = n
		}

		if err := g.generateBatch(ctx, tx, i, end); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				// Log rollback error but continue with original error
				fmt.Printf("rollback error: %v\n", rollbackErr)
			}
			return err
		}
	}

	return tx.Commit()
}

// generateBatch generates a batch of members
func (g *MemberGenerator) generateBatch(ctx context.Context, tx *sqlx.Tx, start, end int) error {
	// Prepare query
	query := `
		INSERT INTO members (
			membership_number, membership_type, name, surnames, 
			address, postcode, city, province, country,
			state, registration_date, leaving_date, birth_date, 
			identity_card, email, profession, nationality, remarks,
			created_at, updated_at
		) VALUES (
			:membership_number, :membership_type, :name, :surnames, 
			:address, :postcode, :city, :province, :country,
			:state, :registration_date, :leaving_date, :birth_date, 
			:identity_card, :email, :profession, :nationality, :remarks,
			NOW(), NOW()
		)
	`

	// Generate members
	members := make([]Member, 0, end-start)
	for i := start; i < end; i++ {
		member := g.generateMember()
		members = append(members, member)
	}

	// Insert batch
	_, err := tx.NamedExecContext(ctx, query, members)
	if err != nil {
		return fmt.Errorf("failed to insert members: %w", err)
	}

	return nil
}

// generateMember creates a single random member
func (g *MemberGenerator) generateMember() Member {
	// Generate random data
	gender := []string{"male", "female"}[g.rand.Intn(2)]
	firstName, lastName := GenerateRandomName(g.rand, gender)

	address, postalCode, city, province := GenerateRandomAddress(g.rand)

	// Generate ID based on nationality
	var documentID string
	var nationality string

	// 70% Spanish, 30% Senegalese
	if g.rand.Float64() < 0.7 {
		documentID = GenerateRandomDNI(g.rand)
		nationality = "España"
	} else {
		documentID = GenerateRandomNIE(g.rand)
		nationality = "Senegal"
	}

	// Determine membership type (70% individual, 30% family)
	membershipType := "individual"
	if g.rand.Float64() < 0.3 {
		membershipType = "familiar"
	}

	// Generate member number (format: B-XXXX for individual, A-XXXX for family)
	// According to new requirements: Individual members start with 'B', Family members start with 'A'
	prefix := "B"
	if membershipType == "familiar" {
		prefix = "A"
	}
	memberNumber := fmt.Sprintf("%s%d", prefix, g.lastCount)
	g.lastCount++

	// Generate dates
	now := time.Now()
	minDate := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	signupDate := GenerateRandomDate(g.rand, minDate, now)

	var cancellationDate *time.Time

	// 15% chance member is inactive
	var status string
	if g.rand.Float64() < 0.15 {
		status = "inactive"
		// Corregido: crear una variable y luego asignar su dirección
		cancelDate := GenerateRandomDate(g.rand, signupDate.AddDate(0, 3, 0), now)
		cancellationDate = &cancelDate
	} else {
		status = "active"
	}

	// Birthday - between 18 and 80 years ago
	birthYear := now.Year() - g.rand.Intn(62) - 18
	birthMonth := time.Month(g.rand.Intn(12) + 1)
	birthDay := g.rand.Intn(28) + 1
	birthDate := time.Date(birthYear, birthMonth, birthDay, 0, 0, 0, 0, time.UTC)

	email := GenerateRandomEmail(g.rand, firstName, lastName)
	profession := GenerateRandomProfession(g.rand)

	// Generate member
	return Member{
		MembershipNumber: memberNumber,
		MembershipType:   membershipType,
		Name:             firstName,
		Surnames:         lastName,
		Address:          address,
		Postcode:         postalCode,
		City:             city,
		Province:         province,
		Country:          "España",
		State:            status,
		RegistrationDate: signupDate,
		LeavingDate:      cancellationDate,
		BirthDate:        &birthDate,
		IdentityCard:     documentID,
		Email:            email,
		Profession:       profession,
		Nationality:      nationality,
		Remarks:          "",
	}
}

// GetLastInsertedMembers retrieves the last n members inserted into the database
func (g *MemberGenerator) GetLastInsertedMembers(ctx context.Context, n int) ([]Member, error) {
	var members []Member

	query := `
		SELECT 
			id, membership_number, membership_type, name, surnames,
			address, postcode, city, province, country,
			state, registration_date, leaving_date, birth_date,
			identity_card, email, profession, nationality, remarks
		FROM members
		ORDER BY id DESC
		LIMIT $1
	`

	err := g.db.SelectContext(ctx, &members, query, n)
	if err != nil {
		return nil, fmt.Errorf("failed to get last inserted members: %w", err)
	}

	return members, nil
}

// FindIndividualMembers finds members with individual membership
func (g *MemberGenerator) FindIndividualMembers(ctx context.Context, limit int) ([]Member, error) {
	var members []Member

	query := `
		SELECT 
			id, membership_number, membership_type, name, surnames,
			address, postcode, city, province, country,
			state, registration_date, leaving_date, birth_date,
			identity_card, email, profession, nationality, remarks
		FROM members
		WHERE membership_type = 'individual' AND state = 'active'
		LIMIT $1
	`

	err := g.db.SelectContext(ctx, &members, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find individual members: %w", err)
	}

	return members, nil
}

// FindFamilyMembers finds members with family membership
func (g *MemberGenerator) FindFamilyMembers(ctx context.Context, limit int) ([]Member, error) {
	var members []Member

	query := `
		SELECT 
			id, membership_number, membership_type, name, surnames,
			address, postcode, city, province, country,
			state, registration_date, leaving_date, birth_date,
			identity_card, email, profession, nationality, remarks
		FROM members
		WHERE membership_type = 'familiar' AND state = 'active'
		LIMIT $1
	`

	err := g.db.SelectContext(ctx, &members, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find family members: %w", err)
	}

	return members, nil
}
