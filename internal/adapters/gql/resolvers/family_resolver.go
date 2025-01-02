// family_resolver.go
package resolvers

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/domain/models"
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

func (r *familyResolver) mapUpdateInputToFamily(id uint, input *model.UpdateFamilyInput, existing *models.Family) *models.Family {
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
	if err := family.Validate(); err != nil {
		return nil, err
	}

	if family.ID == 0 {
		err := r.familyService.Create(ctx, family)
		if err != nil {
			return nil, err
		}
	} else {
		err := r.familyService.Update(ctx, family)
		if err != nil {
			return nil, err
		}
	}

	return family, nil
}

func (r *familyResolver) validateCreateFamilyInput(input *model.CreateFamilyInput) error {
	if input.NumeroSocio == "" {
		return NewValidationError("numero_socio is required")
	}
	return nil
}

func (r *familyResolver) validateUpdateFamilyInput(input *model.UpdateFamilyInput) error {
	if input.FamiliaID == "" {
		return NewValidationError("familia_id is required")
	}
	return nil
}
