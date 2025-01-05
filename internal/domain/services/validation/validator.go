package validation

import "time"

type MemberValidator interface {
	ValidateDocumentID(documentID string) error
	ValidateContactInfo(email, phone, address string) error
	ValidateMemberStatus(status string, currentStatus string) error
	ValidateDates(registrationDate, cancellationDate *time.Time) error
}
