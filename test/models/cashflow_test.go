package models_test

import (
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Helper para crear un CashFlow válido
func createValidCashFlow() *models.CashFlow {
	return &models.CashFlow{
		OperationType: models.OperationTypeMembershipFee,
		Amount:        100.0,
		Date:          time.Now(),
		Detail:        "Ingreso por cuota de membresía",
	}
}

// Tests de validaciones básicas
func TestCashFlowValidation(t *testing.T) {
	cashFlow := createValidCashFlow()

	// Caso válido
	assert.NoError(t, cashFlow.Validate())

	// Caso: OperationType inválido
	cashFlow.OperationType = "invalid"
	err := cashFlow.Validate()
	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidOperationType, err)

	// Caso: Monto inválido
	cashFlow.OperationType = models.OperationTypeMembershipFee // Restaurar OperationType válido
	cashFlow.Amount = 0
	err = cashFlow.Validate()
	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidAmount, err)

	// Caso: Fecha inválida
	cashFlow.Amount = 100.0 // Restaurar monto válido
	cashFlow.Date = time.Time{}
	err = cashFlow.Validate()
	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidDate, err)

	// Caso: Detalle faltante
	cashFlow.Date = time.Now() // Restaurar fecha válida
	cashFlow.Detail = ""
	err = cashFlow.Validate()
	assert.Error(t, err)
	assert.Equal(t, models.ErrMissingDetail, err)
}

// Tests de validación de ingresos y gastos
func TestCashFlowValidation_IncomeExpense(t *testing.T) {
	cashFlow := createValidCashFlow()

	// Caso: Ingreso con monto negativo
	cashFlow.OperationType = models.OperationTypeMembershipFee
	cashFlow.Amount = -100.0
	err := cashFlow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "los ingresos deben tener monto positivo")

	// Caso: Gasto con monto positivo
	cashFlow.OperationType = models.OperationTypeCurrentExpense
	cashFlow.Amount = 100.0
	err = cashFlow.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "los gastos deben tener monto negativo")
}

// Tests de hooks de GORM
func TestCashFlow_BeforeCreate(t *testing.T) {
	cashFlow := createValidCashFlow()

	// Caso válido
	assert.NoError(t, cashFlow.BeforeCreate(nil))

	// Caso inválido
	cashFlow.Amount = 0
	assert.Error(t, cashFlow.BeforeCreate(nil))
}

func TestCashFlow_BeforeUpdate(t *testing.T) {
	cashFlow := createValidCashFlow()

	// Caso válido
	assert.NoError(t, cashFlow.BeforeUpdate(nil))

	// Caso inválido
	cashFlow.OperationType = "invalid"
	assert.Error(t, cashFlow.BeforeUpdate(nil))
}

// Tests de creación desde un Payment
func TestCashFlow_NewFromPayment(t *testing.T) {
	payment := &models.Payment{
		Model:       gorm.Model{ID: 1}, // Inicializar gorm.Model
		MemberID:    1,
		FamilyID:    nil,
		Amount:      50.0,
		PaymentDate: time.Now(),
		MembershipFee: &models.MembershipFee{
			Year:  2023,
			Month: 1,
		},
	}

	cashFlow, err := models.NewFromPayment(payment)
	assert.NoError(t, err)

	// Validar los campos creados
	assert.Equal(t, payment.ID, *cashFlow.PaymentID)
	assert.Equal(t, payment.MemberID, *cashFlow.MemberID)
	assert.Equal(t, payment.Amount, cashFlow.Amount)
	assert.Equal(t, "Cuota de membresía - 1/2023", cashFlow.Detail)
	assert.Equal(t, models.OperationTypeMembershipFee, cashFlow.OperationType)
}

// Tests de creación desde un Payment inválido
func TestCashFlow_NewFromPayment_Invalid(t *testing.T) {
	// Caso: Payment es nil
	cashFlow, err := models.NewFromPayment(nil)
	assert.Nil(t, cashFlow)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment no puede ser nil")
}
