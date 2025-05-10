package output

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// FamilyRepository defines the operations available for family persistence
type FamilyRepository interface {
	Create(ctx context.Context, family *models.Family) error
	GetByID(ctx context.Context, id uint) (*models.Family, error)
	Update(ctx context.Context, family *models.Family) error
	Delete(ctx context.Context, id uint) error

	GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error)
	List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error)

	AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error
	UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error
	RemoveFamiliar(ctx context.Context, familiarID uint) error
	GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error)
}

type FamilyFilters struct {
	SearchTerm *string
	Page       int
	PageSize   int
	OrderBy    string
}
