package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// DefaultMemberValidator implementación por defecto del validador de miembros
type DefaultMemberValidator struct{}

// NewMemberValidator crea una nueva instancia del validador de miembros
func NewMemberValidator() MemberValidator {
	return &DefaultMemberValidator{}
}

// ValidateDocumentID valida el formato del documento de identidad (DNI/NIE)
// DEPRECATED: usar ValidateDocumentByType en su lugar
func (v *DefaultMemberValidator) ValidateDocumentID(documentID string) error {
	return v.ValidateDocumentByType(documentID, "DNI_NIE")
}

// ValidateDocumentByType valida el documento según su tipo
// documentType puede ser: "DNI_NIE", "SENEGAL_PASSPORT", "OTHER"
func (v *DefaultMemberValidator) ValidateDocumentByType(documentID, documentType string) error {
	// Si no se proporciona documento, es válido (campo opcional)
	if documentID == "" {
		return nil
	}

	// Si no se proporciona tipo o es "OTHER", no validar formato
	if documentType == "" || documentType == "OTHER" {
		// No validamos, solo verificamos que no esté vacío si se proporcionó
		return nil
	}

	switch documentType {
	case "DNI_NIE":
		// Validar DNI/NIE español
		if !ValidarNIF(documentID) {
			return appErrors.NewValidationError(
				"invalid DNI/NIE format",
				map[string]string{"documento_identidad": "Formato inválido o letra de control incorrecta (DNI/NIE)"},
			)
		}

	case "SENEGAL_PASSPORT":
		// Validar pasaporte senegalés (MRZ 10 caracteres)
		if !ValidarPasaporteSenegal(documentID) {
			return appErrors.NewValidationError(
				"invalid Senegal passport format",
				map[string]string{"documento_identidad": "Formato inválido. Debe ser 10 caracteres de la MRZ (A-Z, 0-9, < más dígito de control)"},
			)
		}

	default:
		// Tipo desconocido, no validar
		return nil
	}

	return nil
}

// ProcessDocumentForStorage procesa el documento para almacenamiento según su tipo
// - DNI/NIE: normaliza (elimina espacios/guiones, mayúsculas)
// - SENEGAL_PASSPORT: extrae solo los 9 caracteres (sin dígito de control)
// - OTHER: normalización básica (trim, mayúsculas)
func ProcessDocumentForStorage(documentID, documentType string) string {
	if documentID == "" {
		return ""
	}

	switch documentType {
	case "DNI_NIE":
		// Normalizar DNI/NIE
		return NormalizarNIF(documentID)

	case "SENEGAL_PASSPORT":
		// Extraer solo los 9 caracteres del número (sin dígito de control)
		return ExtraerNumeroPasaporteSenegal(documentID)

	case "OTHER", "":
		// Normalización básica: trim y mayúsculas
		normalized := strings.TrimSpace(documentID)
		return strings.ToUpper(normalized)

	default:
		// Por defecto, normalización básica
		normalized := strings.TrimSpace(documentID)
		return strings.ToUpper(normalized)
	}
}

// ValidateContactInfo valida el email, teléfono y dirección
func (v *DefaultMemberValidator) ValidateContactInfo(email, phone, address string) error {
	// Validar email
	if email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(email) {
			return appErrors.NewValidationError(
				"invalid email format",
				map[string]string{"email": "Formato inválido"},
			)
		}
	}

	// Validar teléfono (formato español)
	if phone != "" {
		phoneRegex := regexp.MustCompile(`^(?:(?:\+|00)?34)?[6789]\d{8}$`)
		if !phoneRegex.MatchString(phone) {
			return appErrors.NewValidationError(
				"invalid phone number format",
				map[string]string{"phone": "Formato inválido"},
			)
		}
	}

	// Validar que la dirección no esté vacía si se proporciona
	if address != "" && strings.TrimSpace(address) == "" {
		return appErrors.NewValidationError(
			"address cannot be empty if provided",
			map[string]string{"address": "empty"},
		)
	}

	return nil
}

// ValidateMemberStatus valida el estado del miembro y sus transiciones
func (v *DefaultMemberValidator) ValidateMemberStatus(status string, currentStatus string) error {
	validStatuses := map[string]bool{
		"active":   true,
		"inactive": true,
	}

	if !validStatuses[status] {
		return appErrors.NewValidationError(
			"invalid status",
			map[string]string{"status": "Formato inválido"},
		)
	}

	// Validar transiciones de estado
	if currentStatus != "" {
		validTransitions := map[string]map[string]bool{
			"active": {
				"inactive": true,
			},
			"inactive": {
				"active": true,
			},
		}

		if !validTransitions[currentStatus][status] {
			return appErrors.NewValidationError(
				fmt.Sprintf("invalid status transition from %s to %s", currentStatus, status),
				map[string]string{"status": "Invalid transition"},
			)
		}
	}

	return nil
}

// ValidateDates valida las fechas de alta y baja
func (v *DefaultMemberValidator) ValidateDates(registrationDate, cancellationDate *time.Time) error {
	now := time.Now()

	// Validar fecha de alta
	if registrationDate != nil {
		if registrationDate.After(now) {
			return appErrors.NewValidationError(
				"registration date cannot be in the future",
				map[string]string{"registration_date": "Invalid registration date"},
			)
		}
	}

	// Validar fecha de baja
	if cancellationDate != nil {
		if registrationDate != nil && cancellationDate.Before(*registrationDate) {
			return appErrors.NewValidationError(
				"cancellation date cannot be before registration date",
				map[string]string{"cancellation_date": "Invalid cancellation date"},
			)
		}
		if cancellationDate.After(now) {
			return appErrors.NewValidationError(
				"cancellation date cannot be in the future",
				map[string]string{"cancellation_date": "Invalid cancellation date"},
			)
		}
	}

	return nil
}
