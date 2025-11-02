package output

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// PaymentRepository define las operaciones para persistir pagos
type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	Update(ctx context.Context, payment *models.Payment) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*models.Payment, error)
	FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error)
	FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error)
	HasInitialPayment(ctx context.Context, memberID *uint, familyID *uint) (bool, error)
	FindAll(ctx context.Context, filters *PaymentRepositoryFilters) ([]models.Payment, error)
	CountAll(ctx context.Context, filters *PaymentRepositoryFilters) (int64, error)

	// Transaction support
	CreateWithTx(ctx context.Context, tx Transaction, payment *models.Payment) error
}

// PaymentRepositoryFilters defines database-level filters for payment queries
type PaymentRepositoryFilters struct {
	Status        *models.PaymentStatus
	PaymentMethod *string
	StartDate     *time.Time
	EndDate       *time.Time
	MinAmount     *float64
	MaxAmount     *float64
	MemberID      *uint
	Offset        int
	Limit         int
	OrderBy       string
}

// MembershipFeeRepository define las operaciones para persistir cuotas de membresía
type MembershipFeeRepository interface {
	Create(ctx context.Context, fee *models.MembershipFee) error
	Update(ctx context.Context, fee *models.MembershipFee) error
	FindByID(ctx context.Context, id uint) (*models.MembershipFee, error)
	FindByYear(ctx context.Context, year int) (*models.MembershipFee, error)
	FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error)

	// Transaction support
	FindByYearWithTx(ctx context.Context, tx Transaction, year int) (*models.MembershipFee, error)
	CreateWithTx(ctx context.Context, tx Transaction, fee *models.MembershipFee) error
}
