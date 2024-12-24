package input

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"time"
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
	ListMembers(ctx context.Context, filters MemberFilters) ([]models.Member, error)
}

// MemberFilters define los criterios de búsqueda para miembros
type MemberFilters struct {
	Estado        *string
	TipoMembresia *string
	SearchTerm    *string
	Page          int
	PageSize      int
}

// CreateMemberRequest representa los datos necesarios para crear un nuevo miembro
type CreateMemberRequest struct {
	NumeroSocio        string     `json:"numeroSocio"`
	TipoMembresia      string     `json:"tipoMembresia"`
	Nombre             string     `json:"nombre"`
	Apellidos          string     `json:"apellidos"`
	CalleNumeroPiso    string     `json:"calleNumeroPiso"`
	CodigoPostal       string     `json:"codigoPostal"`
	Poblacion          string     `json:"poblacion"`
	Provincia          string     `json:"provincia"`
	Pais               string     `json:"pais"`
	FechaAlta          time.Time  `json:"fechaAlta"`
	FechaNacimiento    *time.Time `json:"fechaNacimiento,omitempty"`
	DocumentoIdentidad *string    `json:"documentoIdentidad,omitempty"`
	CorreoElectronico  *string    `json:"correoElectronico,omitempty"`
	Profesion          *string    `json:"profesion,omitempty"`
	Nacionalidad       string     `json:"nacionalidad"`
	Observaciones      *string    `json:"observaciones,omitempty"`
}

// UpdateMemberRequest representa los datos actualizables de un miembro
type UpdateMemberRequest struct {
	CalleNumeroPiso    *string `json:"calleNumeroPiso,omitempty"`
	CodigoPostal       *string `json:"codigoPostal,omitempty"`
	Poblacion          *string `json:"poblacion,omitempty"`
	Provincia          *string `json:"provincia,omitempty"`
	Pais               *string `json:"pais,omitempty"`
	DocumentoIdentidad *string `json:"documentoIdentidad,omitempty"`
	CorreoElectronico  *string `json:"correoElectronico,omitempty"`
	Profesion          *string `json:"profesion,omitempty"`
	Observaciones      *string `json:"observaciones,omitempty"`
}
