package input

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// PaymentError representa errores específicos del dominio de pagos
type PaymentError struct {
	Code    string
	Message string
}

func (e *PaymentError) Error() string {
	return e.Message
}

// AccountStatement representa el estado de cuenta de un miembro o familia
type AccountStatement struct {
	TotalPaid       float64
	PendingPayments []models.MembershipFee
	PaymentHistory  []models.Payment
	LastPaymentDate *time.Time
	NextPaymentDate *time.Time
	IsDefaulter     bool
	DefaultDays     int
}

// PaymentService define las operaciones disponibles para la gestión de pagos
type PaymentService interface {
	// Registro y gestión de pagos
	RegisterPayment(ctx context.Context, payment *models.Payment) error
	CancelPayment(ctx context.Context, paymentID uint, reason string) error
	GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error)
	GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error)
	GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error)

	// Gestión de cuotas
	GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error
	GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error)
	UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error

	// Consultas financieras
	GetMemberStatement(ctx context.Context, memberID uint) (*AccountStatement, error)
	GetFamilyStatement(ctx context.Context, familyID uint) (*AccountStatement, error)
	GetDefaulters(ctx context.Context) ([]AccountStatement, error)

	// Sistema de notificaciones
	SendPaymentReminder(ctx context.Context, memberID uint) error
	SendPaymentConfirmation(ctx context.Context, paymentID uint) error
	SendDefaulterNotification(ctx context.Context, memberID uint, days int) error
}

// FeeCalculator define la interfaz para el cálculo de cuotas
type FeeCalculator interface {
	CalculateBaseFee(year, month int) float64
	CalculateFamilyFee(year, month int) float64
	CalculateLateFee(daysLate int) float64
}
