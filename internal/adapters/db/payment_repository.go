package db

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) output.PaymentRepository {
	return &paymentRepository{db: db}
}

// Create crea un nuevo pago y su correspondiente movimiento en el flujo de caja
func (r *paymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Crear el pago
		if err := tx.Create(payment).Error; err != nil {
			if IsDuplicateKeyError(err) {
				return appErrors.New(appErrors.ErrDuplicateEntry, "payment already exists")
			}
			return appErrors.DB(err, "error creating payment")
		}

		// 2. Crear el movimiento de caja correspondiente
		cashFlow, err := models.NewFromPayment(payment)
		if err != nil {
			return appErrors.Wrap(err, appErrors.ErrInvalidOperation, "error creating cash flow from payment")
		}

		if err := tx.Create(cashFlow).Error; err != nil {
			return appErrors.DB(err, "error creating cash flow entry")
		}

		return nil
	})
}

func (r *paymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	if err := payment.Validate(); err != nil {
		// Si ya es un AppError, pasarlo directamente
		if appErr, ok := appErrors.AsAppError(err); ok {
			return appErr
		}
		return appErrors.Validation("invalid payment data", "", err.Error())
	}

	result := r.db.WithContext(ctx).Save(payment)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error updating payment")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("payment", nil)
	}

	return nil
}

func (r *paymentRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Payment{}, id)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error deleting payment")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("payment", nil)
	}

	return nil
}

func (r *paymentRepository) FindByID(ctx context.Context, id uint) (*models.Payment, error) {
	var payment models.Payment
	result := r.db.WithContext(ctx).First(&payment, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Patrón consistente: nil, nil para "no encontrado"
		}
		return nil, appErrors.DB(result.Error, "error finding payment")
	}

	return &payment, nil
}

func (r *paymentRepository) FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
	var payments []models.Payment

	result := r.db.WithContext(ctx).
		Where("member_id = ? AND payment_date BETWEEN ? AND ?", memberID, from, to).
		Find(&payments)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error finding member payments")
	}

	return payments, nil
}

func (r *paymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	var payments []models.Payment

	result := r.db.WithContext(ctx).
		Where("family_id = ? AND payment_date BETWEEN ? AND ?", familyID, from, to).
		Find(&payments)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error finding family payments")
	}

	return payments, nil
}

// MembershipFeeRepository
type membershipFeeRepository struct {
	db *gorm.DB
}

func NewMembershipFeeRepository(db *gorm.DB) output.MembershipFeeRepository {
	return &membershipFeeRepository{db: db}
}

func (r *membershipFeeRepository) Create(ctx context.Context, fee *models.MembershipFee) error {
	result := r.db.WithContext(ctx).Create(fee)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "membership fee already exists for this period")
		}
		return appErrors.DB(result.Error, "error creating membership fee")
	}
	return nil
}

func (r *membershipFeeRepository) Update(ctx context.Context, fee *models.MembershipFee) error {
	result := r.db.WithContext(ctx).Save(fee)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error updating membership fee")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("membership fee", nil)
	}

	return nil
}

func (r *membershipFeeRepository) FindByYearMonth(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	var fee models.MembershipFee

	result := r.db.WithContext(ctx).
		Where("year = ? AND month = ?", year, month).
		First(&fee)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Patrón consistente: nil, nil para "no encontrado"
		}
		return nil, appErrors.DB(result.Error, "error finding membership fee")
	}

	return &fee, nil
}

func (r *membershipFeeRepository) FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error) {
	var fees []models.MembershipFee

	// Creamos una query que busca cuotas pendientes
	query := r.db.WithContext(ctx).
		Joins("LEFT JOIN payments ON membership_fees.payment_id = payments.id")

	// Si memberID es 0, traemos todas las cuotas pendientes
	if memberID != 0 {
		query = query.Where("payments.member_id = ?", memberID)
	}

	// Filtramos por estado pendiente
	result := query.Where("membership_fees.status = ?", models.PaymentStatusPending).
		Find(&fees)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error finding pending membership fees")
	}

	return fees, nil
}

func (r *membershipFeeRepository) FindByID(ctx context.Context, id uint) (*models.MembershipFee, error) {
	var fee models.MembershipFee

	result := r.db.WithContext(ctx).First(&fee, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Patrón consistente: nil, nil para "no encontrado"
		}
		return nil, appErrors.DB(result.Error, "error finding membership fee")
	}

	return &fee, nil
}
