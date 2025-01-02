// member_resolver.go
package resolvers

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"time"
)

func (r *memberResolver) handleMemberMutation(ctx context.Context, member *models.Member) (*models.Member, error) {
	if err := member.Validate(); err != nil {
		return nil, NewValidationError(err.Error())
	}

	if member.ID == 0 {
		err := r.memberService.CreateMember(ctx, member)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.memberService.UpdateMember(ctx, member)
		if err != nil {
			return nil, err
		}
	}

	return member, nil
}

func (r *memberResolver) mapCreateInputToMember(input *model.CreateMemberInput) *models.Member {
	member := &models.Member{
		NumeroSocio:     input.NumeroSocio,
		TipoMembresia:   string(input.TipoMembresia),
		Nombre:          input.Nombre,
		Apellidos:       input.Apellidos,
		CalleNumeroPiso: input.CalleNumeroPiso,
		CodigoPostal:    input.CodigoPostal,
		Poblacion:       input.Poblacion,
		Estado:          models.EstadoActivo,
		FechaAlta:       time.Now(),
		FechaNacimiento: input.FechaNacimiento,
	}

	// Campos opcionales con valores por defecto
	if input.Provincia != nil {
		member.Provincia = *input.Provincia
	} else {
		member.Provincia = "Barcelona"
	}

	if input.Pais != nil {
		member.Pais = *input.Pais
	} else {
		member.Pais = "España"
	}

	if input.Nacionalidad != nil {
		member.Nacionalidad = *input.Nacionalidad
	} else {
		member.Nacionalidad = "Senegal"
	}

	// Campos opcionales
	if input.DocumentoIdentidad != nil {
		member.DocumentoIdentidad = input.DocumentoIdentidad
	}
	if input.CorreoElectronico != nil {
		member.CorreoElectronico = input.CorreoElectronico
	}
	if input.Profesion != nil {
		member.Profesion = input.Profesion
	}
	if input.Observaciones != nil {
		member.Observaciones = input.Observaciones
	}

	return member
}

func (r *memberResolver) mapUpdateInputToMember(id uint, input *model.UpdateMemberInput, existing *models.Member) *models.Member {
	member := *existing
	member.ID = id

	// Actualizar solo campos proporcionados
	if input.CalleNumeroPiso != nil {
		member.CalleNumeroPiso = *input.CalleNumeroPiso
	}
	if input.CodigoPostal != nil {
		member.CodigoPostal = *input.CodigoPostal
	}
	if input.Poblacion != nil {
		member.Poblacion = *input.Poblacion
	}
	if input.Provincia != nil {
		member.Provincia = *input.Provincia
	}
	if input.Pais != nil {
		member.Pais = *input.Pais
	}
	if input.DocumentoIdentidad != nil {
		member.DocumentoIdentidad = input.DocumentoIdentidad
	}
	if input.CorreoElectronico != nil {
		member.CorreoElectronico = input.CorreoElectronico
	}
	if input.Profesion != nil {
		member.Profesion = input.Profesion
	}
	if input.Observaciones != nil {
		member.Observaciones = input.Observaciones
	}

	return &member
}

func (r *memberResolver) handleMemberStatus(ctx context.Context, memberID uint, status model.MemberStatus) (*models.Member, error) {
	member, err := r.memberService.GetMemberByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, NewNotFoundError("member not found")
	}

	switch status {
	case model.MemberStatusActive:
		member.Estado = models.EstadoActivo
		member.FechaBaja = nil
	case model.MemberStatusInactive:
		member.Estado = models.EstadoInactivo
		now := time.Now()
		member.FechaBaja = &now
	}

	err = r.memberService.UpdateMember(ctx, member)
	if err != nil {
		return nil, err
	}

	return member, nil
}

// Funciones auxiliares para las validaciones específicas de miembro
func (r *memberResolver) validateCreateInput(input *model.CreateMemberInput) error {
	if input.NumeroSocio == "" {
		return NewValidationError("numero_socio is required")
	}
	if input.Nombre == "" {
		return NewValidationError("nombre is required")
	}
	if input.Apellidos == "" {
		return NewValidationError("apellidos is required")
	}
	if input.CalleNumeroPiso == "" {
		return NewValidationError("direccion is required")
	}
	if input.CodigoPostal == "" {
		return NewValidationError("codigo_postal is required")
	}
	if input.Poblacion == "" {
		return NewValidationError("poblacion is required")
	}
	return nil
}

func (r *memberResolver) validateUpdateInput(input *model.UpdateMemberInput) error {
	if input.MiembroID == "" {
		return NewValidationError("miembro_id is required")
	}
	return nil
}
