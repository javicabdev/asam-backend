package input

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// MemberService define las operaciones de negocio disponibles para la gestión de miembros
type MemberService interface {
	// CreateMember crea un nuevo miembro
	CreateMember(ctx context.Context, member *models.Member) error

	// GetMemberByID obtiene un miembro por su ID
	GetMemberByID(ctx context.Context, id uint) (*models.Member, error)

	// GetMemberByNumeroSocio obtiene un miembro por su número de socio
	GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error)

	// UpdateMember actualiza los datos de un miembro
	UpdateMember(ctx context.Context, member *models.Member) error

	// DeactivateMember da de baja a un miembro
	DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error

	// ListMembers obtiene una lista de miembros según los criterios especificados
	ListMembers(ctx context.Context, filters MemberFilters) ([]*models.Member, error)

	// GetNextMemberNumber obtiene el siguiente número de socio disponible según el tipo
	GetNextMemberNumber(ctx context.Context, isFamily bool) (string, error)

	// CheckMemberNumberExists verifica si un número de socio ya existe
	CheckMemberNumberExists(ctx context.Context, memberNumber string) (bool, error)

	// SearchMembersWithoutUser busca miembros que no tienen usuario asociado
	SearchMembersWithoutUser(ctx context.Context, criteria string) ([]*models.Member, error)
}

// MemberFilters defines search criteria for members
type MemberFilters struct {
	State          *string
	MembershipType *string
	SearchTerm     *string
	Page           int
	PageSize       int
	OrderBy        string
}

// CreateMemberRequest represents the data necessary to create a new member
type CreateMemberRequest struct {
	MembershipNumber string     `json:"membershipNumber"`
	MembershipType   string     `json:"membershipType"`
	Name             string     `json:"name"`
	Surnames         string     `json:"surnames"`
	Address          string     `json:"address"`
	Postcode         string     `json:"postcode"`
	City             string     `json:"city"`
	Province         string     `json:"province"`
	Country          string     `json:"country"`
	RegistrationDate time.Time  `json:"registrationDate"`
	BirthDate        *time.Time `json:"birthDate,omitempty"`
	IdentityCard     *string    `json:"identityCard,omitempty"`
	Email            *string    `json:"email,omitempty"`
	Profession       *string    `json:"profession,omitempty"`
	Nationality      string     `json:"nationality"`
	Remarks          *string    `json:"remarks,omitempty"`
}

// UpdateMemberRequest represents the updateable data of a member
type UpdateMemberRequest struct {
	Address      *string `json:"address,omitempty"`
	Postcode     *string `json:"postcode,omitempty"`
	City         *string `json:"city,omitempty"`
	Province     *string `json:"province,omitempty"`
	Country      *string `json:"country,omitempty"`
	IdentityCard *string `json:"identityCard,omitempty"`
	Email        *string `json:"email,omitempty"`
	Profession   *string `json:"profession,omitempty"`
	Remarks      *string `json:"remarks,omitempty"`
}
