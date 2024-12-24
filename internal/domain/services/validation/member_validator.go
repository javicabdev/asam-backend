// internal/domain/services/validation/member_validator.go

package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type DefaultMemberValidator struct{}

func NewMemberValidator() MemberValidator {
	return &DefaultMemberValidator{}
}

// ValidateDocumentID valida el formato del documento de identidad (DNI/NIE)
func (v *DefaultMemberValidator) ValidateDocumentID(documentID string) error {
	if documentID == "" {
		return &ValidationError{
			Field:   "document_id",
			Message: "document ID is required",
		}
	}

	// Limpiamos el documento de espacios
	documentID = strings.TrimSpace(strings.ToUpper(documentID))

	// Regex para DNI: 8 dígitos seguidos de una letra
	dniRegex := regexp.MustCompile(`^[0-9]{8}[A-Z]$`)
	// Regex para NIE: X, Y o Z seguido de 7 dígitos y una letra
	nieRegex := regexp.MustCompile(`^[XYZ][0-9]{7}[A-Z]$`)

	if !dniRegex.MatchString(documentID) && !nieRegex.MatchString(documentID) {
		return &ValidationError{
			Field:   "document_id",
			Message: "invalid document ID format",
		}
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
			return &ValidationError{
				Field:   "email",
				Message: "invalid email format",
			}
		}
	}

	// Validar teléfono (formato español)
	if phone != "" {
		phoneRegex := regexp.MustCompile(`^(?:(?:\+|00)?34)?[6789]\d{8}$`)
		if !phoneRegex.MatchString(phone) {
			return &ValidationError{
				Field:   "phone",
				Message: "invalid phone number format",
			}
		}
	}

	// Validar que la dirección no esté vacía si se proporciona
	if address != "" && strings.TrimSpace(address) == "" {
		return &ValidationError{
			Field:   "address",
			Message: "address cannot be empty if provided",
		}
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
		return &ValidationError{
			Field:   "status",
			Message: "invalid status",
		}
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
			return &ValidationError{
				Field:   "status",
				Message: fmt.Sprintf("invalid status transition from %s to %s", currentStatus, status),
			}
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
			return &ValidationError{
				Field:   "registration_date",
				Message: "registration date cannot be in the future",
			}
		}
	}

	// Validar fecha de baja
	if cancellationDate != nil {
		if registrationDate != nil && cancellationDate.Before(*registrationDate) {
			return &ValidationError{
				Field:   "cancellation_date",
				Message: "cancellation date cannot be before registration date",
			}
		}
		if cancellationDate.After(now) {
			return &ValidationError{
				Field:   "cancellation_date",
				Message: "cancellation date cannot be in the future",
			}
		}
	}

	return nil
}
