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

// NewPaymentRepository crea una nueva instancia del repositorio de pagos
// que implementa la interfaz output.PaymentRepository.
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
	result := r.db.WithContext(ctx).
		Preload("Member").
		Preload("Family").
		Preload("MembershipFee").
		First(&payment, id)

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
		Preload("Member").
		Preload("Family").
		Preload("MembershipFee").
		Where("member_id = ? AND (payment_date BETWEEN ? AND ? OR payment_date IS NULL)", memberID, from, to).
		Find(&payments)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error finding member payments")
	}

	return payments, nil
}

func (r *paymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	var payments []models.Payment

	result := r.db.WithContext(ctx).
		Preload("Member").
		Preload("Family").
		Preload("MembershipFee").
		Where("family_id = ? AND (payment_date BETWEEN ? AND ? OR payment_date IS NULL)", familyID, from, to).
		Find(&payments)

	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error finding family payments")
	}

	return payments, nil
}

// HasInitialPayment checks if an initial payment already exists for the given member or family
func (r *paymentRepository) HasInitialPayment(ctx context.Context, memberID *uint, familyID *uint) (bool, error) {
	var exists bool

	// Build query to check for existing initial payment
	query := r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Select("1").
		Where("membership_fee_id IS NOT NULL")

	// Add condition for member or family
	switch {
	case memberID != nil && *memberID != 0:
		query = query.Where("member_id = ?", *memberID)
	case familyID != nil && *familyID != 0:
		query = query.Where("family_id = ?", *familyID)
	default:
		// Neither memberID nor familyID provided
		return false, appErrors.NewValidationError(
			"either memberID or familyID must be provided",
			map[string]string{
				"memberID": "required if familyID not provided",
				"familyID": "required if memberID not provided",
			},
		)
	}

	// Use SELECT EXISTS for optimal performance
	result := r.db.WithContext(ctx).Raw(
		"SELECT EXISTS(?)",
		query,
	).Scan(&exists)

	if result.Error != nil {
		return false, appErrors.DB(result.Error, "error checking for existing initial payment")
	}

	return exists, nil
}

// FindAll retrieves all payments matching the given filters
func (r *paymentRepository) FindAll(ctx context.Context, filters *output.PaymentRepositoryFilters) ([]models.Payment, error) {
	var payments []models.Payment

	query := r.db.WithContext(ctx).
		Preload("Member").
		Preload("Family").
		Preload("MembershipFee")

	if filters != nil {
		// Apply status filter
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}

		// Apply payment method filter (case-insensitive partial match)
		if filters.PaymentMethod != nil && *filters.PaymentMethod != "" {
			query = query.Where("LOWER(payment_method) LIKE LOWER(?)", "%"+*filters.PaymentMethod+"%")
		}

		// Apply date range filters
		if filters.StartDate != nil {
			query = query.Where("payment_date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("payment_date <= ?", *filters.EndDate)
		}

		// Apply amount range filters
		if filters.MinAmount != nil {
			query = query.Where("amount >= ?", *filters.MinAmount)
		}
		if filters.MaxAmount != nil {
			query = query.Where("amount <= ?", *filters.MaxAmount)
		}

		// Apply member filter
		if filters.MemberID != nil {
			query = query.Where("member_id = ?", *filters.MemberID)
		}

		// Apply family filter
		if filters.FamilyID != nil {
			query = query.Where("family_id = ?", *filters.FamilyID)
		}

		// Apply ordering
		if filters.OrderBy != "" {
			query = query.Order(filters.OrderBy)
		}

		// Apply pagination
		if filters.Limit > 0 {
			query = query.Limit(filters.Limit)
		}
		if filters.Offset > 0 {
			query = query.Offset(filters.Offset)
		}
	}

	if err := query.Find(&payments).Error; err != nil {
		return nil, appErrors.DB(err, "error finding payments")
	}

	return payments, nil
}

// CountAll returns the total count of payments matching the given filters
func (r *paymentRepository) CountAll(ctx context.Context, filters *output.PaymentRepositoryFilters) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.Payment{})

	if filters != nil {
		// Apply the same filters as FindAll (excluding pagination)
		if filters.Status != nil {
			query = query.Where("status = ?", *filters.Status)
		}
		if filters.PaymentMethod != nil && *filters.PaymentMethod != "" {
			query = query.Where("LOWER(payment_method) LIKE LOWER(?)", "%"+*filters.PaymentMethod+"%")
		}
		if filters.StartDate != nil {
			query = query.Where("payment_date >= ?", *filters.StartDate)
		}
		if filters.EndDate != nil {
			query = query.Where("payment_date <= ?", *filters.EndDate)
		}
		if filters.MinAmount != nil {
			query = query.Where("amount >= ?", *filters.MinAmount)
		}
		if filters.MaxAmount != nil {
			query = query.Where("amount <= ?", *filters.MaxAmount)
		}
		if filters.MemberID != nil {
			query = query.Where("member_id = ?", *filters.MemberID)
		}
		if filters.FamilyID != nil {
			query = query.Where("family_id = ?", *filters.FamilyID)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, appErrors.DB(err, "error counting payments")
	}

	return count, nil
}

// CreateWithTx creates a payment within a transaction (without creating CashFlow)
func (r *paymentRepository) CreateWithTx(ctx context.Context, tx output.Transaction, payment *models.Payment) error {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	result := gormTx.tx.WithContext(ctx).Create(payment)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "payment already exists")
		}
		return appErrors.DB(result.Error, "error creating payment")
	}
	return nil
}

// MembershipFeeRepository
type membershipFeeRepository struct {
	db *gorm.DB
}

// NewMembershipFeeRepository crea una nueva instancia del repositorio de cuotas de membresía
// que implementa la interfaz output.MembershipFeeRepository.
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

func (r *membershipFeeRepository) FindByYear(ctx context.Context, year int) (*models.MembershipFee, error) {
	var fee models.MembershipFee

	result := r.db.WithContext(ctx).
		Where("year = ?", year).
		First(&fee)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error finding annual membership fee")
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

// FindByYearWithTx finds a membership fee by year within a transaction
func (r *membershipFeeRepository) FindByYearWithTx(ctx context.Context, tx output.Transaction, year int) (*models.MembershipFee, error) {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return nil, appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	var fee models.MembershipFee
	result := gormTx.tx.WithContext(ctx).
		Where("year = ?", year).
		First(&fee)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, appErrors.DB(result.Error, "error finding annual membership fee")
	}

	return &fee, nil
}

// CreateWithTx creates a membership fee within a transaction
func (r *membershipFeeRepository) CreateWithTx(ctx context.Context, tx output.Transaction, fee *models.MembershipFee) error {
	gormTx, ok := tx.(*gormTransaction)
	if !ok {
		return appErrors.New(appErrors.ErrInternalError, "invalid transaction type")
	}

	result := gormTx.tx.WithContext(ctx).Create(fee)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "membership fee already exists for this period")
		}
		return appErrors.DB(result.Error, "error creating membership fee")
	}
	return nil
}
