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

// ValidateDocumentID valida el documento de identidad
// La validación del formato específico se realiza en el frontend
// Esta función solo verifica que el documento no esté vacío cuando es requerido
func (v *DefaultMemberValidator) ValidateDocumentID(documentID string) error {
	// Permitir documentos vacíos ya que el campo es opcional
	// Si se proporciona, aceptamos cualquier formato (DNI, NIE, pasaporte, etc.)
	// La validación específica del formato se hace en el frontend
	return nil
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
