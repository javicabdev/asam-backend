package output

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// ReportRepository define las operaciones de repositorio para reportes
type ReportRepository interface {
	// GetPendingPayments obtiene todos los pagos con estado PENDING
	GetPendingPayments(ctx context.Context) ([]models.Payment, error)

	// GetLastPaidPaymentForMember obtiene el último pago PAID de un socio
	GetLastPaidPaymentForMember(ctx context.Context, memberID uint) (*LastPaidPayment, error)

	// GetLastPaidPaymentForFamily obtiene el último pago PAID de una familia (a través del miembro origen)
	GetLastPaidPaymentForFamily(ctx context.Context, familyID uint) (*LastPaidPayment, error)
}

// LastPaidPayment contiene información del último pago realizado
type LastPaidPayment struct {
	ID          uint
	Amount      float64
	PaymentDate time.Time
	CreatedAt   time.Time
}
