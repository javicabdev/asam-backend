package resolvers

import (
	"context"
	stdErrors "errors"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"time"
)

func (r *memberResolver) handleMemberMutation(ctx context.Context, member *models.Member) (*models.Member, error) {
	// Valida el miembro antes de continuar
	if err := member.Validate(); err != nil {
		// Si ya es un AppError, devuélvelo tal cual
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		// Si fuera un error genérico, lo convertimos:
		return nil, errors.NewValidationError(err.Error(), nil)
	}

	// Si se trata de una creación, se verifica que el usuario tenga permisos de administrador.
	if member.ID == 0 {
		// Obtener el usuario desde el contexto
		userInterface := ctx.Value(constants.UserContextKey)
		if userInterface == nil {
			return nil, errors.NewBusinessError(errors.ErrUnauthorized, "usuario no autenticado")
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			return nil, errors.NewBusinessError(errors.ErrUnauthorized, "usuario inválido")
		}

		// Verificar que el usuario sea administrador
		if user.Role != models.RoleAdmin {
			return nil, stdErrors.New("insufficient permissions")
		}

		// Como el usuario es admin, se procede a crear el miembro
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

func (r *memberResolver) mapTipoMembresia(tipo model.MembershipType) (string, error) {
	switch tipo {
	case model.MembershipTypeIndividual:
		return models.TipoMembresiaPIndividual, nil
	case model.MembershipTypeFamily:
		return models.TipoMembresiaPFamiliar, nil
	default:
		return "", errors.NewValidationError(
			"tipo de membresía no válido",
			map[string]string{"tipo_membresia": "debe ser INDIVIDUAL o FAMILY"},
		)
	}
}

func (r *memberResolver) mapCreateInputToMember(input *model.CreateMemberInput) (*models.Member, error) {
	tipoMembresia, err := r.mapTipoMembresia(input.TipoMembresia)
	if err != nil {
		return nil, err
	}

	member := &models.Member{
		NumeroSocio:     input.NumeroSocio,
		TipoMembresia:   tipoMembresia,
		Nombre:          input.Nombre,
		Apellidos:       input.Apellidos,
		Direccion:       input.Direccion,
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

	return member, nil
}

func (r *memberResolver) mapUpdateInputToMember(id uint, input *model.UpdateMemberInput, existing *models.Member) *models.Member {
	member := *existing
	member.ID = id

	// Actualizar solo campos proporcionados
	if input.CalleNumeroPiso != nil {
		member.Direccion = *input.CalleNumeroPiso
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
	if input.Direccion == "" {
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
