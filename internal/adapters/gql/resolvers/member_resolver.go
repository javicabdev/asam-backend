package resolvers

import (
	"context"
	stdErrors "errors"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"time"
)

func (r *memberResolver) handleMemberMutation(ctx context.Context, member *models.Member) (*models.Member, error) {
	if err := member.Validate(); err != nil {
		// Si ya es un AppError, devuélvelo tal cual
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		// Si fuera un error genérico, entonces lo convertimos:
		return nil, errors.NewValidationError(err.Error(), nil)
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
		return nil, errors.NewBusinessError(
			errors.ErrDatabaseError,
			"Failed to fetch member",
		)
	}
	if member == nil {
		return nil, errors.NewNotFoundError("member")
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

	if err = r.memberService.UpdateMember(ctx, member); err != nil {
		return nil, errors.NewBusinessError(
			errors.ErrInternalError,
			"Failed to update member status",
		)
	}

	return member, nil
}

// Funciones auxiliares para las validaciones específicas de miembro
func (r *memberResolver) validateCreateInput(input *model.CreateMemberInput) error {
	fields := make(map[string]string)

	if input.NumeroSocio == "" {
		fields["numero_socio"] = "Member number is required"
	}
	if input.Nombre == "" {
		fields["nombre"] = "Name is required"
	}
	if input.Apellidos == "" {
		fields["apellidos"] = "Last name is required"
	}
	if input.CalleNumeroPiso == "" {
		fields["direccion"] = "Address is required"
	}
	if input.CodigoPostal == "" {
		fields["codigo_postal"] = "Postal code is required"
	}
	if input.Poblacion == "" {
		fields["poblacion"] = "City is required"
	}

	if len(fields) > 0 {
		return errors.NewValidationError("Invalid input data", fields)
	}

	return nil
}

func (r *memberResolver) validateUpdateInput(input *model.UpdateMemberInput) error {
	if input.MiembroID == "" {
		return errors.NewValidationError(
			"Invalid input data",
			map[string]string{"miembro_id": "Member ID is required"},
		)
	}
	return nil
}
