package input

import (
	"context"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// FamilyService define las operaciones de negocio disponibles para familias
type FamilyService interface {
	Create(ctx context.Context, family *models.Family) error
	Update(ctx context.Context, family *models.Family) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Family, error)
	GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error)
	GetByOriginMemberID(ctx context.Context, memberID uint) (*models.Family, error)
	List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error)
	AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error
	UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error
	RemoveFamiliar(ctx context.Context, familiarID uint) error
	GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error)

	// CreateFamilyAtomic creates a family with optional member and familiares in a single transaction
	CreateFamilyAtomic(ctx context.Context, req *CreateFamilyAtomicRequest) (*models.Family, error)
}
