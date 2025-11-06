package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository crea una nueva instancia del repositorio de reportes
func NewReportRepository(db *gorm.DB) output.ReportRepository {
	return &reportRepository{db: db}
}

// GetPendingPayments obtiene todos los pagos con estado PENDING
func (r *reportRepository) GetPendingPayments(ctx context.Context) ([]models.Payment, error) {
	var payments []models.Payment

	result := r.db.WithContext(ctx).
		Preload("Member").
		Where("status = ?", models.PaymentStatusPending).
		Order("created_at ASC").
		Find(&payments)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error getting pending payments")
	}

	return payments, nil
}

// GetLastPaidPaymentForMember obtiene el último pago PAID de un socio
func (r *reportRepository) GetLastPaidPaymentForMember(ctx context.Context, memberID uint) (*output.LastPaidPayment, error) {
	var payment models.Payment

	result := r.db.WithContext(ctx).
		Where("member_id = ? AND status = ?", memberID, models.PaymentStatusPaid).
		Order("payment_date DESC").
		First(&payment)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No hay pagos pagados, no es un error
		}
		return nil, appErrors.DB(result.Error, "error getting last paid payment for member")
	}

	// Verificar que payment_date no sea nil
	if payment.PaymentDate == nil {
		return nil, nil // Pago PAID sin fecha, datos inconsistentes
	}

	return &output.LastPaidPayment{
		ID:          payment.ID,
		Amount:      payment.Amount,
		PaymentDate: *payment.PaymentDate,
		CreatedAt:   payment.CreatedAt,
	}, nil
}

// GetLastPaidPaymentForFamily obtiene el último pago PAID de una familia (a través del miembro origen)
func (r *reportRepository) GetLastPaidPaymentForFamily(ctx context.Context, familyID uint) (*output.LastPaidPayment, error) {
	// Primero obtenemos el miembro origen de la familia
	var family models.Family
	if err := r.db.WithContext(ctx).First(&family, familyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, appErrors.NotFound("family", nil)
		}
		return nil, appErrors.DB(err, "error getting family")
	}

	if family.MiembroOrigenID == nil {
		return nil, nil // Familia sin miembro origen, no tiene pagos
	}

	// Obtener el último pago del miembro origen
	return r.GetLastPaidPaymentForMember(ctx, *family.MiembroOrigenID)
}
