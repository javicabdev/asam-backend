package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"gorm.io/gorm"
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
		return fmt.Errorf("error creating member: %w", result.Error)
	}
	return nil
}

// GetByID busca un miembro por su ID
func (r *memberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).First(&member, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting member by ID: %w", result.Error)
	}
	return &member, nil
}

// GetByNumeroSocio busca un miembro por su número de socio
func (r *memberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	var member models.Member
	result := r.db.WithContext(ctx).Where("numero_socio = ?", numeroSocio).First(&member)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting member by numero socio: %w", result.Error)
	}
	return &member, nil
}

// Update actualiza un miembro existente
func (r *memberRepository) Update(ctx context.Context, member *models.Member) error {
	result := r.db.WithContext(ctx).Save(member)
	if result.Error != nil {
		return fmt.Errorf("error updating member: %w", result.Error)
	}
	return nil
}

// Delete elimina un miembro por su ID
func (r *memberRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Member{}, id)
	if result.Error != nil {
		return fmt.Errorf("error deleting member: %w", result.Error)
	}
	return nil
}

// List obtiene una lista de miembros según los filtros proporcionados
func (r *memberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	var members []models.Member
	query := r.db.WithContext(ctx)

	// Aplicar filtros
	if filters.Estado != nil {
		query = query.Where("estado = ?", *filters.Estado)
	}

	if filters.TipoMembresia != nil {
		query = query.Where("tipo_membresia = ?", *filters.TipoMembresia)
	}

	if filters.SearchTerm != nil {
		searchTerm := "%" + *filters.SearchTerm + "%"
		query = query.Where(
			"numero_socio ILIKE ? OR nombre ILIKE ? OR apellidos ILIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Aplicar paginación
	if filters.Page > 0 && filters.PageSize > 0 {
		offset := (filters.Page - 1) * filters.PageSize
		query = query.Offset(offset).Limit(filters.PageSize)
	}

	result := query.Find(&members)
	if result.Error != nil {
		return nil, fmt.Errorf("error listing members: %w", result.Error)
	}

	return members, nil
}
