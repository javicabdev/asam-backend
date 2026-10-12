package integration

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// TestCashFlowOperationTypeCheck verifies that cash_flows_operation_type_check
// (migration 000004) accepts every models.AllOperationTypes value and rejects
// anything else. If a type is added to the domain without a migration that
// updates the constraint, the app would fail to insert it in production: this
// test fails first.
func TestCashFlowOperationTypeCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	database, cleanup := setupTestDB(t)
	defer cleanup()

	var validated bool
	require.NoError(t, database.Raw(`
		SELECT convalidated FROM pg_constraint
		WHERE conname = 'cash_flows_operation_type_check' AND conrelid = 'cash_flows'::regclass`,
	).Row().Scan(&validated), "cash_flows_operation_type_check is missing: run the migrations")
	assert.True(t, validated, "cash_flows_operation_type_check is NOT VALID: this database had "+
		"invalid operation types when migrated; fix them and run "+
		"ALTER TABLE cash_flows VALIDATE CONSTRAINT cash_flows_operation_type_check")

	// Raw SQL on purpose: models.CashFlow.Validate would reject invalid types
	// before they ever reached the database.
	insert := func(opType string, amount float64) error {
		return database.Exec(
			"INSERT INTO cash_flows (operation_type, amount, date, detail) VALUES (?, ?, ?, ?)",
			opType, amount, time.Now(), "operation_type check test",
		).Error
	}

	for _, opType := range models.AllOperationTypes() {
		amount := 10.0
		if opType.IsExpense() {
			amount = -amount
		}
		assert.NoError(t, insert(string(opType), amount), "domain type %q must be accepted by the database", opType)
	}

	// The lowercase types are the ones the domain used before 2025-11-02.
	for _, opType := range []string{"ingreso_cuota", "gasto_corriente", "entrega_fondo", "otros_ingresos", "Ingreso_Cuota", ""} {
		err := insert(opType, 10)
		var pgErr *pgconn.PgError
		if assert.ErrorAs(t, err, &pgErr, "operation_type %q must be rejected", opType) {
			assert.Equal(t, "23514", pgErr.Code, "expected check_violation for %q", opType)
			assert.Equal(t, "cash_flows_operation_type_check", pgErr.ConstraintName)
		}
	}
}
