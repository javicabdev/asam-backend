package services

import (
	"context"
	"strconv"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
	"go.uber.org/zap"
)

type memberService struct {
	repository  output.MemberRepository
	appLogger   logger.Logger
	auditLogger audit.Logger
}

// NewMemberService crea una nueva instancia del servicio de miembros
func NewMemberService(repository output.MemberRepository, appLogger logger.Logger, auditLogger audit.Logger) input.MemberService {
	return &memberService{
		repository:  repository,
		appLogger:   appLogger,
		auditLogger: auditLogger,
	}
}

// CreateMember implementa la lógica de creación de un nuevo miembro
func (s *memberService) CreateMember(ctx context.Context, member *models.Member) error {
	// Logging al inicio de la operación
	s.appLogger.Info("Creating new member",
		zap.String("numero_socio", member.MembershipNumber),
		zap.String("tipo_membresia", member.MembershipType))

	// Verificar si ya existe un miembro con el mismo número de socio
	existing, err := s.repository.GetByNumeroSocio(ctx, member.MembershipNumber)
	if err != nil {
		s.appLogger.Error("Error checking existing member",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al verificar miembro existente", err)
		return errors.DB(err, "error verificando miembro existente")
	}

	if existing != nil {
		s.appLogger.Warn("Attempted to create duplicate member",
			zap.String("numero_socio", member.MembershipNumber))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Intento de crear miembro duplicado", nil)
		return errors.New(errors.ErrDuplicateEntry,
			"ya existe un miembro con el número de socio "+member.MembershipNumber)
	}

	// Establecer valores por defecto
	if member.State == "" {
		member.State = models.EstadoActivo
	}
	if member.Province == "" {
		member.Province = "Barcelona"
	}
	if member.Country == "" {
		member.Country = "España"
	}
	if member.Nationality == "" {
		member.Nationality = "Senegal"
	}

	// Validar el miembro antes de crear
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error en la validación del miembro", err)

		// Conservar el error de validación si ya es un AppError, sino convertirlo
		appErr, ok := errors.AsAppError(err)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", err.Error())
	}

	// Crear el miembro en la base de datos
	if err := s.repository.Create(ctx, member); err != nil {
		s.appLogger.Error("Failed to create member",
			zap.String("numero_socio", member.MembershipNumber),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.MembershipNumber,
			"Error al crear miembro en base de datos", err)
		return errors.DB(err, "error creando miembro")
	}

	// Actualizar métricas de miembros
	metrics.MembersByStatus.WithLabelValues(
		member.State,
		member.MembershipType,
	).Inc()

	// Registrar la acción en el log de auditoría
	s.auditLogger.LogAction(ctx,
		audit.ActionCreate,
		audit.EntityMember,
		member.MembershipNumber,
		"Created new member")

	s.appLogger.Info("Member created successfully",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))

	return nil
}

// GetMemberByID obtiene un miembro por su ID
func (s *memberService) GetMemberByID(ctx context.Context, id uint) (*models.Member, error) {
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.appLogger.Error("Error getting member by ID",
			zap.Uint("id", id),
			zap.Error(err))
		return nil, errors.DB(err, "error obteniendo miembro por ID")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	return member, nil
}

// GetMemberByNumeroSocio obtiene un miembro por su número de socio
func (s *memberService) GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	member, err := s.repository.GetByNumeroSocio(ctx, numeroSocio)
	if err != nil {
		s.appLogger.Error("Error getting member by numero socio",
			zap.String("numero_socio", numeroSocio),
			zap.Error(err))
		return nil, errors.DB(err, "error obteniendo miembro por numero socio")
	}

	if member == nil {
		return nil, errors.NotFound("member", nil)
	}

	return member, nil
}

// UpdateMember actualiza los datos de un miembro existente
func (s *memberService) UpdateMember(ctx context.Context, member *models.Member) error {
	// Verificar que el miembro existe
	existing, err := s.repository.GetByID(ctx, member.ID)
	if err != nil {
		s.appLogger.Error("Error checking existing member",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error al verificar existencia del miembro", err)
		return errors.DB(err, "error verificando miembro existente")
	}

	if existing == nil {
		s.appLogger.Error("Member not found", zap.Uint("id", member.ID))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Miembro no encontrado", nil)
		return errors.NotFound("member", nil)
	}

	// No permitir cambios en campos inmutables
	member.MembershipNumber = existing.MembershipNumber
	member.RegistrationDate = existing.RegistrationDate

	// Validar el miembro antes de actualizar
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error en la validación del miembro", err)

		// Conservar el error de validación si ya es un AppError, sino convertirlo
		appErr, ok := errors.AsAppError(err)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", err.Error())
	}

	// Actualizar el miembro
	if err = s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to update member",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(member.ID),
			"Error al actualizar miembro en base de datos", err)
		return errors.DB(err, "error actualizando miembro")
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		numToStr(member.ID),
		existing, // datos anteriores
		member,   // datos nuevos
		"Updated member with numero_socio "+member.MembershipNumber)

	s.appLogger.Info("Member updated successfully",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))

	return nil
}

// DeactivateMember implementa la lógica de baja de un miembro
func (s *memberService) DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error {
	// Obtener el miembro
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		s.appLogger.Error("Error getting member",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error al obtener miembro para desactivación", err)
		return errors.DB(err, "error obteniendo miembro")
	}

	if member == nil {
		s.appLogger.Error("Member not found", zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Miembro no encontrado", nil)
		return errors.NotFound("member", nil)
	}

	// Guardar estado anterior para métricas
	previousStatus := member.State
	previousType := member.MembershipType

	// Verificar que no esté ya inactivo
	if member.State == models.EstadoInactivo {
		s.appLogger.Warn("Member already inactive", zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Intento de desactivar miembro ya inactivo", nil)
		return errors.New(errors.ErrInvalidOperation, "el miembro ya está dado de baja")
	}

	// Guardar estado anterior para el log de auditoría
	previousState := *member

	// Establecer fecha de baja
	if fechaBaja == nil {
		now := time.Now()
		fechaBaja = &now
	}
	member.LeavingDate = fechaBaja
	member.State = models.EstadoInactivo

	// Validar y guardar cambios
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error en la validación del miembro", err)

		appErr, ok := errors.AsAppError(err)
		if ok {
			return appErr
		}
		return errors.Validation("Error validando miembro", "", err.Error())
	}

	if err := s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to deactivate member",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			numToStr(id),
			"Error al desactivar miembro en base de datos", err)
		return errors.DB(err, "error desactivando miembro")
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		numToStr(id),
		&previousState,
		member,
		"Deactivated member with numero_socio "+member.MembershipNumber)

	s.appLogger.Info("Member deactivated successfully",
		zap.String("numero_socio", member.MembershipNumber),
		zap.Uint("member_id", member.ID))

	// Actualizar métricas
	metrics.MembersByStatus.WithLabelValues(
		previousStatus,
		previousType,
	).Dec()

	metrics.MembersByStatus.WithLabelValues(
		member.State,
		member.MembershipType,
	).Inc()

	return nil
}

// ListMembers obtiene una lista de miembros según los criterios especificados
func (s *memberService) ListMembers(ctx context.Context, filters input.MemberFilters) ([]*models.Member, error) {
	// Convertir filtros de input a output
	repoFilters := output.MemberFilters{
		Estado:        filters.State,
		TipoMembresia: filters.MembershipType,
		SearchTerm:    filters.SearchTerm,
		Page:          filters.Page,
		PageSize:      filters.PageSize,
		OrderBy:       filters.OrderBy,
	}

	members, err := s.repository.List(ctx, repoFilters)
	if err != nil {
		s.appLogger.Error("Error listing members", zap.Error(err))
		return nil, errors.DB(err, "error al listar miembros")
	}

	// Convertir []models.Member a []*models.Member
	result := make([]*models.Member, len(members))
	for i := range members {
		result[i] = &members[i]
	}

	return result, nil
}

// numToStr es una función auxiliar para convertir un número a string para los logs
func numToStr(num uint) string {
	return strconv.FormatUint(uint64(num), 10)
}
