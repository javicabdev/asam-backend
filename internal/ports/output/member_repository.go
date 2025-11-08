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
	GetByIdentityCard(ctx context.Context, identityCard string) (*models.Member, error)
	Update(ctx context.Context, member *models.Member) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filters MemberFilters) ([]models.Member, int, error) // Returns members, total count, error
	GetLastMemberNumberByPrefix(ctx context.Context, prefix string) (string, error)
	SearchWithoutUser(ctx context.Context, criteria string) ([]models.Member, error)
	GetAllActive(ctx context.Context) ([]*models.Member, error)

	// Transaction support
	BeginTransaction(ctx context.Context) (Transaction, error)
	CreateWithTx(ctx context.Context, tx Transaction, member *models.Member) error
	GetByIDWithTx(ctx context.Context, tx Transaction, id uint) (*models.Member, error)
	GetByNumeroSocioWithTx(ctx context.Context, tx Transaction, numeroSocio string) (*models.Member, error)
	GetByIdentityCardWithTx(ctx context.Context, tx Transaction, identityCard string) (*models.Member, error)
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
