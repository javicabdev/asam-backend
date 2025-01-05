package validation

import (
	"github.com/javicabdev/asam-backend/pkg/errors"
	"regexp"
	"time"
)

var (
	dniRegex   = regexp.MustCompile(`^[0-9]{8}[A-Z]$`)
	nieRegex   = regexp.MustCompile(`^[XYZ][0-9]{7}[A-Z]$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^(?:(?:\+|00)?34)?[6789]\d{8}$`)
)

type Validator struct{}

// ValidateDocumentID valida DNI/NIE
func (v *Validator) ValidateDocumentID(id string) error {
	if !dniRegex.MatchString(id) && !nieRegex.MatchString(id) {
		return errors.NewValidationError("Invalid document ID format", map[string]string{
			"document_id": "Must be a valid DNI or NIE",
		})
	}
	return nil
}

// ValidateEmail valida formato de email
func (v *Validator) ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return errors.NewValidationError("Invalid email format", map[string]string{
			"email": "Must be a valid email address",
		})
	}
	return nil
}

// ValidatePhone valida formato de teléfono
func (v *Validator) ValidatePhone(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return errors.NewValidationError("Invalid phone format", map[string]string{
			"phone": "Must be a valid Spanish phone number",
		})
	}
	return nil
}

// ValidateDate valida que una fecha sea válida y no futura
func (v *Validator) ValidateDate(date time.Time, allowFuture bool) error {
	if date.IsZero() {
		return errors.NewValidationError("Invalid date", map[string]string{
			"date": "Date cannot be empty",
		})
	}

	if !allowFuture && date.After(time.Now()) {
		return errors.NewValidationError("Invalid date", map[string]string{
			"date": "Date cannot be in the future",
		})
	}
	return nil
}

// ValidateAmount valida un importe
func (v *Validator) ValidateAmount(amount float64) error {
	if amount == 0 {
		return errors.NewValidationError("Invalid amount", map[string]string{
			"amount": "Amount cannot be zero",
		})
	}
	return nil
}
