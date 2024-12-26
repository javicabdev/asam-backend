package db

import (
	"context"
	"errors"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"gorm.io/gorm"
)

type familyRepository struct {
	db *gorm.DB
}

// NewFamilyRepository crea una nueva instancia del repositorio
func NewFamilyRepository(db *gorm.DB) *familyRepository {
	return &familyRepository{
		db: db,
	}
}

// Create inserta una nueva familia en la base de datos
func (r *familyRepository) Create(ctx context.Context, family *models.Family) error {
	result := r.db.WithContext(ctx).Create(family)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetByID obtiene una familia por su ID
func (r *familyRepository) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	var family models.Family
	result := r.db.WithContext(ctx).
		Preload("Familiares").
		Preload("Telefonos").
		First(&family, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &family, nil
}

// GetByNumeroSocio obtiene una familia por su número de socio
func (r *familyRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	var family models.Family
	result := r.db.WithContext(ctx).
		Preload("Familiares").
		Preload("Telefonos").
		Where("numero_socio = ?", numeroSocio).
		First(&family)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &family, nil
}

// Update actualiza los datos de una familia existente
func (r *familyRepository) Update(ctx context.Context, family *models.Family) error {
	result := r.db.WithContext(ctx).Save(family)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Delete elimina una familia (soft delete)
func (r *familyRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Family{}, id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// List obtiene una lista paginada de familias
func (r *familyRepository) List(ctx context.Context, offset, limit int) ([]*models.Family, int, error) {
	var families []*models.Family
	var total int64

	// Obtener el total de registros
	if err := r.db.WithContext(ctx).Model(&models.Family{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Obtener los registros paginados
	result := r.db.WithContext(ctx).
		Preload("Familiares").
		Preload("Telefonos").
		Offset(offset).
		Limit(limit).
		Find(&families)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return families, int(total), nil
}

// Operaciones de familiares

// AddFamiliar añade un nuevo familiar a una familia
func (r *familyRepository) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	familiar.FamiliaID = familyID
	result := r.db.WithContext(ctx).Create(familiar)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateFamiliar actualiza los datos de un familiar
func (r *familyRepository) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	result := r.db.WithContext(ctx).Save(familiar)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveFamiliar elimina un familiar
func (r *familyRepository) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Familiar{}, familiarID)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetFamiliares obtiene todos los familiares de una familia
func (r *familyRepository) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	var familiares []*models.Familiar
	result := r.db.WithContext(ctx).
		Where("familia_id = ?", familyID).
		Find(&familiares)

	if result.Error != nil {
		return nil, result.Error
	}
	return familiares, nil
}
