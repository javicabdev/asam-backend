package output

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"time"
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

	// List obtiene una lista de movimientos con filtros opcionales
	List(ctx context.Context, filter CashFlowFilter) ([]*models.CashFlow, error)

	// GetBalance calcula el balance actual
	GetBalance(ctx context.Context) (float64, error)
}

// CashFlowFilter define los filtros disponibles para buscar movimientos
type CashFlowFilter struct {
	MemberID      *uint
	FamilyID      *uint
	PaymentID     *uint
	OperationType *models.OperationType
	StartDate     *time.Time
	EndDate       *time.Time
	Page          int
	PageSize      int
	OrderBy       string // Añadido
}
