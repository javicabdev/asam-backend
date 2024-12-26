package validation

import (
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
		return &ValidationError{
			Field:   "numero_socio",
			Message: "número de socio es requerido",
		}
	}

	numeroSocio = strings.TrimSpace(strings.ToUpper(numeroSocio))
	socioRegex := regexp.MustCompile(`^[A-Z]\d{4}$`)

	if !socioRegex.MatchString(numeroSocio) {
		return &ValidationError{
			Field:   "numero_socio",
			Message: "formato de número de socio inválido (debe ser letra mayúscula seguida de 4 dígitos)",
		}
	}

	return nil
}

// ValidateConyuges valida que al menos uno de los cónyuges tenga datos
func (v *DefaultFamilyValidator) ValidateConyuges(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error {
	if (esposoNombre == "" && esposoApellidos == "") &&
		(esposaNombre == "" && esposaApellidos == "") {
		return &ValidationError{
			Field:   "conyuges",
			Message: "se requiere información de al menos un cónyuge",
		}
	}
	return nil
}

// ValidateDocumentIDs valida el formato de los documentos de identidad
func (v *DefaultFamilyValidator) ValidateDocumentIDs(esposoDoc, esposaDoc string) error {
	// Reutilizar el validador de miembros para los documentos
	memberValidator := NewMemberValidator()

	if esposoDoc != "" {
		if err := memberValidator.ValidateDocumentID(esposoDoc); err != nil {
			return &ValidationError{
				Field:   "esposo_documento",
				Message: "documento de identidad del esposo inválido",
			}
		}
	}

	if esposaDoc != "" {
		if err := memberValidator.ValidateDocumentID(esposaDoc); err != nil {
			return &ValidationError{
				Field:   "esposa_documento",
				Message: "documento de identidad de la esposa inválido",
			}
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
			return &ValidationError{
				Field:   "esposo_email",
				Message: "correo electrónico del esposo inválido",
			}
		}
	}

	if esposaEmail != "" {
		if err := memberValidator.ValidateContactInfo(esposaEmail, "", ""); err != nil {
			return &ValidationError{
				Field:   "esposa_email",
				Message: "correo electrónico de la esposa inválido",
			}
		}
	}

	return nil
}

// ValidateDates valida las fechas de nacimiento
func (v *DefaultFamilyValidator) ValidateDates(esposoFechaNac, esposaFechaNac *time.Time) error {
	now := time.Now()

	if esposoFechaNac != nil && esposoFechaNac.After(now) {
		return &ValidationError{
			Field:   "esposo_fecha_nacimiento",
			Message: "fecha de nacimiento del esposo no puede ser futura",
		}
	}

	if esposaFechaNac != nil && esposaFechaNac.After(now) {
		return &ValidationError{
			Field:   "esposa_fecha_nacimiento",
			Message: "fecha de nacimiento de la esposa no puede ser futura",
		}
	}

	return nil
}
