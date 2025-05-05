package resolvers

import (
	"context"
	stdErrors "errors"
	"fmt"
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

func (r *familyResolver) mapUpdateInputToFamily(input *model.UpdateFamilyInput,
	existing *models.Family) *models.Family {
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
	// Validate family
	if err := family.Validate(); err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errors.NewValidationError(err.Error(), nil)
	}

	// If there's an origin member, verify it exists
	if family.MiembroOrigenID != nil {
		member, err := r.memberService.GetMemberByID(ctx, *family.MiembroOrigenID)
		if err != nil {
			return nil, errors.Wrap(err, errors.ErrDatabaseError, "Error verifying origin member")
		}
		if member == nil {
			return nil, errors.NotFound("origin member", nil)
		}
	}

	// Create or update
	var err error
	if family.ID == 0 {
		err = r.familyService.Create(ctx, family)
	} else {
		err = r.familyService.Update(ctx, family)
	}

	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "Error processing family operation")
	}

	return family, nil
}

func (r *familyResolver) handleFamiliarMutation(ctx context.Context, familyID uint, familiar *models.Familiar) (*models.Family, error) {
	// Verify that the family exists
	family, err := r.familyService.GetByID(ctx, familyID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrDatabaseError, "Error fetching family")
	}
	if family == nil {
		return nil, errors.NotFound("familia con ID "+fmt.Sprintf("%d", familyID), nil)
	}

	// Validate familiar data
	if err := familiar.Validate(); err != nil {
		var appErr *errors.AppError
		if stdErrors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, errors.NewValidationError(err.Error(), nil)
	}

	// Add the familiar
	err = r.familyService.AddFamiliar(ctx, familyID, familiar)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrInternalError, "Error adding familiar")
	}

	// Reload the family with updated familiares
	return r.familyService.GetByID(ctx, familyID)
}

func (r *familyResolver) validateCreateFamilyInput(input *model.CreateFamilyInput) error {
	fields := make(map[string]string)

	if input.NumeroSocio == "" {
		fields["numeroSocio"] = "El número de socio es obligatorio"
	}

	if input.EsposoNombre == "" {
		fields["esposoNombre"] = "El nombre del esposo es obligatorio"
	}

	if input.EsposoApellidos == "" {
		fields["esposoApellidos"] = "Los apellidos del esposo son obligatorios"
	}

	if input.EsposaNombre == "" {
		fields["esposaNombre"] = "El nombre de la esposa es obligatorio"
	}

	if input.EsposaApellidos == "" {
		fields["esposaApellidos"] = "Los apellidos de la esposa son obligatorios"
	}

	if input.EsposoDocumentoIdentidad == nil {
		fields["esposoDocumentoIdentidad"] = "El documento de identidad del esposo es obligatorio"
	}

	if input.EsposaDocumentoIdentidad == nil {
		fields["esposaDocumentoIdentidad"] = "El documento de identidad de la esposa es obligatorio"
	}

	if len(fields) > 0 {
		return errors.NewValidationError("Datos de familia inválidos", fields)
	}

	return nil
}

func (r *familyResolver) validateUpdateFamilyInput(input *model.UpdateFamilyInput) error {
	if input.FamiliaID == "" {
		return errors.NewValidationError("Datos de entrada inválidos", map[string]string{
			"familiaId": "El ID de familia es obligatorio",
		})
	}

	// At least one field must be provided for update
	hasUpdates := input.EsposoNombre != nil ||
		input.EsposoApellidos != nil ||
		input.EsposaNombre != nil ||
		input.EsposaApellidos != nil ||
		input.EsposoDocumentoIdentidad != nil ||
		input.EsposoCorreoElectronico != nil ||
		input.EsposaDocumentoIdentidad != nil ||
		input.EsposaCorreoElectronico != nil

	if !hasUpdates {
		return errors.NewValidationError("Datos de entrada inválidos", map[string]string{
			"update": "Se debe proporcionar al menos un campo para actualizar",
		})
	}

	return nil
}
