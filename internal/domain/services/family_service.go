package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"strings"
)

var (
	ErrFamilyNotFound      = errors.New("familia no encontrada")
	ErrInvalidFamilyData   = errors.New("datos de familia inválidos")
	ErrInvalidFamiliarData = errors.New("datos de familiar inválidos")
	ErrFamiliarNotFound    = errors.New("familiar no encontrado")
	ErrMemberNotFound      = errors.New("miembro no encontrado")
)

type familyService struct {
	familyRepo output.FamilyRepository
	memberRepo output.MemberRepository
}

// NewFamilyService crea una nueva instancia del servicio
func NewFamilyService(
	familyRepo output.FamilyRepository,
	memberRepo output.MemberRepository,
) input.FamilyService {
	return &familyService{
		familyRepo: familyRepo,
		memberRepo: memberRepo,
	}
}

// Create crea una nueva familia
func (s *familyService) Create(ctx context.Context, family *models.Family) error {
	// Validar datos de la familia
	if err := family.Validate(); err != nil {
		return ErrInvalidFamilyData
	}

	// Si hay miembro origen, verificar que existe
	if family.MiembroOrigenID != nil {
		member, err := s.memberRepo.GetByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return err
		}
		if member == nil {
			return ErrMemberNotFound
		}
	}

	// Crear la familia
	if err := s.familyRepo.Create(ctx, family); err != nil {
		return err
	}

	return nil
}

// Update actualiza una familia existente
func (s *familyService) Update(ctx context.Context, family *models.Family) error {
	// Verificar que la familia existe
	existingFamily, err := s.familyRepo.GetByID(ctx, family.ID)
	if err != nil {
		return err
	}
	if existingFamily == nil {
		return ErrFamilyNotFound
	}

	// Validar datos actualizados
	if err := family.Validate(); err != nil {
		return ErrInvalidFamilyData
	}

	// Si cambia el miembro origen, verificar que existe
	if family.MiembroOrigenID != nil &&
		(existingFamily.MiembroOrigenID == nil ||
			*existingFamily.MiembroOrigenID != *family.MiembroOrigenID) {

		member, err := s.memberRepo.GetByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return err
		}
		if member == nil {
			return ErrMemberNotFound
		}
	}

	// Actualizar la familia
	return s.familyRepo.Update(ctx, family)
}

// Delete elimina una familia
func (s *familyService) Delete(ctx context.Context, id uint) error {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if family == nil {
		return ErrFamilyNotFound
	}

	return s.familyRepo.Delete(ctx, id)
}

// GetByID obtiene una familia por su ID
func (s *familyService) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	family, err := s.familyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if family == nil {
		return nil, ErrFamilyNotFound
	}
	return family, nil
}

// GetByNumeroSocio obtiene una familia por su número de socio
func (s *familyService) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	return s.familyRepo.GetByNumeroSocio(ctx, numeroSocio)
}

// List obtiene una lista paginada de familias
func (s *familyService) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	// Validaciones básicas
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	// Validar ordenamiento si se proporciona
	if orderBy != "" {
		// Validar que el campo de ordenamiento sea válido
		validFields := map[string]bool{
			"numero_socio":     true,
			"esposo_nombre":    true,
			"esposo_apellidos": true,
			"esposa_nombre":    true,
			"esposa_apellidos": true,
		}

		// Extraer el campo de ordenamiento (quitando el ASC/DESC)
		parts := strings.Fields(orderBy)
		if !validFields[strings.ToLower(parts[0])] {
			return nil, 0, fmt.Errorf("campo de ordenamiento inválido: %s", parts[0])
		}
	}

	// Llamar al repositorio con los parámetros validados
	families, total, err := s.familyRepo.List(ctx, page, pageSize, searchTerm, orderBy)
	if err != nil {
		return nil, 0, fmt.Errorf("error al listar familias: %w", err)
	}

	return families, total, nil
}

// AddFamiliar añade un nuevo familiar a una familia
func (s *familyService) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, familyID)
	if err != nil {
		return err
	}
	if family == nil {
		return ErrFamilyNotFound
	}

	// Validar datos del familiar
	if err := familiar.Validate(); err != nil {
		return ErrInvalidFamiliarData
	}

	return s.familyRepo.AddFamiliar(ctx, familyID, familiar)
}

// UpdateFamiliar actualiza un familiar existente
func (s *familyService) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	// Validar datos del familiar
	if err := familiar.Validate(); err != nil {
		return ErrInvalidFamiliarData
	}

	return s.familyRepo.UpdateFamiliar(ctx, familiar)
}

// RemoveFamiliar elimina un familiar
func (s *familyService) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	return s.familyRepo.RemoveFamiliar(ctx, familiarID)
}

// GetFamiliares obtiene todos los familiares de una familia
func (s *familyService) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, familyID)
	if err != nil {
		return nil, err
	}
	if family == nil {
		return nil, ErrFamilyNotFound
	}

	return s.familyRepo.GetFamiliares(ctx, familyID)
}
