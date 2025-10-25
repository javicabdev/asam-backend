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
}

// MembershipFeeRepository define las operaciones para persistir cuotas de membresía
type MembershipFeeRepository interface {
	Create(ctx context.Context, fee *models.MembershipFee) error
	Update(ctx context.Context, fee *models.MembershipFee) error
	FindByID(ctx context.Context, id uint) (*models.MembershipFee, error)
	FindByYear(ctx context.Context, year int) (*models.MembershipFee, error)
	FindByYearMonth(ctx context.Context, year, month int) (*models.MembershipFee, error) // DEPRECATED - mantener por compatibilidad
	FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error)
}
