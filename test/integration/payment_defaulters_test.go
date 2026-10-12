package integration

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetDefaultersData verifica que la query agregada de morosos filtra por los
// estados tal y como se guardan ("pending"/"paid"). Con literales en mayúsculas
// devolvía siempre una lista vacía.
func TestGetDefaultersData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	database, cleanup := setupTestDB(t)
	defer cleanup()

	paymentRepo := setupPaymentRepository(database)
	memberRepo := setupMemberRepository(database)
	membershipFeeRepo := setupMembershipFeeRepository(database)

	now := time.Now()
	overdueFee := createTestMembershipFee(t, ctx, membershipFeeRepo, now.Year()-1) // vencida el 31/12 del año pasado
	paidFee := createTestMembershipFee(t, ctx, membershipFeeRepo, now.Year()-2)
	futureFee := createTestMembershipFee(t, ctx, membershipFeeRepo, now.Year()+1) // aún no vencida

	// Moroso: cuota del año pasado pendiente y una cuota anterior pagada hace un mes
	defaulter := createDefaultersTestMember(t, ctx, memberRepo, "B90001")
	createDefaultersTestPayment(t, ctx, paymentRepo, defaulter.ID, overdueFee.ID, 50.0, models.PaymentStatusPending, nil)
	paidAt := now.AddDate(0, -1, 0)
	createDefaultersTestPayment(t, ctx, paymentRepo, defaulter.ID, paidFee.ID, 40.0, models.PaymentStatusPaid, &paidAt)

	// Al corriente: cuota vencida ya pagada y una pendiente que aún no ha vencido
	upToDate := createDefaultersTestMember(t, ctx, memberRepo, "B90002")
	createDefaultersTestPayment(t, ctx, paymentRepo, upToDate.ID, overdueFee.ID, 50.0, models.PaymentStatusPaid, &paidAt)
	createDefaultersTestPayment(t, ctx, paymentRepo, upToDate.ID, futureFee.ID, 50.0, models.PaymentStatusPending, nil)

	defaulters, err := paymentRepo.GetDefaultersData(ctx)
	require.NoError(t, err, "GetDefaultersData should not fail")
	require.Len(t, defaulters, 1, "only the member with an overdue pending fee is a defaulter")

	d := defaulters[0]
	assert.Equal(t, defaulter.ID, d.MemberID)
	assert.Equal(t, defaulter.MembershipNumber, d.NumeroSocio)
	assert.Equal(t, 1, d.OverdueCount, "the paid fee must not count as overdue")
	assert.Equal(t, overdueFee.DueDate.Format(time.DateOnly), d.OldestPendingDue.Format(time.DateOnly))
	assert.Equal(t, 40.0, d.TotalPaid, "total_paid should sum the member's paid payments")
	require.NotNil(t, d.LastPaymentDate, "last_payment_date should come from the paid payment")
	assert.WithinDuration(t, paidAt, *d.LastPaymentDate, time.Second)

	require.Len(t, d.PendingPayments, 1)
	assert.Equal(t, overdueFee.Year, d.PendingPayments[0].Year)
	assert.Len(t, d.PaymentHistory, 2, "history should include both the pending and the paid payment")
}

func createDefaultersTestMember(t *testing.T, ctx context.Context, memberRepo output.MemberRepository, number string) *models.Member {
	email := number + "@example.com"
	member := &models.Member{
		Name:             "Test Member",
		Surnames:         number,
		MembershipNumber: number,
		MembershipType:   models.TipoMembresiaPIndividual,
		Email:            &email,
		Address:          "Test Address 123",
		Postcode:         "08001",
		City:             "Barcelona",
		Province:         "Barcelona",
		Country:          "España",
		State:            models.EstadoActivo,
		RegistrationDate: time.Now().AddDate(-3, 0, 0),
	}

	err := memberRepo.Create(ctx, member)
	require.NoError(t, err, "Failed to create test member")
	return member
}

func createDefaultersTestPayment(
	t *testing.T,
	ctx context.Context,
	paymentRepo output.PaymentRepository,
	memberID, feeID uint,
	amount float64,
	status models.PaymentStatus,
	paymentDate *time.Time,
) {
	payment := &models.Payment{
		MemberID:        memberID,
		MembershipFeeID: &feeID,
		Amount:          amount,
		Status:          status,
		PaymentMethod:   "cash",
		PaymentDate:     paymentDate,
	}

	err := paymentRepo.Create(ctx, payment)
	require.NoError(t, err, "Failed to create test payment")
}
