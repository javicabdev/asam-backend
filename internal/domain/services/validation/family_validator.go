package validation

import (
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"regexp"
	"strings"
	"time"
)

type FamilyValidator interface {
	ValidateNumeroSocio(numeroSocio string) error
	ValidateConyuges(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error
	ValidateDocumentIDs(esposoDoc, esposaDoc string) error
	ValidateContactInfo(esposoEmail, esposaEmail string) error
	ValidateDates(esposoFechaNac, esposaFechaNac *time.Time) error
}

type DefaultFamilyValidator struct{}

func NewFamilyValidator() FamilyValidator {
	return &DefaultFamilyValidator{}
}

// ValidateNumeroSocio valida el formato del número de socio.
func (v *DefaultFamilyValidator) ValidateNumeroSocio(numeroSocio string) error {
	if numeroSocio == "" {
		return appErrors.NewValidationError(
			"El número de socio es obligatorio",
			map[string]string{"numeroSocio": "El número de socio es obligatorio"},
		)
	}

	numeroSocio = strings.TrimSpace(strings.ToUpper(numeroSocio))
	socioRegex := regexp.MustCompile(`^[A-Z]\d{4}$`)

	if !socioRegex.MatchString(numeroSocio) {
		return appErrors.NewValidationError(
			"Formato de número de socio inválido",
			map[string]string{"numeroSocio": "El formato debe ser una letra mayúscula seguida de 4 dígitos"},
		)
	}

	return nil
}

// ValidateConyuges valida que al menos uno de los cónyuges tenga datos.
func (v *DefaultFamilyValidator) ValidateConyuges(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error {
	errDetails := make(map[string]string)

	if esposoNombre == "" {
		errDetails["esposoNombre"] = "El nombre del esposo es obligatorio"
	}

	if esposoApellidos == "" {
		errDetails["esposoApellidos"] = "Los apellidos del esposo son obligatorios"
	}

	if esposaNombre == "" {
		errDetails["esposaNombre"] = "El nombre de la esposa es obligatorio"
	}

	if esposaApellidos == "" {
		errDetails["esposaApellidos"] = "Los apellidos de la esposa son obligatorios"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError(
			"Se requiere información de al menos un cónyuge",
			errDetails,
		)
	}
	return nil
}

// ValidateDocumentIDs valida el formato de los documentos de identidad.
func (v *DefaultFamilyValidator) ValidateDocumentIDs(esposoDoc, esposaDoc string) error {
	errDetails := make(map[string]string)

	// Reutilizar el validador de miembros para los documentos
	memberValidator := NewMemberValidator()
	if memberValidator == nil {
		return appErrors.NewValidationError(
			"Error al inicializar el validador de miembros",
			nil,
		)
	}

	if esposoDoc == "" {
		errDetails["esposoDocumentoIdentidad"] = "El documento de identidad del esposo es obligatorio"
	} else if err := memberValidator.ValidateDocumentID(esposoDoc); err != nil {
		errDetails["esposoDocumentoIdentidad"] = "Documento de identidad del esposo inválido"
	}

	if esposaDoc == "" {
		errDetails["esposaDocumentoIdentidad"] = "El documento de identidad de la esposa es obligatorio"
	} else if err := memberValidator.ValidateDocumentID(esposaDoc); err != nil {
		errDetails["esposaDocumentoIdentidad"] = "Documento de identidad de la esposa inválido"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError(
			"Información de documentos inválida",
			errDetails,
		)
	}

	return nil
}

// ValidateContactInfo valida el formato de los correos electrónicos
func (v *DefaultFamilyValidator) ValidateContactInfo(esposoEmail, esposaEmail string) error {
	errDetails := make(map[string]string)

	// Reutilizar el validador de miembros para los correos
	memberValidator := NewMemberValidator()

	if esposoEmail != "" {
		if err := memberValidator.ValidateContactInfo(esposoEmail, "", ""); err != nil {
			errDetails["esposoCorreoElectronico"] = "Correo electrónico del esposo inválido"
		}
	}

	if esposaEmail != "" {
		if err := memberValidator.ValidateContactInfo(esposaEmail, "", ""); err != nil {
			errDetails["esposaCorreoElectronico"] = "Correo electrónico de la esposa inválido"
		}
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError(
			"Información de contacto inválida",
			errDetails,
		)
	}

	return nil
}

// ValidateDates valida las fechas de nacimiento
func (v *DefaultFamilyValidator) ValidateDates(esposoFechaNac, esposaFechaNac *time.Time) error {
	errDetails := make(map[string]string)
	now := time.Now()

	if esposoFechaNac != nil && esposoFechaNac.After(now) {
		errDetails["esposoFechaNacimiento"] = "Fecha de nacimiento del esposo no puede ser futura"
	}

	if esposaFechaNac != nil && esposaFechaNac.After(now) {
		errDetails["esposaFechaNacimiento"] = "Fecha de nacimiento de la esposa no puede ser futura"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError(
			"Fechas de nacimiento inválidas",
			errDetails,
		)
	}

	return nil
}
