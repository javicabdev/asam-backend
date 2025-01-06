package services

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/javicabdev/asam-backend/pkg/metrics"
	"go.uber.org/zap"
	"time"
)

type memberService struct {
	repository  output.MemberRepository
	appLogger   logger.Logger
	auditLogger *audit.Audit
}

// NewMemberService crea una nueva instancia del servicio de miembros
func NewMemberService(repository output.MemberRepository, appLogger logger.Logger, auditLogger *audit.Audit) input.MemberService {
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
		zap.String("numero_socio", member.NumeroSocio),
		zap.String("tipo_membresia", member.TipoMembresia))

	// Verificar si ya existe un miembro con el mismo número de socio
	existing, err := s.repository.GetByNumeroSocio(ctx, member.NumeroSocio)
	if err != nil {
		s.appLogger.Error("Error checking existing member",
			zap.String("numero_socio", member.NumeroSocio),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.NumeroSocio,
			"Error al verificar miembro existente", err)
		return fmt.Errorf("error checking existing member: %w", err)
	}
	if existing != nil {
		s.appLogger.Warn("Attempted to create duplicate member",
			zap.String("numero_socio", member.NumeroSocio))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.NumeroSocio,
			"Intento de crear miembro duplicado", fmt.Errorf("miembro ya existe"))
		return fmt.Errorf("ya existe un miembro con el número de socio %s", member.NumeroSocio)
	}

	// Establecer valores por defecto
	if member.Estado == "" {
		member.Estado = models.EstadoActivo
	}
	if member.Provincia == "" {
		member.Provincia = "Barcelona"
	}
	if member.Pais == "" {
		member.Pais = "España"
	}
	if member.Nacionalidad == "" {
		member.Nacionalidad = "Senegal"
	}

	// Validar el miembro antes de crear
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.String("numero_socio", member.NumeroSocio),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.NumeroSocio,
			"Error en la validación del miembro", err)
		return fmt.Errorf("error validating member: %w", err)
	}

	// Crear el miembro en la base de datos
	if err := s.repository.Create(ctx, member); err != nil {
		s.appLogger.Error("Failed to create member",
			zap.String("numero_socio", member.NumeroSocio),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionCreate, audit.EntityMember, member.NumeroSocio,
			"Error al crear miembro en base de datos", err)
		return fmt.Errorf("error creating member: %w", err)
	}

	// Actualizar métricas de miembros
	metrics.MembersByStatus.WithLabelValues(
		member.Estado,
		member.TipoMembresia,
	).Inc()

	// Registrar la acción en el log de auditoría
	s.auditLogger.LogAction(ctx,
		audit.ActionCreate,
		audit.EntityMember,
		fmt.Sprintf("%d", member.ID),
		fmt.Sprintf("Created new member with numero_socio %s", member.NumeroSocio))

	s.appLogger.Info("Member created successfully",
		zap.String("numero_socio", member.NumeroSocio),
		zap.Uint("member_id", member.ID))

	return nil
}

// GetMemberByID obtiene un miembro por su ID
func (s *memberService) GetMemberByID(ctx context.Context, id uint) (*models.Member, error) {
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting member by ID: %w", err)
	}
	return member, nil
}

// GetMemberByNumeroSocio obtiene un miembro por su número de socio
func (s *memberService) GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	member, err := s.repository.GetByNumeroSocio(ctx, numeroSocio)
	if err != nil {
		return nil, fmt.Errorf("error getting member by numero socio: %w", err)
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
			fmt.Sprintf("%d", member.ID),
			"Error al verificar existencia del miembro", err)
		return fmt.Errorf("error checking existing member: %w", err)
	}
	if existing == nil {
		s.appLogger.Error("Member not found",
			zap.Uint("id", member.ID))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", member.ID),
			"Miembro no encontrado", fmt.Errorf("no existe un miembro con el ID %d", member.ID))
		return fmt.Errorf("no existe un miembro con el ID %d", member.ID)
	}

	// No permitir cambios en campos inmutables
	member.NumeroSocio = existing.NumeroSocio
	member.FechaAlta = existing.FechaAlta

	// Validar el miembro antes de actualizar
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", member.ID),
			"Error en la validación del miembro", err)
		return fmt.Errorf("error validating member: %w", err)
	}

	// Actualizar el miembro
	if err := s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to update member",
			zap.Uint("id", member.ID),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", member.ID),
			"Error al actualizar miembro en base de datos", err)
		return fmt.Errorf("error updating member: %w", err)
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		fmt.Sprintf("%d", member.ID),
		existing, // datos anteriores
		member,   // datos nuevos
		fmt.Sprintf("Updated member with numero_socio %s", member.NumeroSocio))

	s.appLogger.Info("Member updated successfully",
		zap.String("numero_socio", member.NumeroSocio),
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
			fmt.Sprintf("%d", id),
			"Error al obtener miembro para desactivación", err)
		return fmt.Errorf("error getting member: %w", err)
	}
	if member == nil {
		s.appLogger.Error("Member not found",
			zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", id),
			"Miembro no encontrado", fmt.Errorf("no existe un miembro con el ID %d", id))
		return fmt.Errorf("no existe un miembro con el ID %d", id)
	}

	// Guardar estado anterior para métricas
	previousStatus := member.Estado
	previousType := member.TipoMembresia

	// Verificar que no esté ya inactivo
	if member.Estado == models.EstadoInactivo {
		s.appLogger.Warn("Member already inactive",
			zap.Uint("id", id))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", id),
			"Intento de desactivar miembro ya inactivo", fmt.Errorf("el miembro ya está dado de baja"))
		return fmt.Errorf("el miembro ya está dado de baja")
	}

	// Guardar estado anterior para el log de auditoría
	previousState := *member

	// Establecer fecha de baja
	if fechaBaja == nil {
		now := time.Now()
		fechaBaja = &now
	}
	member.FechaBaja = fechaBaja
	member.Estado = models.EstadoInactivo

	// Validar y guardar cambios
	if err := member.Validate(); err != nil {
		s.appLogger.Error("Member validation failed",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", id),
			"Error en la validación del miembro", err)
		return fmt.Errorf("error validating member: %w", err)
	}

	if err := s.repository.Update(ctx, member); err != nil {
		s.appLogger.Error("Failed to deactivate member",
			zap.Uint("id", id),
			zap.Error(err))
		s.auditLogger.LogError(ctx, audit.ActionUpdate, audit.EntityMember,
			fmt.Sprintf("%d", id),
			"Error al desactivar miembro en base de datos", err)
		return fmt.Errorf("error deactivating member: %w", err)
	}

	// Log de auditoría con los cambios
	s.auditLogger.LogChange(ctx, audit.ActionUpdate, audit.EntityMember,
		fmt.Sprintf("%d", id),
		&previousState,
		member,
		fmt.Sprintf("Deactivated member with numero_socio %s", member.NumeroSocio))

	s.appLogger.Info("Member deactivated successfully",
		zap.String("numero_socio", member.NumeroSocio),
		zap.Uint("member_id", member.ID))

	// Actualizar métricas
	metrics.MembersByStatus.WithLabelValues(
		previousStatus,
		previousType,
	).Dec()

	metrics.MembersByStatus.WithLabelValues(
		member.Estado,
		member.TipoMembresia,
	).Inc()

	return nil
}

// ListMembers obtiene una lista de miembros según los criterios especificados
func (s *memberService) ListMembers(ctx context.Context, filters input.MemberFilters) ([]models.Member, error) {
	// Convertir filtros de input a output
	repoFilters := output.MemberFilters{
		Estado:        filters.Estado,
		TipoMembresia: filters.TipoMembresia,
		SearchTerm:    filters.SearchTerm,
		Page:          filters.Page,
		PageSize:      filters.PageSize,
	}

	members, err := s.repository.List(ctx, repoFilters)
	if err != nil {
		return nil, fmt.Errorf("error listing members: %w", err)
	}

	return members, nil
}
