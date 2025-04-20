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
	// Validate member before continuing
	if err := member.Validate(); err != nil {
		// If it's already an AppError, return it as is
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		// If it's a generic error, convert it:
		return nil, errors.NewValidationError(err.Error(), nil)
	}

	// For creation, verify that the user has admin permissions
	if member.ID == 0 {
		// Get user from context
		userInterface := ctx.Value(constants.UserContextKey)
		if userInterface == nil {
			return nil, errors.NewBusinessError(errors.ErrUnauthorized, "User not authenticated")
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			return nil, errors.NewBusinessError(errors.ErrUnauthorized, "Invalid user")
		}

		// Verify that user is admin
		if user.Role != models.RoleAdmin {
			return nil, errors.NewBusinessError(errors.ErrForbidden, "Insufficient permissions")
		}

		// Since user is admin, proceed to create the member
		err := r.memberService.CreateMember(ctx, member)
		if err != nil {
			return nil, err // Service errors are already wrapped properly
		}
	} else {
		err := r.memberService.UpdateMember(ctx, member)
		if err != nil {
			return nil, err // Service errors are already wrapped properly
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
			"Invalid membership type",
			map[string]string{"tipoMembresia": "Must be INDIVIDUAL or FAMILY"},
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

	// Optional fields with default values
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

	// Optional fields
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

	// Update only provided fields
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
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Failed to fetch member")
	}
	if member == nil {
		return nil, errors.NotFound("Member", nil)
	}

	switch status {
	case model.MemberStatusActive:
		member.Estado = models.EstadoActivo
		member.FechaBaja = nil
	case model.MemberStatusInactive:
		member.Estado = models.EstadoInactivo
		now := time.Now()
		member.FechaBaja = &now
	default:
		return nil, errors.NewValidationError(
			"Invalid member status",
			map[string]string{"status": "Must be ACTIVE or INACTIVE"},
		)
	}

	if err = r.memberService.UpdateMember(ctx, member); err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "Failed to update member status")
	}

	return member, nil
}

// Helper functions for specific member validations
func (r *memberResolver) validateCreateInput(input *model.CreateMemberInput) error {
	fields := make(map[string]string)

	if input.NumeroSocio == "" {
		fields["numeroSocio"] = "Member number is required"
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
		fields["codigoPostal"] = "Postal code is required"
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
			map[string]string{"miembroId": "Member ID is required"},
		)
	}

	// Check if at least one field is provided for update
	hasUpdates := input.CalleNumeroPiso != nil ||
		input.CodigoPostal != nil ||
		input.Poblacion != nil ||
		input.Provincia != nil ||
		input.Pais != nil ||
		input.DocumentoIdentidad != nil ||
		input.CorreoElectronico != nil ||
		input.Profesion != nil ||
		input.Observaciones != nil

	if !hasUpdates {
		return errors.NewValidationError(
			"No fields to update",
			map[string]string{"update": "At least one field must be provided for update"},
		)
	}

	return nil
}
