package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/test/seed/generators"
)

// TestSeedCashflowGenerator_ProducesValidDomainRows runs the fake-data cash flow
// generator against the real schema and checks every row against the domain
// model. The seed writes with raw SQL (bypassing CashFlow.Validate) and
// cash_flows.operation_type has no CHECK constraint, so this test is the only
// thing stopping the seed from storing values the app and GraphQL layer can't map.
func TestSeedCashflowGenerator_ProducesValidDomainRows(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	database, cleanup := setupTestDB(t)
	defer cleanup()

	sqlDB, err := database.DB()
	require.NoError(t, err)
	db := sqlx.NewDb(sqlDB, "pgx")

	const seed, numCashflows = 42, 500
	require.NoError(t, generators.NewMemberGenerator(db, seed).Generate(ctx, 20))
	require.NoError(t, generators.NewCashflowGenerator(db, seed).Generate(ctx, numCashflows))

	var rows []struct {
		ID            int       `db:"id"`
		MemberID      *int      `db:"member_id"`
		OperationType string    `db:"operation_type"`
		Amount        float64   `db:"amount"`
		Date          time.Time `db:"date"`
		Detail        string    `db:"detail"`
	}
	require.NoError(t, db.SelectContext(ctx, &rows,
		"SELECT id, member_id, operation_type, amount, date, detail FROM cash_flows"))
	require.Len(t, rows, numCashflows)

	var incomes, expenses, withMember int
	for _, r := range rows {
		opType := models.OperationType(r.OperationType)
		cf := models.CashFlow{OperationType: opType, Amount: r.Amount, Date: r.Date, Detail: r.Detail}
		// Validate covers IsValid() and the sign rules (income > 0, expense < 0).
		require.NoError(t, cf.Validate(), "cash_flows id=%d type=%q amount=%v", r.ID, r.OperationType, r.Amount)

		if opType == models.OperationTypeMembershipFee || opType == models.OperationTypeRepatriation {
			assert.NotNil(t, r.MemberID, "cash_flows id=%d type=%q must be linked to a member", r.ID, r.OperationType)
		}
		if opType.IsIncome() {
			incomes++
		} else {
			expenses++
		}
		if r.MemberID != nil {
			withMember++
		}
	}

	assert.Positive(t, incomes, "seed should generate income movements")
	assert.Positive(t, expenses, "seed should generate expense movements")
	assert.Positive(t, withMember, "seed should link some movements to active members")
}
