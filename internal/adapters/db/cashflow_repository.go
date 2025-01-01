package db

import (
	"context"
	"errors"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"gorm.io/gorm"
)

// CashFlowRepository implementa la interfaz output.CashFlowRepository
type CashFlowRepository struct {
	db *gorm.DB
}

// NewCashFlowRepository crea una nueva instancia de CashFlowRepository
func NewCashFlowRepository(db *gorm.DB) *CashFlowRepository {
	return &CashFlowRepository{db: db}
}

// Create implementa la creación de un nuevo movimiento
func (r *CashFlowRepository) Create(ctx context.Context, cashFlow *models.CashFlow) error {
	result := r.db.WithContext(ctx).Create(cashFlow)
	return result.Error
}

// GetByID implementa la obtención de un movimiento por su ID
func (r *CashFlowRepository) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	var cashFlow models.CashFlow
	result := r.db.WithContext(ctx).First(&cashFlow, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &cashFlow, nil
}

// GetByPaymentID obtiene el movimiento de caja asociado a un pago específico
func (r *CashFlowRepository) GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error) {
	var cashFlow models.CashFlow
	result := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		First(&cashFlow)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &cashFlow, nil
}

// Update implementa la actualización de un movimiento
func (r *CashFlowRepository) Update(ctx context.Context, cashFlow *models.CashFlow) error {
	result := r.db.WithContext(ctx).Save(cashFlow)
	return result.Error
}

// Delete implementa el borrado suave de un movimiento
func (r *CashFlowRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.CashFlow{}, id)
	return result.Error
}

// List implementa la obtención de movimientos con filtros
// List implementa la obtención de movimientos con filtros
func (r *CashFlowRepository) List(ctx context.Context, filter output.CashFlowFilter) ([]*models.CashFlow, error) {
	var cashFlows []*models.CashFlow

	query := r.db.WithContext(ctx)

	// Aplicar filtros
	if filter.MemberID != nil {
		query = query.Where("member_id = ?", *filter.MemberID)
	}
	if filter.FamilyID != nil {
		query = query.Where("family_id = ?", *filter.FamilyID)
	}
	if filter.PaymentID != nil {
		query = query.Where("payment_id = ?", *filter.PaymentID)
	}
	if filter.OperationType != nil {
		query = query.Where("operation_type = ?", *filter.OperationType)
	}
	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}

	// Aplicar ordenamiento
	if filter.OrderBy != "" {
		query = query.Order(filter.OrderBy)
	} else {
		// Orden por defecto
		query = query.Order("date DESC")
	}

	// Aplicar paginación
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Cargar relaciones
	query = query.Preload("Payment").
		Preload("Member").
		Preload("Family")

	result := query.Find(&cashFlows)
	return cashFlows, result.Error
}

// GetBalance implementa el cálculo del balance actual
func (r *CashFlowRepository) GetBalance(ctx context.Context) (float64, error) {
	var balance float64
	result := r.db.WithContext(ctx).
		Model(&models.CashFlow{}).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&balance)
	return balance, result.Error
}
