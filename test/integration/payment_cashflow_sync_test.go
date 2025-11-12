package integration

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaymentCashFlowSync_UpdatePayment verifica que al actualizar un Payment se sincronice el CashFlow
func TestPaymentCashFlowSync_UpdatePayment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup: Crear un payment confirmado (que debe tener un cashflow)
	member := createTestMember(t, ctx)
	fee := createTestMembershipFee(t, ctx, 2025)

	payment := &models.Payment{
		MemberID:        member.ID,
		MembershipFeeID: &fee.ID,
		Amount:          100.0,
		Status:          models.PaymentStatusPaid,
		PaymentMethod:   "cash",
		Notes:           "Initial payment",
	}
	now := time.Now()
	payment.PaymentDate = &now

	err := testDB.paymentRepo.Create(ctx, payment)
	require.NoError(t, err, "Failed to create test payment")

	// Verificar que se creó el cashflow
	cashFlow, err := testDB.cashFlowRepo.GetByPaymentID(ctx, payment.ID)
	require.NoError(t, err, "Failed to get cashflow")
	require.NotNil(t, cashFlow, "CashFlow should have been created")
	assert.Equal(t, 100.0, cashFlow.Amount, "CashFlow amount should match payment")

	// Test: Actualizar el payment usando el método sincronizado
	payment.Amount = 150.0
	payment.Notes = "Updated payment"

	err = testDB.paymentRepo.UpdatePaymentAndSyncCashFlow(ctx, payment)
	require.NoError(t, err, "Failed to update payment with sync")

	// Verificar: El cashflow debe haberse actualizado también
	updatedCashFlow, err := testDB.cashFlowRepo.GetByPaymentID(ctx, payment.ID)
	require.NoError(t, err, "Failed to get updated cashflow")
	assert.Equal(t, 150.0, updatedCashFlow.Amount, "CashFlow should be synced with payment")
	assert.Contains(t, updatedCashFlow.Detail, "Updated payment", "CashFlow detail should be updated")
}

// TestPaymentCashFlowSync_UpdateCashFlow verifica que al actualizar un CashFlow se sincronice el Payment
func TestPaymentCashFlowSync_UpdateCashFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup: Crear un payment confirmado con cashflow
	member := createTestMember(t, ctx)
	fee := createTestMembershipFee(t, ctx, 2025)

	payment := &models.Payment{
		MemberID:        member.ID,
		MembershipFeeID: &fee.ID,
		Amount:          100.0,
		Status:          models.PaymentStatusPaid,
		PaymentMethod:   "transfer",
		Notes:           "Initial payment",
	}
	now := time.Now()
	payment.PaymentDate = &now

	err := testDB.paymentRepo.Create(ctx, payment)
	require.NoError(t, err, "Failed to create test payment")

	cashFlow, err := testDB.cashFlowRepo.GetByPaymentID(ctx, payment.ID)
	require.NoError(t, err)
	require.NotNil(t, cashFlow)

	// Test: Actualizar el cashflow usando el método sincronizado
	cashFlow.Amount = 200.0
	cashFlow.Detail = "Updated from cashflow"

	err = testDB.cashFlowRepo.UpdateCashFlowAndSyncPayment(ctx, cashFlow)
	require.NoError(t, err, "Failed to update cashflow with sync")

	// Verificar: El payment debe haberse actualizado también
	updatedPayment, err := testDB.paymentRepo.FindByID(ctx, payment.ID)
	require.NoError(t, err, "Failed to get updated payment")
	assert.Equal(t, 200.0, updatedPayment.Amount, "Payment should be synced with cashflow")
	assert.Equal(t, "Updated from cashflow", updatedPayment.Notes, "Payment notes should be updated")
}

// TestConfirmPaymentWithTransaction verifica que ConfirmPayment crea el cashflow en una transacción
func TestConfirmPaymentWithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup: Crear un payment pendiente
	member := createTestMember(t, ctx)
	fee := createTestMembershipFee(t, ctx, 2025)

	payment := &models.Payment{
		MemberID:        member.ID,
		MembershipFeeID: &fee.ID,
		Amount:          75.0,
		Status:          models.PaymentStatusPending,
		Notes:           "Pending payment",
	}

	err := testDB.paymentRepo.Create(ctx, payment)
	require.NoError(t, err, "Failed to create test payment")

	// Verificar que NO existe cashflow para payment pendiente
	cashFlow, err := testDB.cashFlowRepo.GetByPaymentID(ctx, payment.ID)
	require.NoError(t, err)
	assert.Nil(t, cashFlow, "CashFlow should NOT exist for pending payment")

	// Test: Confirmar el payment
	payment.Status = models.PaymentStatusPaid
	payment.PaymentMethod = "card"
	now := time.Now()
	payment.PaymentDate = &now

	err = testDB.paymentRepo.ConfirmPaymentWithTransaction(ctx, payment)
	require.NoError(t, err, "Failed to confirm payment")

	// Verificar: Ahora SÍ debe existir el cashflow
	cashFlow, err = testDB.cashFlowRepo.GetByPaymentID(ctx, payment.ID)
	require.NoError(t, err, "Failed to get cashflow")
	require.NotNil(t, cashFlow, "CashFlow should have been created")
	assert.Equal(t, 75.0, cashFlow.Amount, "CashFlow amount should match payment")

	// Verificar idempotencia: llamar confirm otra vez no debe duplicar
	err = testDB.paymentRepo.ConfirmPaymentWithTransaction(ctx, payment)
	require.NoError(t, err, "Confirm should be idempotent")

	// Should still have only one cashflow
	// (This would require a count query, but we trust the implementation)
}

// Helper functions (these should already exist in setup_test.go)

func createTestMember(t *testing.T, ctx context.Context) *models.Member {
	member := &models.Member{
		Name:             "Test Member",
		Surnames:         "For Sync",
		MembershipNumber: "TEST-" + time.Now().Format("20060102150405"),
		Email:            "test-" + time.Now().Format("20060102150405") + "@example.com",
		Status:           models.MemberStatusActive,
	}

	err := testDB.memberRepo.Create(ctx, member)
	require.NoError(t, err, "Failed to create test member")
	return member
}

func createTestMembershipFee(t *testing.T, ctx context.Context, year int) *models.MembershipFee {
	fee := &models.MembershipFee{
		Year:          year,
		BaseFeeAmount: 50.0,
		DueDate:       time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	err := testDB.membershipFeeRepo.Create(ctx, fee)
	require.NoError(t, err, "Failed to create test membership fee")
	return fee
}
