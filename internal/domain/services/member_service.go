package services

import (
	"context"
	"fmt"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"time"
)

type memberService struct {
	repository output.MemberRepository
}

// NewMemberService crea una nueva instancia del servicio de miembros
func NewMemberService(repository output.MemberRepository) input.MemberService {
	return &memberService{
		repository: repository,
	}
}

// CreateMember implementa la lógica de creación de un nuevo miembro
func (s *memberService) CreateMember(ctx context.Context, member *models.Member) error {
	// Verificar si ya existe un miembro con el mismo número de socio
	existing, err := s.repository.GetByNumeroSocio(ctx, member.NumeroSocio)
	if err != nil {
		return fmt.Errorf("error checking existing member: %w", err)
	}
	if existing != nil {
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
		return fmt.Errorf("error validating member: %w", err)
	}

	// Crear el miembro en la base de datos
	return s.repository.Create(ctx, member)
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
		return fmt.Errorf("error checking existing member: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("no existe un miembro con el ID %d", member.ID)
	}

	// No permitir cambios en campos inmutables
	member.NumeroSocio = existing.NumeroSocio
	member.FechaAlta = existing.FechaAlta

	// Validar el miembro antes de actualizar
	if err := member.Validate(); err != nil {
		return fmt.Errorf("error validating member: %w", err)
	}

	return s.repository.Update(ctx, member)
}

// DeactivateMember implementa la lógica de baja de un miembro
func (s *memberService) DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error {
	// Obtener el miembro
	member, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting member: %w", err)
	}
	if member == nil {
		return fmt.Errorf("no existe un miembro con el ID %d", id)
	}

	// Verificar que no esté ya inactivo
	if member.Estado == models.EstadoInactivo {
		return fmt.Errorf("el miembro ya está dado de baja")
	}

	// Establecer fecha de baja
	if fechaBaja == nil {
		now := time.Now()
		fechaBaja = &now
	}
	member.FechaBaja = fechaBaja
	member.Estado = models.EstadoInactivo

	// Validar y guardar cambios
	if err := member.Validate(); err != nil {
		return fmt.Errorf("error validating member: %w", err)
	}

	return s.repository.Update(ctx, member)
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
