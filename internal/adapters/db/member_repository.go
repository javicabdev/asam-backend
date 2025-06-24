package db

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type memberRepository struct {
	db *gorm.DB
}

// NewMemberRepository crea una nueva instancia del repositorio
func NewMemberRepository(db *gorm.DB) output.MemberRepository {
	return &memberRepository{db: db}
}

// Create crea un nuevo miembro en la base de datos
func (r *memberRepository) Create(ctx context.Context, member *models.Member) error {
	result := r.db.WithContext(ctx).Create(member)
	if result.Error != nil {
		// Check for specific database errors
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "member with the same key already exists")
		}
		return appErrors.DB(result.Error, "error creating member")
	}
	return nil
}

// GetByID busca un miembro por su ID
func (r *memberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).First(&member, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Consistently return nil, nil for not found
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by ID")
	}

	return &member, nil
}

// GetByNumeroSocio busca un miembro por su número de socio
func (r *memberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).Where("membership_number = ?", numeroSocio).First(&member)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Consistently return nil, nil for not found
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error getting member by numero socio")
	}

	return &member, nil
}

// Update actualiza un miembro existente
func (r *memberRepository) Update(ctx context.Context, member *models.Member) error {
	result := r.db.WithContext(ctx).Save(member)
	if result.Error != nil {
		// Check for specific database errors
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("member", result.Error)
		}
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "member with the same key already exists")
		}
		return appErrors.DB(result.Error, "error updating member")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("member", nil)
	}

	return nil
}

// Delete elimina un miembro por su ID
func (r *memberRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Member{}, id)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error deleting member")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("member", nil)
	}

	return nil
}

// List obtiene una lista de miembros según los filtros proporcionados
func (r *memberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	var members []models.Member
	query := r.db.WithContext(ctx)

	// Aplicar filtros
	if filters.Estado != nil {
		query = query.Where("state = ?", *filters.Estado)
	}

	if filters.TipoMembresia != nil {
		query = query.Where("membership_type = ?", *filters.TipoMembresia)
	}

	if filters.SearchTerm != nil {
		searchTerm := "%" + *filters.SearchTerm + "%"
		query = query.Where(
			"membership_number ILIKE ? OR name ILIKE ? OR surnames ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Aplicar ordenamiento
	if filters.OrderBy != "" {
		query = query.Order(filters.OrderBy)
	}

	// Aplicar paginación
	if filters.Page > 0 && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	result := query.Find(&members)
	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error listing members")
	}

	return members, nil
}
