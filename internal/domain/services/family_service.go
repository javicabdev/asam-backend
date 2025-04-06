package services

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"strings"
)

// Reemplazamos las variables de error estándar con constantes para centralizar los mensajes
const (
	msgFamilyNotFound      = "familia no encontrada"
	msgInvalidFamilyData   = "datos de familia inválidos"
	msgInvalidFamiliarData = "datos de familiar inválidos"
	msgFamiliarNotFound    = "familiar no encontrado"
	msgMemberNotFound      = "miembro no encontrado"
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
		return errors.NewValidationError(msgInvalidFamilyData, nil)
	}

	// Si hay miembro origen, verificar que existe
	if family.MiembroOrigenID != nil {
		member, err := s.memberRepo.GetByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return errors.DB(err, "error verificando miembro origen")
		}
		if member == nil {
			return errors.NotFound(msgMemberNotFound, nil)
		}
	}

	// Crear la familia
	if err := s.familyRepo.Create(ctx, family); err != nil {
		return errors.DB(err, "error creando familia")
	}

	return nil
}

// Update actualiza una familia existente
func (s *familyService) Update(ctx context.Context, family *models.Family) error {
	// Verificar que la familia existe
	existingFamily, err := s.familyRepo.GetByID(ctx, family.ID)
	if err != nil {
		return errors.DB(err, "error verificando existencia de familia")
	}
	if existingFamily == nil {
		return errors.NotFound(msgFamilyNotFound, nil)
	}

	// Validar datos actualizados
	if err := family.Validate(); err != nil {
		return errors.NewValidationError(msgInvalidFamilyData, nil)
	}

	// Si cambia el miembro origen, verificar que existe
	if family.MiembroOrigenID != nil &&
		(existingFamily.MiembroOrigenID == nil ||
			*existingFamily.MiembroOrigenID != *family.MiembroOrigenID) {

		member, err := s.memberRepo.GetByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return errors.DB(err, "error verificando miembro origen")
		}
		if member == nil {
			return errors.NotFound(msgMemberNotFound, nil)
		}
	}

	// Actualizar la familia
	if err := s.familyRepo.Update(ctx, family); err != nil {
		return errors.DB(err, "error actualizando familia")
	}
	return nil
}

// Delete elimina una familia
func (s *familyService) Delete(ctx context.Context, id uint) error {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, id)
	if err != nil {
		return errors.DB(err, "error verificando existencia de familia")
	}
	if family == nil {
		return errors.NotFound(msgFamilyNotFound, nil)
	}

	if err := s.familyRepo.Delete(ctx, id); err != nil {
		return errors.DB(err, "error eliminando familia")
	}
	return nil
}

// GetByID obtiene una familia por su ID
func (s *familyService) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	family, err := s.familyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo familia por ID")
	}
	if family == nil {
		return nil, errors.NotFound(msgFamilyNotFound, nil)
	}
	return family, nil
}

// GetByNumeroSocio obtiene una familia por su número de socio
func (s *familyService) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	family, err := s.familyRepo.GetByNumeroSocio(ctx, numeroSocio)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo familia por número de socio")
	}
	if family == nil {
		return nil, errors.NotFound(msgFamilyNotFound, nil)
	}
	return family, nil
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
			return nil, 0, errors.Validation(
				"campo de ordenamiento inválido",
				"orderBy",
				parts[0],
			)
		}
	}

	// Llamar al repositorio con los parámetros validados
	families, total, err := s.familyRepo.List(ctx, page, pageSize, searchTerm, orderBy)
	if err != nil {
		return nil, 0, errors.DB(err, "error al listar familias")
	}

	return families, total, nil
}

// AddFamiliar añade un nuevo familiar a una familia
func (s *familyService) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, familyID)
	if err != nil {
		return errors.DB(err, "error verificando existencia de familia")
	}
	if family == nil {
		return errors.NotFound(msgFamilyNotFound, nil)
	}

	// Validar datos del familiar
	if err := familiar.Validate(); err != nil {
		return errors.NewValidationError(msgInvalidFamiliarData, nil)
	}

	if err := s.familyRepo.AddFamiliar(ctx, familyID, familiar); err != nil {
		return errors.DB(err, "error añadiendo familiar")
	}
	return nil
}

// UpdateFamiliar actualiza un familiar existente
func (s *familyService) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	// Validar datos del familiar
	if err := familiar.Validate(); err != nil {
		return errors.NewValidationError(msgInvalidFamiliarData, nil)
	}

	if err := s.familyRepo.UpdateFamiliar(ctx, familiar); err != nil {
		return errors.DB(err, "error actualizando familiar")
	}
	return nil
}

// RemoveFamiliar elimina un familiar
func (s *familyService) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	if err := s.familyRepo.RemoveFamiliar(ctx, familiarID); err != nil {
		return errors.DB(err, "error eliminando familiar")
	}
	return nil
}

// GetFamiliares obtiene todos los familiares de una familia
func (s *familyService) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	// Verificar que la familia existe
	family, err := s.familyRepo.GetByID(ctx, familyID)
	if err != nil {
		return nil, errors.DB(err, "error verificando existencia de familia")
	}
	if family == nil {
		return nil, errors.NotFound(msgFamilyNotFound, nil)
	}

	familiares, err := s.familyRepo.GetFamiliares(ctx, familyID)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo familiares")
	}

	return familiares, nil
}
