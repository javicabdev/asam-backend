package input

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// FamilyService define las operaciones de negocio disponibles para familias
type FamilyService interface {
	// Operaciones de familia
	Create(ctx context.Context, family *models.Family) error
	Update(ctx context.Context, family *models.Family) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*models.Family, error)
	GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error)
	List(ctx context.Context, page, pageSize int) ([]*models.Family, int, error)

	// Operaciones de familiares
	AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error
	UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error
	RemoveFamiliar(ctx context.Context, familiarID uint) error
	GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error)
}
