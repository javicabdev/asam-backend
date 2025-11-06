package integration

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunningBalance_Case1_NoFiltersNoPagination verifica el cálculo de running_balance
// sin filtros y sin paginación (Caso 1 de la especificación)
func TestRunningBalance_Case1_NoFiltersNoPagination(t *testing.T) {
	// Arrange: Crear datos de prueba
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Crear transacciones de prueba
	transactions := []struct {
		date          string
		operationType models.OperationType
		amount        float64
		detail        string
	}{
		{"2024-01-01", models.OperationTypeMembershipFee, 100, "Cuota enero"},
		{"2024-01-02", models.OperationTypeDonation, 50, "Donación"},
		{"2024-01-03", models.OperationTypeBankFees, -10, "Comisión bancaria"},
		{"2024-01-04", models.OperationTypeRepatriation, -1500, "Repatriación"},
	}

	for _, tx := range transactions {
		date, _ := time.Parse("2006-01-02", tx.date)
		cashFlow := &models.CashFlow{
			OperationType: tx.operationType,
			Amount:        tx.amount,
			Date:          date,
			Detail:        tx.detail,
		}
		err := service.RegisterMovement(ctx, cashFlow)
		require.NoError(t, err, "Error creating test transaction")
	}

	// Act: Obtener movimientos sin filtros
	filter := input.CashFlowFilter{
		Page:     1,
		PageSize: 100,
		OrderBy:  "date ASC",
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, movements, 4, "Should return 4 transactions")

	// Verificar running_balance de cada transacción
	expectedBalances := []float64{100, 150, 140, -1360}
	for i, expected := range expectedBalances {
		assert.Equal(t, expected, movements[i].RunningBalance,
			"Transaction %d should have running_balance %.2f, got %.2f",
			i+1, expected, movements[i].RunningBalance)
	}
}

// TestRunningBalance_Case2_WithDateFilter verifica el cálculo con filtro de fecha
// (Caso 2 de la especificación: Balance inicial debe incluir transacciones anteriores)
func TestRunningBalance_Case2_WithDateFilter(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Crear transacciones de prueba
	transactions := []struct {
		date          string
		operationType models.OperationType
		amount        float64
		detail        string
	}{
		{"2024-01-01", models.OperationTypeMembershipFee, 100, "Cuota 1"},
		{"2024-01-02", models.OperationTypeDonation, 50, "Donación"},
		{"2024-01-03", models.OperationTypeBankFees, -10, "Comisión"},
		{"2024-01-04", models.OperationTypeRepatriation, -1500, "Repatriación"},
	}

	for _, tx := range transactions {
		date, _ := time.Parse("2006-01-02", tx.date)
		cashFlow := &models.CashFlow{
			OperationType: tx.operationType,
			Amount:        tx.amount,
			Date:          date,
			Detail:        tx.detail,
		}
		err := service.RegisterMovement(ctx, cashFlow)
		require.NoError(t, err)
	}

	// Act: Filtrar desde 2024-01-03 (debería incluir balance inicial de 150)
	startDate, _ := time.Parse("2006-01-02", "2024-01-03")
	filter := input.CashFlowFilter{
		StartDate: &startDate,
		Page:      1,
		PageSize:  100,
		OrderBy:   "date ASC",
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, movements, 2, "Should return 2 transactions (from 2024-01-03)")

	// Verificar que el primer running_balance incluye el balance inicial
	// Balance inicial: 100 + 50 = 150
	// Primera transacción: 150 - 10 = 140
	assert.Equal(t, 140.0, movements[0].RunningBalance,
		"First transaction should have running_balance 140 (including initial balance of 150)")

	// Segunda transacción: 140 - 1500 = -1360
	assert.Equal(t, -1360.0, movements[1].RunningBalance,
		"Second transaction should have running_balance -1360")
}

// TestRunningBalance_Case3_WithPagination verifica el cálculo con paginación
// (Caso 3 de la especificación: Página 2 debe incluir balance de página 1)
func TestRunningBalance_Case3_WithPagination(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Crear 6 transacciones (3 páginas de 2 elementos cada una)
	transactions := []struct {
		date          string
		operationType models.OperationType
		amount        float64
	}{
		{"2024-01-01", models.OperationTypeMembershipFee, 100},
		{"2024-01-02", models.OperationTypeDonation, 50},
		{"2024-01-03", models.OperationTypeBankFees, -10},
		{"2024-01-04", models.OperationTypeMembershipFee, 200},
		{"2024-01-05", models.OperationTypeBankFees, -50},
		{"2024-01-06", models.OperationTypeMembershipFee, 100},
	}

	for _, tx := range transactions {
		date, _ := time.Parse("2006-01-02", tx.date)
		cashFlow := &models.CashFlow{
			OperationType: tx.operationType,
			Amount:        tx.amount,
			Date:          date,
			Detail:        "Test transaction",
		}
		err := service.RegisterMovement(ctx, cashFlow)
		require.NoError(t, err)
	}

	// Act: Obtener página 2 (elementos 3 y 4)
	filter := input.CashFlowFilter{
		Page:     2,
		PageSize: 2,
		OrderBy:  "date ASC",
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, movements, 2, "Should return 2 transactions (page 2)")

	// Verificar running_balance considerando transacciones anteriores
	// Tx 1: 100, Tx 2: 150, Tx 3: 140, Tx 4: 340
	assert.Equal(t, 140.0, movements[0].RunningBalance,
		"Transaction 3 should have running_balance 140 (100 + 50 - 10)")

	assert.Equal(t, 340.0, movements[1].RunningBalance,
		"Transaction 4 should have running_balance 340 (140 + 200)")
}

// TestRunningBalance_Case4_FilterByOperationType verifica el filtro por tipo de operación
// (Caso 4 de la especificación)
func TestRunningBalance_Case4_FilterByOperationType(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Crear transacciones mixtas
	transactions := []struct {
		date          string
		operationType models.OperationType
		amount        float64
	}{
		{"2024-01-01", models.OperationTypeMembershipFee, 100},
		{"2024-01-02", models.OperationTypeDonation, 50},
		{"2024-01-03", models.OperationTypeBankFees, -10},
		{"2024-01-04", models.OperationTypeMembershipFee, 200},
	}

	for _, tx := range transactions {
		date, _ := time.Parse("2006-01-02", tx.date)
		cashFlow := &models.CashFlow{
			OperationType: tx.operationType,
			Amount:        tx.amount,
			Date:          date,
			Detail:        "Test",
		}
		err := service.RegisterMovement(ctx, cashFlow)
		require.NoError(t, err)
	}

	// Act: Filtrar solo INGRESO_CUOTA
	opType := models.OperationTypeMembershipFee
	filter := input.CashFlowFilter{
		OperationType: &opType,
		Page:          1,
		PageSize:      100,
		OrderBy:       "date ASC",
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, movements, 2, "Should return only INGRESO_CUOTA transactions")

	// Verificar que running_balance solo considera transacciones del tipo filtrado
	assert.Equal(t, 100.0, movements[0].RunningBalance,
		"First INGRESO_CUOTA should have running_balance 100")

	assert.Equal(t, 300.0, movements[1].RunningBalance,
		"Second INGRESO_CUOTA should have running_balance 300 (100 + 200)")
}

// TestRunningBalance_Case5_NegativeBalance verifica saldos negativos
// (Caso 5 de la especificación: Edge case - Primera transacción negativa)
func TestRunningBalance_Case5_NegativeBalance(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Crear transacciones comenzando con gasto
	transactions := []struct {
		date          string
		operationType models.OperationType
		amount        float64
	}{
		{"2024-01-01", models.OperationTypeRepatriation, -1500},
		{"2024-01-02", models.OperationTypeMembershipFee, 100},
	}

	for _, tx := range transactions {
		date, _ := time.Parse("2006-01-02", tx.date)
		cashFlow := &models.CashFlow{
			OperationType: tx.operationType,
			Amount:        tx.amount,
			Date:          date,
			Detail:        "Test",
		}
		err := service.RegisterMovement(ctx, cashFlow)
		require.NoError(t, err)
	}

	// Act
	filter := input.CashFlowFilter{
		Page:     1,
		PageSize: 100,
		OrderBy:  "date ASC",
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	require.Len(t, movements, 2)

	// Verificar saldos negativos
	assert.Equal(t, -1500.0, movements[0].RunningBalance,
		"First transaction should have negative running_balance -1500")

	assert.Equal(t, -1400.0, movements[1].RunningBalance,
		"Second transaction should have running_balance -1400")
}

// TestRunningBalance_Case6_EmptyResult verifica resultado vacío
// (Caso 6 de la especificación: Edge case - Sin transacciones)
func TestRunningBalance_Case6_EmptyResult(t *testing.T) {
	// Arrange
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := setupCashFlowRepository(db)
	service := setupCashFlowService(repo)

	// Act: Filtrar con fecha futura (sin resultados)
	futureDate, _ := time.Parse("2006-01-02", "2099-01-01")
	filter := input.CashFlowFilter{
		StartDate: &futureDate,
		Page:      1,
		PageSize:  100,
	}
	movements, err := service.GetMovementsByPeriod(ctx, filter)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, movements, "Should return empty array")
}
