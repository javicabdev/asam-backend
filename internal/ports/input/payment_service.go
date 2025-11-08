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
	RegisterPayment(ctx context.Context, payment *models.Payment) error
	CancelPayment(ctx context.Context, paymentID uint, reason string) error
	ConfirmPayment(ctx context.Context, paymentID uint, paymentMethod string, paymentDate *time.Time, notes *string) (*models.Payment, error)
	GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error)
	GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error)
	GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error)

	// GenerateAnnualFees genera cuotas anuales para todos los socios activos
	GenerateAnnualFees(ctx context.Context, req *GenerateAnnualFeesRequest) (*GenerateAnnualFeesResponse, error)
	GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error)
	UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error

	GetMemberStatement(ctx context.Context, memberID uint) (*AccountStatement, error)
	GetFamilyStatement(ctx context.Context, familyID uint) (*AccountStatement, error)
	GetDefaulters(ctx context.Context) ([]AccountStatement, error)

	SendPaymentReminder(ctx context.Context, memberID uint) error
	SendPaymentConfirmation(ctx context.Context, paymentID uint) error
	SendDefaulterNotification(ctx context.Context, memberID uint, days int) error

	// ListPayments retrieves a paginated list of payments with optional filters
	ListPayments(ctx context.Context, filters PaymentFilters) ([]*models.Payment, int, error)
}

// PaymentFilters defines search criteria for payments
type PaymentFilters struct {
	Status        *models.PaymentStatus
	PaymentMethod *string
	StartDate     *time.Time
	EndDate       *time.Time
	MinAmount     *float64
	MaxAmount     *float64
	MemberID      *uint
	Page          int
	PageSize      int
	OrderBy       string
}

// FeeCalculator defines the interface for the calculation of quotas
type FeeCalculator interface {
	CalculateBaseFee(year, month int) float64
	CalculateFamilyFee(year, month int) float64
	CalculateLateFee(daysLate int) float64
}

// GenerateAnnualFeesRequest contiene los datos para generar cuotas anuales
type GenerateAnnualFeesRequest struct {
	Year           int     // Año para el cual generar cuotas
	BaseFeeAmount  float64 // Monto base (para socios individuales)
	FamilyFeeExtra float64 // Monto adicional para socios familiares
}

// PaymentGenerationDetail contiene información sobre un pago individual generado
type PaymentGenerationDetail struct {
	MemberID       uint
	MemberNumber   string
	MemberName     string
	Amount         float64
	WasCreated     bool   // true si se creó, false si ya existía
	Error          string // mensaje de error si falló
}

// GenerateAnnualFeesResponse contiene el resultado de generar cuotas anuales
type GenerateAnnualFeesResponse struct {
	Year                int
	MembershipFeeID     uint
	PaymentsGenerated   int
	PaymentsExisting    int
	TotalMembers        int
	TotalExpectedAmount float64
	Details             []PaymentGenerationDetail
}
