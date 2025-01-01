package output

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// FamilyRepository define las operaciones disponibles para persistencia de familias
type FamilyRepository interface {
	// Operaciones principales
	Create(ctx context.Context, family *models.Family) error
	GetByID(ctx context.Context, id uint) (*models.Family, error)
	Update(ctx context.Context, family *models.Family) error
	Delete(ctx context.Context, id uint) error

	// Operaciones de búsqueda
	GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error)
	List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error)

	// Operaciones de familiares
	AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error
	UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error
	RemoveFamiliar(ctx context.Context, familiarID uint) error
	GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error)
}

type FamilyFilters struct {
	SearchTerm *string
	Page       int
	PageSize   int
	OrderBy    string // Añadido
}
