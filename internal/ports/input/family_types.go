package input

import "github.com/javicabdev/asam-backend/internal/domain/models"

// CreateFamilyAtomicRequest encapsulates all data needed for atomic family creation
type CreateFamilyAtomicRequest struct {
	// Family data
	Family *models.Family

	// Optional: Create new member if not provided in Family.MiembroOrigenID
	CreateMemberIfNotExists bool
	MemberData              *CreateMemberData

	// Optional: Additional family members
	Familiares []*models.Familiar
}

// CreateMemberData contains data for creating the origin member
type CreateMemberData struct {
	Address   string
	Postcode  string
	City      string
	Province  string
	Country   string
	Telefonos []models.Telephone
}
