package resolvers

import (
	"context"
	stdErrors "errors"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
)

// safeStringDeref safely dereferences a string pointer, returning empty string if nil
func safeStringDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *familyResolver) mapCreateInputToFamily(input *model.CreateFamilyInput) *models.Family {
	var miembroOrigenID *uint
	if input.MiembroOrigenID != nil {
		id, err := parseID(*input.MiembroOrigenID)
		if err != nil {
			return nil
		}
		miembroOrigenID = &id
	}

	return &models.Family{
		NumeroSocio:              input.NumeroSocio,
		MiembroOrigenID:          miembroOrigenID,
		EsposoNombre:             input.EsposoNombre,
		EsposoApellidos:          input.EsposoApellidos,
		EsposaNombre:             safeStringDeref(input.EsposaNombre),
		EsposaApellidos:          safeStringDeref(input.EsposaApellidos),
		EsposoFechaNacimiento:    input.EsposoFechaNacimiento,
		EsposoDocumentoIdentidad: safeStringDeref(input.EsposoDocumentoIdentidad),
		EsposoCorreoElectronico:  safeStringDeref(input.EsposoCorreoElectronico),
		EsposaFechaNacimiento:    input.EsposaFechaNacimiento,
		EsposaDocumentoIdentidad: safeStringDeref(input.EsposaDocumentoIdentidad),
		EsposaCorreoElectronico:  safeStringDeref(input.EsposaCorreoElectronico),
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

func (r *familyResolver) mapCreateInputToAtomicRequest(
	familyInput *model.CreateFamilyInput,
) *input.CreateFamilyAtomicRequest {
	family := r.mapCreateInputToFamily(familyInput)

	// Preparar datos del member si no se proporcionó miembro_origen_id
	var memberData *input.CreateMemberData
	createMember := familyInput.MiembroOrigenID == nil

	if createMember {
		memberData = &input.CreateMemberData{
			Address:  safeStringDeref(familyInput.Direccion),
			Postcode: safeStringDeref(familyInput.CodigoPostal),
			City:     safeStringDeref(familyInput.Poblacion),
			Province: safeStringDeref(familyInput.Provincia),
			Country:  safeStringDeref(familyInput.Pais),
		}
	}

	// Mapear familiares
	var familiares []*models.Familiar
	if familyInput.Familiares != nil {
		familiares = make([]*models.Familiar, len(familyInput.Familiares))
		for i, famInput := range familyInput.Familiares {
			familiares[i] = r.mapFamiliarInputToModel(famInput)
		}
	}

	return &input.CreateFamilyAtomicRequest{
		Family:                  family,
		CreateMemberIfNotExists: createMember,
		MemberData:              memberData,
		Familiares:              familiares,
	}
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

	// Esposa opcional, pero si hay nombre, debe haber apellidos y viceversa
	if input.EsposaNombre != nil && *input.EsposaNombre != "" &&
		(input.EsposaApellidos == nil || *input.EsposaApellidos == "") {
		fields["esposaApellidos"] = "Si proporciona nombre de esposa, los apellidos son obligatorios"
	}

	if input.EsposaApellidos != nil && *input.EsposaApellidos != "" &&
		(input.EsposaNombre == nil || *input.EsposaNombre == "") {
		fields["esposaNombre"] = "Si proporciona apellidos de esposa, el nombre es obligatorio"
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
