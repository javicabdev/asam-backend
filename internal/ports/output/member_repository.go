package output

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// MemberRepository define las operaciones disponibles para la persistencia de miembros
type MemberRepository interface {
	Create(ctx context.Context, member *models.Member) error
	GetByID(ctx context.Context, id uint) (*models.Member, error)
	GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error)
	Update(ctx context.Context, member *models.Member) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filters MemberFilters) ([]models.Member, error)
	GetLastMemberNumberByPrefix(ctx context.Context, prefix string) (string, error)
}

// MemberFilters define los filtros disponibles para buscar miembros
type MemberFilters struct {
	Estado        *string
	TipoMembresia *string
	SearchTerm    *string // Para búsqueda por nombre, apellidos o número de socio
	Page          int
	PageSize      int
	OrderBy       string
}
