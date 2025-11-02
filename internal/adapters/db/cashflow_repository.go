// Package db provides database repository implementations for the ASAM backend.
// It contains GORM-based repositories that implement the output port interfaces.
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

// cashFlowRepository implementa la interfaz output.CashFlowRepository
type cashFlowRepository struct {
	db *gorm.DB
}

// NewCashFlowRepository crea una nueva instancia de CashFlowRepository
func NewCashFlowRepository(db *gorm.DB) output.CashFlowRepository {
	return &cashFlowRepository{db: db}
}

// Create implementa la creación de un nuevo movimiento
func (r *cashFlowRepository) Create(ctx context.Context, cashFlow *models.CashFlow) error {
	result := r.db.WithContext(ctx).Create(cashFlow)
	if result.Error != nil {
		// Manejar posibles errores de duplicación o violación de restricciones
		if IsDuplicateKeyError(result.Error) {
			return appErrors.New(appErrors.ErrDuplicateEntry, "cash flow entry already exists")
		}
		return appErrors.DB(result.Error, "error creating cash flow")
	}

	if result.RowsAffected == 0 {
		return appErrors.New(appErrors.ErrInternalError, "cash flow was not created")
	}

	return nil
}

// GetByID implementa la obtención de un movimiento por su ID
func (r *cashFlowRepository) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	var cashFlow models.CashFlow
	result := r.db.WithContext(ctx).
		Preload("Member").
		Preload("Payment").
		First(&cashFlow, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Patrón consistente: nil, nil para "no encontrado"
		}
		return nil, appErrors.DB(result.Error, "error finding cash flow by ID")
	}
	return &cashFlow, nil
}

// GetByPaymentID obtiene el movimiento de caja asociado a un pago específico
func (r *cashFlowRepository) GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error) {
	var cashFlow models.CashFlow
	result := r.db.WithContext(ctx).
		Preload("Member").
		Preload("Payment").
		Where("payment_id = ?", paymentID).
		First(&cashFlow)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Patrón consistente: nil, nil para "no encontrado"
		}
		return nil, appErrors.DB(result.Error, "error finding cash flow by payment ID")
	}
	return &cashFlow, nil
}

// Update implementa la actualización de un movimiento
func (r *cashFlowRepository) Update(ctx context.Context, cashFlow *models.CashFlow) error {
	result := r.db.WithContext(ctx).Save(cashFlow)
	if result.Error != nil {
		// Verificar errores de restricciones o integridad
		if IsConstraintViolationError(result.Error) {
			return appErrors.New(appErrors.ErrInvalidOperation, "invalid operation due to constraint violation")
		}
		return appErrors.DB(result.Error, "error updating cash flow")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("cash flow", nil)
	}

	return nil
}

// Delete implementa el borrado suave de un movimiento
func (r *cashFlowRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.CashFlow{}, id)
	if result.Error != nil {
		return appErrors.DB(result.Error, "error deleting cash flow")
	}

	if result.RowsAffected == 0 {
		return appErrors.NotFound("cash flow", nil)
	}

	return nil
}

// List implementa la obtención de movimientos con filtros
func (r *cashFlowRepository) List(ctx context.Context, filter output.CashFlowFilter) ([]*models.CashFlow, error) {
	var cashFlows []*models.CashFlow

	query := r.db.WithContext(ctx)

	// Aplicar filtros
	query = r.applyFilters(query, filter)

	// Aplicar ordenamiento
	if filter.OrderBy != "" {
		query = query.Order(filter.OrderBy)
	} else {
		// Orden por defecto
		query = query.Order("date DESC")
	}

	// Aplicar paginación
	if filter.PageSize > 0 {
		offset := 0
		if filter.Page > 0 {
			offset = (filter.Page - 1) * filter.PageSize
		}
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Cargar relaciones
	query = query.Preload("Payment").
		Preload("Member")

	result := query.Find(&cashFlows)
	if result.Error != nil {
		return nil, appErrors.DB(result.Error, "error listing cash flows")
	}

	return cashFlows, nil
}

// Count retorna el total de registros que coinciden con los filtros
func (r *cashFlowRepository) Count(ctx context.Context, filter output.CashFlowFilter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.CashFlow{})

	// Aplicar los mismos filtros que en List
	query = r.applyFilters(query, filter)

	result := query.Count(&count)
	if result.Error != nil {
		return 0, appErrors.DB(result.Error, "error counting cash flows")
	}

	return count, nil
}

// GetBalance calcula el balance actual (total ingresos - total gastos)
// Si memberID no es nil, calcula solo el balance de ese miembro
func (r *cashFlowRepository) GetBalance(ctx context.Context, memberID *uint) (*output.CashFlowBalance, error) {
	query := r.db.WithContext(ctx).Model(&models.CashFlow{})

	// Filtrar por miembro si se especifica
	if memberID != nil {
		query = query.Where("member_id = ?", *memberID)
	}

	var result struct {
		TotalIncome   float64
		TotalExpenses float64
	}

	err := query.Select(
		"COALESCE(SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) as total_income, "+
			"COALESCE(SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END), 0) as total_expenses",
	).Scan(&result).Error

	if err != nil {
		return nil, appErrors.DB(err, "error calculating balance")
	}

	balance := &output.CashFlowBalance{
		TotalIncome:    result.TotalIncome,
		TotalExpenses:  result.TotalExpenses,
		CurrentBalance: result.TotalIncome - result.TotalExpenses,
	}

	return balance, nil
}

// GetStats obtiene estadísticas por categoría y tendencia mensual
func (r *cashFlowRepository) GetStats(ctx context.Context, startDate, endDate time.Time, memberID *uint) (*output.CashFlowStats, error) {
	stats := &output.CashFlowStats{}

	// 1. Obtener ingresos por categoría
	var incomeResults []struct {
		Category models.OperationType
		Amount   float64
		Count    int
	}

	incomeQuery := r.db.WithContext(ctx).Model(&models.CashFlow{}).
		Select("operation_type as category, SUM(amount) as amount, COUNT(*) as count").
		Where("date BETWEEN ? AND ? AND amount > 0", startDate, endDate).
		Group("operation_type")

	if memberID != nil {
		incomeQuery = incomeQuery.Where("member_id = ?", *memberID)
	}

	err := incomeQuery.Scan(&incomeResults).Error

	if err != nil {
		return nil, appErrors.DB(err, "error getting income by category")
	}

	for _, result := range incomeResults {
		stats.IncomeByCategory = append(stats.IncomeByCategory, output.CategoryAmount{
			Category: result.Category,
			Amount:   result.Amount,
			Count:    result.Count,
		})
	}

	// 2. Obtener gastos por categoría
	var expenseResults []struct {
		Category models.OperationType
		Amount   float64
		Count    int
	}

	expenseQuery := r.db.WithContext(ctx).Model(&models.CashFlow{}).
		Select("operation_type as category, SUM(ABS(amount)) as amount, COUNT(*) as count").
		Where("date BETWEEN ? AND ? AND amount < 0", startDate, endDate).
		Group("operation_type")

	if memberID != nil {
		expenseQuery = expenseQuery.Where("member_id = ?", *memberID)
	}

	err = expenseQuery.Scan(&expenseResults).Error

	if err != nil {
		return nil, appErrors.DB(err, "error getting expenses by category")
	}

	for _, result := range expenseResults {
		stats.ExpensesByCategory = append(stats.ExpensesByCategory, output.CategoryAmount{
			Category: result.Category,
			Amount:   result.Amount,
			Count:    result.Count,
		})
	}

	// 3. Obtener tendencia mensual
	var monthlyResults []struct {
		Month    string
		Income   float64
		Expenses float64
	}

	err = r.db.WithContext(ctx).Model(&models.CashFlow{}).
		Select(
			"TO_CHAR(date, 'YYYY-MM') as month, "+
				"SUM(CASE WHEN amount > 0 THEN amount ELSE 0 END) as income, "+
				"SUM(CASE WHEN amount < 0 THEN ABS(amount) ELSE 0 END) as expenses",
		).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("month").
		Scan(&monthlyResults).Error

	if err != nil {
		return nil, appErrors.DB(err, "error getting monthly trend")
	}

	for _, result := range monthlyResults {
		stats.MonthlyTrend = append(stats.MonthlyTrend, output.MonthlyAmount{
			Month:    result.Month,
			Income:   result.Income,
			Expenses: result.Expenses,
			Balance:  result.Income - result.Expenses,
		})
	}

	return stats, nil
}

// ExistsByPaymentID verifica si ya existe un cash_flow para un payment_id (para idempotencia)
func (r *cashFlowRepository) ExistsByPaymentID(ctx context.Context, paymentID uint) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.CashFlow{}).
		Where("payment_id = ?", paymentID).
		Count(&count)

	if result.Error != nil {
		return false, appErrors.DB(result.Error, "error checking if cash flow exists by payment ID")
	}

	return count > 0, nil
}

// applyFilters aplica los filtros comunes a una query
func (r *cashFlowRepository) applyFilters(query *gorm.DB, filter output.CashFlowFilter) *gorm.DB {
	if filter.MemberID != nil {
		query = query.Where("member_id = ?", *filter.MemberID)
	}
	if filter.PaymentID != nil {
		query = query.Where("payment_id = ?", *filter.PaymentID)
	}
	if filter.OperationType != nil {
		query = query.Where("operation_type = ?", *filter.OperationType)
	}
	if filter.Category != nil {
		switch *filter.Category {
		case "INGRESO":
			query = query.Where("amount > 0")
		case "GASTO":
			query = query.Where("amount < 0")
		}
	}
	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}

	return query
}
