package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Reemplazamos las variables de error estándar con constantes para centralizar los mensajes
const (
	msgFamilyNotFound      = "familia no encontrada"
	msgInvalidFamilyData   = "datos de familia inválidos"
	msgInvalidFamiliarData = "datos de familiar inválidos"
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

// GetByOriginMemberID obtiene una familia por el ID del miembro origen
func (s *familyService) GetByOriginMemberID(ctx context.Context, memberID uint) (*models.Family, error) {
	family, err := s.familyRepo.GetByOriginMemberID(ctx, memberID)
	if err != nil {
		return nil, errors.DB(err, "error obteniendo familia por ID de miembro origen")
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

// CreateFamilyAtomic creates a family with optional member and familiares in a single atomic transaction
func (s *familyService) CreateFamilyAtomic(ctx context.Context, req *input.CreateFamilyAtomicRequest) (*models.Family, error) {
	// 1. Validación previa a transacción

	// Validar que el DNI del esposo sea obligatorio para el miembro origen
	if req.CreateMemberIfNotExists && req.Family.EsposoDocumentoIdentidad == "" {
		return nil, errors.NewValidationError(
			"El documento de identidad del esposo es obligatorio",
			map[string]string{"esposoDocumentoIdentidad": "El documento de identidad es obligatorio para el miembro principal"},
		)
	}

	if err := req.Family.Validate(); err != nil {
		return nil, err
	}

	for i, fam := range req.Familiares {
		if err := fam.Validate(); err != nil {
			return nil, errors.Wrap(err, errors.ErrValidationFailed,
				fmt.Sprintf("Familiar at index %d is invalid", i))
		}
	}

	// 2. Iniciar transacción
	tx, err := s.familyRepo.BeginTransaction(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	// 3. Manejar Member origen
	var originMemberID uint
	if req.Family.MiembroOrigenID != nil {
		originMemberID = *req.Family.MiembroOrigenID
		// Verificar que existe y es válido
		member, err := s.memberRepo.GetByIDWithTx(ctx, tx, originMemberID)
		if err != nil || member == nil {
			_ = tx.Rollback()
			return nil, errors.New(errors.ErrNotFound, "Origin member not found")
		}
		if member.MembershipType != models.TipoMembresiaPFamiliar {
			_ = tx.Rollback()
			return nil, errors.New(errors.ErrInvalidOperation,
				"Origin member must be of type 'familiar'")
		}
	} else if req.CreateMemberIfNotExists {
		// Verificar si ya existe un miembro con este numero_socio
		existingMember, err := s.memberRepo.GetByNumeroSocioWithTx(ctx, tx, req.Family.NumeroSocio)
		if err == nil && existingMember != nil {
			// Ya existe un miembro con este número
			if existingMember.MembershipType != models.TipoMembresiaPFamiliar {
				_ = tx.Rollback()
				return nil, errors.New(errors.ErrInvalidOperation,
					"Existing member with this number is not of type 'familiar'")
			}
			// Verificar que no tenga ya una familia asociada
			existingFamily, _ := s.familyRepo.GetByOriginMemberID(ctx, existingMember.ID)
			if existingFamily != nil {
				_ = tx.Rollback()
				return nil, errors.New(errors.ErrDuplicateEntry,
					"Member already has a family associated")
			}
			originMemberID = existingMember.ID
		} else {
			// Verificar que no exista otro miembro con el mismo DNI
			// Nota: La validación de DNI duplicado se maneja mediante la constraint UNIQUE en la BD.
			// Si se requiere validación previa, añadir función GetByIdentityCardWithTx al repositorio.

			// Crear nuevo Member con DNI
			member := &models.Member{
				MembershipNumber: req.Family.NumeroSocio,
				MembershipType:   models.TipoMembresiaPFamiliar,
				Name:             req.Family.EsposoNombre,
				Surnames:         req.Family.EsposoApellidos,
				State:            models.EstadoActivo,
				RegistrationDate: time.Now(),
			}

			// Añadir documento de identidad (obligatorio)
			if req.Family.EsposoDocumentoIdentidad != "" {
				dni := req.Family.EsposoDocumentoIdentidad
				member.IdentityCard = &dni
			}

			// Añadir email si está disponible
			if req.Family.EsposoCorreoElectronico != "" {
				email := req.Family.EsposoCorreoElectronico
				member.Email = &email
			}

			// Añadir fecha de nacimiento si está disponible
			if req.Family.EsposoFechaNacimiento != nil {
				member.BirthDate = req.Family.EsposoFechaNacimiento
			}

			if req.MemberData != nil {
				member.Address = req.MemberData.Address
				member.Postcode = req.MemberData.Postcode
				member.City = req.MemberData.City
				member.Province = req.MemberData.Province
				if req.MemberData.Province == "" {
					member.Province = "Barcelona"
				}
				member.Country = req.MemberData.Country
				if req.MemberData.Country == "" {
					member.Country = "España"
				}
			}

			if err := s.memberRepo.CreateWithTx(ctx, tx, member); err != nil {
				_ = tx.Rollback()
				// Si es error de duplicado, dar mensaje claro
				appErr, isAppErr := errors.AsAppError(err)
				if isAppErr && appErr.Code == errors.ErrDuplicateEntry {
					return nil, errors.NewValidationError(
						"El documento de identidad ya está registrado",
						map[string]string{"esposoDocumentoIdentidad": "Ya existe un miembro con este documento de identidad"},
					)
				}
				return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to create origin member")
			}

			originMemberID = member.ID
		}

		req.Family.MiembroOrigenID = &originMemberID
	}

	// 4. Crear Family
	if err := s.familyRepo.CreateWithTx(ctx, tx, req.Family); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to create family")
	}

	// 5. Crear Familiares
	for i, fam := range req.Familiares {
		fam.FamiliaID = req.Family.ID
		if err := s.familyRepo.AddFamiliarWithTx(ctx, tx, req.Family.ID, fam); err != nil {
			_ = tx.Rollback()
			return nil, errors.Wrap(err, errors.ErrDatabaseError,
				fmt.Sprintf("Failed to create familiar at index %d", i))
		}
	}

	// 6. Commit
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to commit transaction")
	}

	// 7. Recargar con relaciones
	family, err := s.familyRepo.GetByID(ctx, req.Family.ID)
	if err != nil {
		// No fallamos si falla la recarga, retornamos la familia creada
		return req.Family, nil
	}

	return family, nil
}
