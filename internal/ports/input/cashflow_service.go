package input

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// CashFlowService define las operaciones disponibles para la gestión del flujo de caja
type CashFlowService interface {
	// RegisterMovement Operaciones básicas
	RegisterMovement(ctx context.Context, movement *models.CashFlow) error
	GetMovement(ctx context.Context, id uint) (*models.CashFlow, error)
	UpdateMovement(ctx context.Context, movement *models.CashFlow) error
	DeleteMovement(ctx context.Context, id uint) error
	GetMovementsByPeriod(ctx context.Context, filter CashFlowFilter) ([]*models.CashFlow, error)

	// GetCurrentBalance Gestión de balance
	GetCurrentBalance(ctx context.Context) (*BalanceReport, error)
	GetBalanceByPeriod(ctx context.Context, startDate, endDate time.Time) (*BalanceReport, error)
	ValidateBalance(ctx context.Context) (*BalanceValidation, error)

	// GetFinancialReport Reportes y análisis
	GetFinancialReport(ctx context.Context, reportType ReportType, period Period) (*FinancialReport, error)
	GetCashFlowTrends(ctx context.Context, period Period) (*TrendAnalysis, error)
	GetProjections(ctx context.Context, months int) (*FinancialProjection, error)

	// GetFinancialAlerts Alertas y monitoreo
	GetFinancialAlerts(ctx context.Context) ([]FinancialAlert, error)
}

// Period representa un período de tiempo para reportes
type Period struct {
	StartDate time.Time
	EndDate   time.Time
}

// ReportType define los tipos de reportes disponibles
type ReportType string

const (
	// ReportTypeBalance reporte de balance
	ReportTypeBalance ReportType = "balance"
	// ReportTypeIncome reporte de ingresos
	ReportTypeIncome ReportType = "income"
	// ReportTypeCashFlow reporte de flujo de caja
	ReportTypeCashFlow ReportType = "cashflow"
)

// BalanceReport contiene información detallada del balance
type BalanceReport struct {
	CurrentBalance   float64
	TotalIncome      float64
	TotalExpenses    float64
	PeriodStart      time.Time
	PeriodEnd        time.Time
	MovementsSummary []MovementSummary
}

// MovementSummary representa un resumen de movimientos por categoría
type MovementSummary struct {
	OperationType models.OperationType
	Total         float64
	Count         int
}

// BalanceValidation contiene el resultado de la validación del balance
type BalanceValidation struct {
	IsValid         bool
	ExpectedBalance float64
	ActualBalance   float64
	Discrepancy     float64
	Details         string
}

// FinancialReport contiene información detallada para reportes financieros
type FinancialReport struct {
	Type       ReportType
	Period     Period
	Data       map[string]float64
	Categories []CategorySummary
	Totals     TotalsSummary
}

// CategorySummary representa un resumen por categoría
type CategorySummary struct {
	Category   string
	Amount     float64
	Percentage float64
}

// TotalsSummary contiene los totales del reporte
type TotalsSummary struct {
	Income    float64
	Expenses  float64
	NetResult float64
}

// TrendAnalysis contiene análisis de tendencias
type TrendAnalysis struct {
	Period        Period
	MonthlyTrends []MonthlyTrend
	Indicators    map[string]float64
}

// MonthlyTrend representa la tendencia de un mes
type MonthlyTrend struct {
	Month    time.Time
	Income   float64
	Expenses float64
	Balance  float64
	Growth   float64
}

// FinancialProjection contiene proyecciones financieras
type FinancialProjection struct {
	Months      int
	Projections []MonthlyProjection
	Confidence  float64
}

// MonthlyProjection representa la proyección para un mes
type MonthlyProjection struct {
	Month            time.Time
	ExpectedIncome   float64
	ExpectedExpenses float64
	ExpectedBalance  float64
	Variance         float64
}

// FinancialAlert representa una alerta financiera
type FinancialAlert struct {
	Type         string
	Severity     string
	Message      string
	Threshold    float64
	CurrentValue float64
	CreatedAt    time.Time
}

// CashFlowFilter define los filtros para consultar movimientos de caja
type CashFlowFilter struct {
	MemberID      *uint
	StartDate     *time.Time
	EndDate       *time.Time
	OperationType *models.OperationType
	Page          int
	PageSize      int
	OrderBy       string
}
