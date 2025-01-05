package resolvers

import (
	"context"
	stdErrors "errors"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

func (r *familyResolver) mapCreateInputToFamily(input *model.CreateFamilyInput) *models.Family {
	var miembroOrigenID *uint
	if input.MiembroOrigenID != nil {
		id := parseID(*input.MiembroOrigenID)
		miembroOrigenID = &id
	}

	return &models.Family{
		NumeroSocio:              input.NumeroSocio,
		MiembroOrigenID:          miembroOrigenID,
		EsposoNombre:             input.EsposoNombre,
		EsposoApellidos:          input.EsposoApellidos,
		EsposaNombre:             input.EsposaNombre,
		EsposaApellidos:          input.EsposaApellidos,
		EsposoFechaNacimiento:    input.EsposoFechaNacimiento,
		EsposoDocumentoIdentidad: *input.EsposoDocumentoIdentidad,
		EsposoCorreoElectronico:  *input.EsposoCorreoElectronico,
		EsposaFechaNacimiento:    input.EsposaFechaNacimiento,
		EsposaDocumentoIdentidad: *input.EsposaDocumentoIdentidad,
		EsposaCorreoElectronico:  *input.EsposaCorreoElectronico,
	}
}

func (r *familyResolver) mapUpdateInputToFamily(input *model.UpdateFamilyInput, existing *models.Family) *models.Family {
	family := *existing

	if input.EsposoNombre != nil {
		family.EsposoNombre = *input.EsposoNombre
	}
	if input.EsposoApellidos != nil {
		family.EsposoApellidos = *input.EsposoApellidos
	}
	if input.EsposaNombre != nil {
		family.EsposaNombre = *input.EsposaNombre
	}
	if input.EsposaApellidos != nil {
		family.EsposaApellidos = *input.EsposaApellidos
	}
	if input.EsposoDocumentoIdentidad != nil {
		family.EsposoDocumentoIdentidad = *input.EsposoDocumentoIdentidad
	}
	if input.EsposoCorreoElectronico != nil {
		family.EsposoCorreoElectronico = *input.EsposoCorreoElectronico
	}
	if input.EsposaDocumentoIdentidad != nil {
		family.EsposaDocumentoIdentidad = *input.EsposaDocumentoIdentidad
	}
	if input.EsposaCorreoElectronico != nil {
		family.EsposaCorreoElectronico = *input.EsposaCorreoElectronico
	}

	return &family
}

func (r *familyResolver) mapFamiliarInputToModel(input *model.FamiliarInput) *models.Familiar {
	var dni, email string
	if input.DniNie != nil {
		dni = *input.DniNie
	}
	if input.CorreoElectronico != nil {
		email = *input.CorreoElectronico
	}

	return &models.Familiar{
		Nombre:            input.Nombre,
		Apellidos:         input.Apellidos,
		FechaNacimiento:   input.FechaNacimiento,
		DNINIE:            dni,
		CorreoElectronico: email,
		Parentesco:        input.Parentesco,
	}
}

func (r *familyResolver) handleFamilyMutation(ctx context.Context, family *models.Family) (*models.Family, error) {
	// Validar la familia
	if err := family.Validate(); err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errors.NewValidationError(err.Error(), nil)
	}

	// Si hay miembro origen, verificar que existe
	if family.MiembroOrigenID != nil {
		member, err := r.memberService.GetMemberByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return nil, errors.NewBusinessError(
				errors.ErrDatabaseError,
				"Failed to verify origin member",
			)
		}
		if member == nil {
			return nil, errors.NewNotFoundError("origin member")
		}
	}

	// Crear o actualizar
	var err error
	if family.ID == 0 {
		err = r.familyService.Create(ctx, family)
	} else {
		err = r.familyService.Update(ctx, family)
	}

	if err != nil {
		return nil, errors.NewBusinessError(
			errors.ErrInternalError,
			"Failed to process family operation",
		)
	}

	return family, nil
}

func (r *familyResolver) handleFamiliarMutation(ctx context.Context, familyID uint, familiar *models.Familiar) (*models.Family, error) {
	// Verificar que la familia existe
	family, err := r.familyService.GetByID(ctx, familyID)
	if err != nil {
		return nil, errors.NewBusinessError(
			errors.ErrDatabaseError,
			"Failed to fetch family",
		)
	}
	if family == nil {
		return nil, errors.NewNotFoundError("family")
	}

	// Validar datos del familiar
	if err := familiar.Validate(); err != nil {
		return nil, errors.NewValidationError("Invalid familiar data", map[string]string{
			"details": err.Error(),
		})
	}

	// Añadir el familiar
	err = r.familyService.AddFamiliar(ctx, familyID, familiar)
	if err != nil {
		return nil, errors.NewBusinessError(
			errors.ErrInternalError,
			"Failed to add familiar",
		)
	}

	// Recargar la familia con los familiares actualizados
	return r.familyService.GetByID(ctx, familyID)
}

func (r *familyResolver) validateCreateFamilyInput(input *model.CreateFamilyInput) error {
	fields := make(map[string]string)

	if input.NumeroSocio == "" {
		fields["numero_socio"] = "Family number is required"
	}

	if input.EsposoNombre == "" {
		fields["esposo_nombre"] = "Husband's name is required"
	}

	if input.EsposoApellidos == "" {
		fields["esposo_apellidos"] = "Husband's last name is required"
	}

	if input.EsposaNombre == "" {
		fields["esposa_nombre"] = "Wife's name is required"
	}

	if input.EsposaApellidos == "" {
		fields["esposa_apellidos"] = "Wife's last name is required"
	}

	if input.EsposoDocumentoIdentidad == nil {
		fields["esposo_documento_identidad"] = "Husband's ID document is required"
	}

	if input.EsposaDocumentoIdentidad == nil {
		fields["esposa_documento_identidad"] = "Wife's ID document is required"
	}

	if len(fields) > 0 {
		return errors.NewValidationError("Invalid family input", fields)
	}

	return nil
}

func (r *familyResolver) validateUpdateFamilyInput(input *model.UpdateFamilyInput) error {
	if input.FamiliaID == "" {
		return errors.NewValidationError("Invalid input data", map[string]string{
			"familia_id": "Family ID is required",
		})
	}

	// Al menos un campo debe ser proporcionado para actualizar
	hasUpdates := input.EsposoNombre != nil ||
		input.EsposoApellidos != nil ||
		input.EsposaNombre != nil ||
		input.EsposaApellidos != nil ||
		input.EsposoDocumentoIdentidad != nil ||
		input.EsposoCorreoElectronico != nil ||
		input.EsposaDocumentoIdentidad != nil ||
		input.EsposaCorreoElectronico != nil

	if !hasUpdates {
		return errors.NewValidationError("Invalid input data", map[string]string{
			"update": "At least one field must be provided for update",
		})
	}

	return nil
}
