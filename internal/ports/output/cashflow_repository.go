package output

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// CashFlowRepository define las operaciones disponibles para el repositorio de CashFlow
type CashFlowRepository interface {
	// Create crea un nuevo movimiento de caja
	Create(ctx context.Context, cashFlow *models.CashFlow) error

	// GetByID obtiene un movimiento por su ID
	GetByID(ctx context.Context, id uint) (*models.CashFlow, error)

	// GetByPaymentID obtiene el movimiento de caja asociado a un pago específico
	GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error)

	// Update actualiza un movimiento existente
	Update(ctx context.Context, cashFlow *models.CashFlow) error

	// Delete elimina un movimiento (soft delete)
	Delete(ctx context.Context, id uint) error

	// List obtiene una lista de movimientos con filtros opcionales y paginación
	List(ctx context.Context, filter CashFlowFilter) ([]*models.CashFlow, error)

	// Count retorna el total de registros que coinciden con los filtros
	Count(ctx context.Context, filter CashFlowFilter) (int64, error)

	// GetBalance calcula el balance actual (total ingresos - total gastos)
	// Si memberID no es nil, calcula solo el balance de ese miembro
	GetBalance(ctx context.Context, memberID *uint) (*CashFlowBalance, error)

	// GetStats obtiene estadísticas por categoría y tendencia mensual
	// Si memberID no es nil, calcula solo las estadísticas de ese miembro
	GetStats(ctx context.Context, startDate, endDate time.Time, memberID *uint) (*CashFlowStats, error)

	// ExistsByPaymentID verifica si ya existe un cash_flow para un payment_id (para idempotencia)
	ExistsByPaymentID(ctx context.Context, paymentID uint) (bool, error)

	// ListWithRunningBalance obtiene una lista de movimientos con running_balance calculado
	// usando una única query SQL con window functions para máxima eficiencia.
	// El running_balance se calcula considerando:
	// - El balance inicial (suma de todos los movimientos anteriores al filtro de fecha)
	// - Los filtros aplicados (operation_type, member_id, etc.)
	// - El ordenamiento cronológico (date ASC, created_at ASC)
	ListWithRunningBalance(ctx context.Context, filter CashFlowFilter) ([]*models.CashFlow, error)

	// UpdateCashFlowAndSyncPayment actualiza un cashflow y sincroniza su payment asociado en una transacción
	// Si el cashflow tiene un payment vinculado (PaymentID != nil), lo actualiza con los mismos datos
	// Garantiza consistencia de datos mediante transacción ACID
	UpdateCashFlowAndSyncPayment(ctx context.Context, cashFlow *models.CashFlow) error
}

// CashFlowFilter define los filtros disponibles para buscar movimientos
type CashFlowFilter struct {
	MemberID      *uint
	PaymentID     *uint
	OperationType *models.OperationType
	Category      *string // "INGRESO" o "GASTO"
	StartDate     *time.Time
	EndDate       *time.Time
	Page          int
	PageSize      int
	OrderBy       string
}

// CashFlowBalance representa el balance de ingresos y gastos
type CashFlowBalance struct {
	TotalIncome    float64
	TotalExpenses  float64
	CurrentBalance float64
}

// CategoryAmount representa el monto total por categoría
type CategoryAmount struct {
	Category models.OperationType
	Amount   float64
	Count    int
}

// MonthlyAmount representa el balance mensual
type MonthlyAmount struct {
	Month    string // Formato: "2025-10"
	Income   float64
	Expenses float64
	Balance  float64
}

// CashFlowStats representa las estadísticas de cash flow
type CashFlowStats struct {
	IncomeByCategory   []CategoryAmount
	ExpensesByCategory []CategoryAmount
	MonthlyTrend       []MonthlyAmount
}
