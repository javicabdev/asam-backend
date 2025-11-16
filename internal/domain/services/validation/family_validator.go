// Package validation proporciona funcionalidades de validación para las entidades de dominio,
// implementando reglas de negocio específicas para verificar la integridad de los datos.
package validation

import (
	"regexp"
	"strings"
	"time"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// FamilyValidator define la interfaz para validar familias
type FamilyValidator interface {
	ValidateNumeroSocio(numeroSocio string) error
	ValidateConyuges(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error
	ValidateConyugesFlexible(esposoNombre, esposoApellidos, esposaNombre, esposaApellidos string) error
	ValidateDocumentIDs(esposoDoc, esposaDoc string) error
	ValidateDocumentIDsWithTypes(esposoDoc, esposoDocType, esposaDoc, esposaDocType string) error
	ValidateContactInfo(esposoEmail, esposaEmail string) error
	ValidateDates(esposoFechaNac, esposaFechaNac *time.Time) error
}

// DefaultFamilyValidator implementación por defecto del validador de familias
type DefaultFamilyValidator struct{}

// NewFamilyValidator crea una nueva instancia del validador de familias
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
	socioRegex := regexp.MustCompile(`^[A-Z]\d{5,}$`)

	if !socioRegex.MatchString(numeroSocio) {
		return appErrors.NewValidationError(
			"Formato de número de socio inválido",
			map[string]string{"numeroSocio": "El formato debe ser una letra mayúscula seguida de al menos 5 dígitos"},
		)
	}

	return nil
}

// ValidateConyuges valida que al menos uno de los cónyuges tenga datos.
func (v *DefaultFamilyValidator) ValidateConyuges(
	esposoNombre, esposoApellidos,
	esposaNombre, esposaApellidos string,
) error {
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

// ValidateConyugesFlexible valida los datos de los cónyuges de forma flexible:
// - Esposo: siempre obligatorio (nombre y apellidos)
// - Esposa: opcional, pero si se proporciona nombre debe incluir apellidos y viceversa
func (v *DefaultFamilyValidator) ValidateConyugesFlexible(
	esposoNombre, esposoApellidos,
	esposaNombre, esposaApellidos string,
) error {
	errDetails := make(map[string]string)

	// Esposo es obligatorio
	if esposoNombre == "" {
		errDetails["esposoNombre"] = "El nombre del esposo es obligatorio"
	}
	if esposoApellidos == "" {
		errDetails["esposoApellidos"] = "Los apellidos del esposo son obligatorios"
	}

	// Esposa es opcional, pero si se proporciona nombre, apellidos también es obligatorio y viceversa
	if esposaNombre != "" && esposaApellidos == "" {
		errDetails["esposaApellidos"] = "Si proporciona nombre de esposa, los apellidos son obligatorios"
	}
	if esposaApellidos != "" && esposaNombre == "" {
		errDetails["esposaNombre"] = "Si proporciona apellidos de esposa, el nombre es obligatorio"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError(
			"Datos de cónyuges inválidos",
			errDetails,
		)
	}

	return nil
}

// ValidateDocumentIDs valida el formato de los documentos de identidad.
// Deprecated: usar ValidateDocumentIDsWithTypes en su lugar
func (v *DefaultFamilyValidator) ValidateDocumentIDs(esposoDoc, esposaDoc string) error {
	return v.ValidateDocumentIDsWithTypes(esposoDoc, "", esposaDoc, "")
}

// ValidateDocumentIDsWithTypes valida el formato de los documentos de identidad según su tipo.
// Esposo: documento obligatorio si se proporciona
// Esposa: documento opcional, pero si se proporciona debe ser válido
func (v *DefaultFamilyValidator) ValidateDocumentIDsWithTypes(esposoDoc, esposoDocType, esposaDoc, esposaDocType string) error {
	errDetails := make(map[string]string)

	// Reutilizar el validador de miembros para los documentos
	memberValidator := NewMemberValidator()
	if memberValidator == nil {
		return appErrors.NewValidationError(
			"Error al inicializar el validador de miembros",
			nil,
		)
	}

	// Validar documento del esposo (opcional pero recomendado)
	// Si se proporciona, debe ser válido
	if esposoDoc != "" {
		if err := memberValidator.ValidateDocumentByType(esposoDoc, esposoDocType); err != nil {
			if valErr, ok := appErrors.AsAppError(err); ok {
				for _, val := range valErr.Fields {
					errDetails["esposoDocumentoIdentidad"] = val
					break
				}
			} else {
				errDetails["esposoDocumentoIdentidad"] = "Documento de identidad del esposo inválido"
			}
		}
	}

	// Validar documento de la esposa (opcional)
	// Solo validar si se proporciona
	if esposaDoc != "" {
		if err := memberValidator.ValidateDocumentByType(esposaDoc, esposaDocType); err != nil {
			if valErr, ok := appErrors.AsAppError(err); ok {
				for _, val := range valErr.Fields {
					errDetails["esposaDocumentoIdentidad"] = val
					break
				}
			} else {
				errDetails["esposaDocumentoIdentidad"] = "Documento de identidad de la esposa inválido"
			}
		}
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
