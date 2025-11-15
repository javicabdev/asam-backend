package validation

import "time"

// MemberValidator define la interfaz para validar miembros
type MemberValidator interface {
	ValidateDocumentID(documentID string) error
	ValidateDocumentByType(documentID, documentType string) error
	ValidateContactInfo(email, phone, address string) error
	ValidateMemberStatus(status string, currentStatus string) error
	ValidateDates(registrationDate, cancellationDate *time.Time) error
}
