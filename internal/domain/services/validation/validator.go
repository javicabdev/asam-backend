package validation

import "time"

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type MemberValidator interface {
	ValidateDocumentID(documentID string) error
	ValidateContactInfo(email, phone, address string) error
	ValidateMemberStatus(status string, currentStatus string) error
	ValidateDates(registrationDate, cancellationDate *time.Time) error
}
