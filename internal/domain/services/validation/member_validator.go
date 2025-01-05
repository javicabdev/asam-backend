package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type DefaultMemberValidator struct{}

func NewMemberValidator() MemberValidator {
	return &DefaultMemberValidator{}
}

// ValidateDocumentID valida el formato del documento de identidad (DNI/NIE)
func (v *DefaultMemberValidator) ValidateDocumentID(documentID string) error {
	if documentID == "" {
		return appErrors.NewValidationError(
			"document ID is required",
			map[string]string{"document_id": "documento requerido"},
		)
	}

	// Limpiamos el documento de espacios
	documentID = strings.TrimSpace(strings.ToUpper(documentID))

	// Regex para DNI: 8 dígitos seguidos de una letra
	dniRegex := regexp.MustCompile(`^[0-9]{8}[A-Z]$`)
	// Regex para NIE: X, Y o Z seguido de 7 dígitos y una letra
	nieRegex := regexp.MustCompile(`^[XYZ][0-9]{7}[A-Z]$`)

	if !dniRegex.MatchString(documentID) && !nieRegex.MatchString(documentID) {
		return appErrors.NewValidationError(
			"invalid document ID format",
			map[string]string{"document_id": "Formato inválido (DNI/NIE)"},
		)
	}

	// Aquí podríamos añadir la validación del dígito de control
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
