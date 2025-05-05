package input

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// AccountStatement represents the statement of account of a member or family.
type AccountStatement struct {
	TotalPaid       float64
	PendingPayments []models.MembershipFee
	PaymentHistory  []models.Payment
	LastPaymentDate *time.Time
	NextPaymentDate *time.Time
	IsDefaulter     bool
	DefaultDays     int
}

// PaymentService defines the transactions available for payment processing
type PaymentService interface {
	// Registration and payment management
	RegisterPayment(ctx context.Context, payment *models.Payment) error
	CancelPayment(ctx context.Context, paymentID uint, reason string) error
	GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error)
	GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error)
	GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error)

	// Quota management
	GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error
	GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error)
	UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error

	// Financial enquiries
	GetMemberStatement(ctx context.Context, memberID uint) (*AccountStatement, error)
	GetFamilyStatement(ctx context.Context, familyID uint) (*AccountStatement, error)
	GetDefaulters(ctx context.Context) ([]AccountStatement, error)

	// Notification system
	SendPaymentReminder(ctx context.Context, memberID uint) error
	SendPaymentConfirmation(ctx context.Context, paymentID uint) error
	SendDefaulterNotification(ctx context.Context, memberID uint, days int) error
}

// FeeCalculator defines the interface for the calculation of quotas
type FeeCalculator interface {
	CalculateBaseFee(year, month int) float64
	CalculateFamilyFee(year, month int) float64
	CalculateLateFee(daysLate int) float64
}
