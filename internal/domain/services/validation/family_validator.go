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

// ValidateNumeroSocio valida el formato del número de socio
func (v *DefaultFamilyValidator) ValidateNumeroSocio(numeroSocio string) error {
	if numeroSocio == "" {
		return appErrors.NewValidationError(
			"número de socio es requerido",
			map[string]string{"numero_socio": "requerido"},
		)
	}

	numeroSocio = strings.TrimSpace(strings.ToUpper(numeroSocio))
	socioRegex := regexp.MustCompile(`^[A-Z]\d{4}$`)

	if !socioRegex.MatchString(numeroSocio) {
		return appErrors.NewValidationError(
			"formato de número de socio inválido (debe ser letra mayúscula seguida de 4 dígitos)",
			map[string]string{"numero_socio": "formato inválido"},
		)
	}

	return nil
}

// ValidateConyuges valida que al menos uno de los cónyuges tenga datos
func (v *DefaultFamilyValidator) ValidateConyuges(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error {
	if (esposoNombre == "" && esposoApellidos == "") &&
		(esposaNombre == "" && esposaApellidos == "") {
		return appErrors.NewValidationError(
			"se requiere información de al menos un cónyuge",
			map[string]string{"conyuges": "requerido"},
		)
	}

	return nil
}

// ValidateDocumentIDs valida el formato de los documentos de identidad
func (v *DefaultFamilyValidator) ValidateDocumentIDs(esposoDoc, esposaDoc string) error {
	// Reutilizar el validador de miembros para los documentos
	memberValidator := NewMemberValidator()

	if esposoDoc != "" {
		if err := memberValidator.ValidateDocumentID(esposoDoc); err != nil {
			return appErrors.NewValidationError(
				"documento de identidad del esposo inválido",
				map[string]string{"esposo_documento": "formato inválido"},
			)
		}
	}

	if esposaDoc != "" {
		if err := memberValidator.ValidateDocumentID(esposaDoc); err != nil {
			return appErrors.NewValidationError(
				"documento de identidad de la esposa inválido",
				map[string]string{"esposa_documento": "formato inválido"},
			)
		}
	}

	return nil
}

// ValidateContactInfo valida el formato de los correos electrónicos
func (v *DefaultFamilyValidator) ValidateContactInfo(esposoEmail, esposaEmail string) error {
	// Reutilizar el validador de miembros para los correos
	memberValidator := NewMemberValidator()

	if esposoEmail != "" {
		if err := memberValidator.ValidateContactInfo(esposoEmail, "", ""); err != nil {
			return appErrors.NewValidationError(
				"correo electrónico del esposo inválido",
				map[string]string{"esposo_email": "formato inválido"},
			)
		}
	}

	if esposaEmail != "" {
		if err := memberValidator.ValidateContactInfo(esposaEmail, "", ""); err != nil {
			return appErrors.NewValidationError(
				"correo electrónico de la esposa inválido",
				map[string]string{"esposa_email": "formato inválido"},
			)
		}
	}

	return nil
}

// ValidateDates valida las fechas de nacimiento
func (v *DefaultFamilyValidator) ValidateDates(esposoFechaNac, esposaFechaNac *time.Time) error {
	now := time.Now()

	if esposoFechaNac != nil && esposoFechaNac.After(now) {
		return appErrors.NewValidationError(
			"fecha de nacimiento del esposo no puede ser futura",
			map[string]string{"esposo_fecha_nacimiento": "formato inválido"},
		)
	}

	if esposaFechaNac != nil && esposaFechaNac.After(now) {
		return appErrors.NewValidationError(
			"fecha de nacimiento de la esposa no puede ser futura",
			map[string]string{"esposa_fecha_nacimiento": "formato inválido"},
		)
	}

	return nil
}
