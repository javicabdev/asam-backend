package services

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test/unit/testutils"
)

// TestGenerateAnnualFees_Success verifica que se generan cuotas exitosamente para todos los miembros activos
func TestGenerateAnnualFees_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	currentYear := time.Now().Year()

	member1 := &models.Member{
		ID:               1,
		MembershipNumber: "B00001",
		Name:             "Juan",
		Surnames:         "Pérez",
		MembershipType:   models.TipoMembresiaPIndividual,
		State:            models.EstadoActivo,
	}

	member2 := &models.Member{
		ID:               2,
		MembershipNumber: "A00001",
		Name:             "María",
		Surnames:         "García",
		MembershipType:   models.TipoMembresiaPFamiliar,
		State:            models.EstadoActivo,
	}

	mockMemberRepo := &MockMemberRepository{
		ActiveMembers: []*models.Member{member1, member2},
	}

	mockPaymentRepo := &MockPaymentRepository{
		CreatedPayments: []*models.Payment{},
	}

	mockFeeRepo := &MockMembershipFeeRepository{
		Fees: []*models.MembershipFee{},
	}

	mockFamilyRepo := &MockFamilyRepository{}
	mockCashFlowRepo := &MockCashFlowRepository{}

	paymentService := services.NewPaymentService(
		mockPaymentRepo,
		mockFeeRepo,
		mockMemberRepo,
		mockFamilyRepo,
		mockCashFlowRepo,
		nil, // feeCalculator not needed for this test
	)

	req := &input.GenerateAnnualFeesRequest{
		Year:           currentYear,
		BaseFeeAmount:  100.00,
		FamilyFeeExtra: 50.00,
	}

	// Act
	result, err := paymentService.GenerateAnnualFees(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, but got nil")
	}

	// Verificar que se creó la cuota de membresía
	if len(mockFeeRepo.Fees) != 1 {
		t.Errorf("Expected 1 membership fee created, but got %d", len(mockFeeRepo.Fees))
	}

	createdFee := mockFeeRepo.Fees[0]
	if createdFee.Year != currentYear {
		t.Errorf("Expected fee year %d, but got %d", currentYear, createdFee.Year)
	}

	if createdFee.BaseFeeAmount != 100.00 {
		t.Errorf("Expected base fee amount 100.00, but got %.2f", createdFee.BaseFeeAmount)
	}

	if createdFee.FamilyFeeExtra != 50.00 {
		t.Errorf("Expected family fee extra 50.00, but got %.2f", createdFee.FamilyFeeExtra)
	}

	// Verificar estadísticas de la respuesta
	if result.Year != currentYear {
		t.Errorf("Expected response year %d, but got %d", currentYear, result.Year)
	}

	if result.TotalMembers != 2 {
		t.Errorf("Expected 2 total members, but got %d", result.TotalMembers)
	}

	if result.PaymentsGenerated != 2 {
		t.Errorf("Expected 2 payments generated, but got %d", result.PaymentsGenerated)
	}

	if result.PaymentsExisting != 0 {
		t.Errorf("Expected 0 payments existing, but got %d", result.PaymentsExisting)
	}

	// Verificar que se crearon los pagos
	if len(mockPaymentRepo.CreatedPayments) != 2 {
		t.Fatalf("Expected 2 payments created, but got %d", len(mockPaymentRepo.CreatedPayments))
	}

	// Verificar pago del miembro individual
	payment1 := mockPaymentRepo.CreatedPayments[0]
	if payment1.MemberID != member1.ID {
		t.Errorf("Expected payment for member %d, but got %d", member1.ID, payment1.MemberID)
	}

	if payment1.Amount != 100.00 {
		t.Errorf("Expected amount 100.00 for individual member, but got %.2f", payment1.Amount)
	}

	if payment1.Status != models.PaymentStatusPending {
		t.Errorf("Expected status PENDING, but got %s", payment1.Status)
	}

	// Verificar pago del miembro familiar
	payment2 := mockPaymentRepo.CreatedPayments[1]
	if payment2.MemberID != member2.ID {
		t.Errorf("Expected payment for member %d, but got %d", member2.ID, payment2.MemberID)
	}

	if payment2.Amount != 150.00 {
		t.Errorf("Expected amount 150.00 for family member (100+50), but got %.2f", payment2.Amount)
	}

	// Verificar detalles
	if len(result.Details) != 2 {
		t.Fatalf("Expected 2 details, but got %d", len(result.Details))
	}

	for _, detail := range result.Details {
		if !detail.WasCreated {
			t.Errorf("Expected all details to have WasCreated=true, but got false for member %d", detail.MemberID)
		}

		if detail.Error != "" {
			t.Errorf("Expected no error in details, but got: %s", detail.Error)
		}
	}
}

// TestGenerateAnnualFees_FutureYear verifica que no se permiten años futuros
func TestGenerateAnnualFees_FutureYear(t *testing.T) {
	// Arrange
	ctx := context.Background()
	futureYear := time.Now().Year() + 1

	mockMemberRepo := &MockMemberRepository{}
	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	mockFamilyRepo := &MockFamilyRepository{}
	mockCashFlowRepo := &MockCashFlowRepository{}

	paymentService := services.NewPaymentService(
		mockPaymentRepo,
		mockFeeRepo,
		mockMemberRepo,
		mockFamilyRepo,
		mockCashFlowRepo,
		nil,
	)

	req := &input.GenerateAnnualFeesRequest{
		Year:           futureYear,
		BaseFeeAmount:  100.00,
		FamilyFeeExtra: 50.00,
	}

	// Act
	result, err := paymentService.GenerateAnnualFees(ctx, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for future year, but got nil")
	}

	if result != nil {
		t.Error("Expected nil result, but got a result object")
	}

	// Verificar que es un error de validación
	if appErr, ok := appErrors.AsAppError(err); ok {
		if appErr.Code != appErrors.ErrValidationFailed {
			t.Errorf("Expected validation error, got: %v", appErr.Code)
		}
	} else {
		t.Error("Expected AppError with validation details")
	}

	// Verificar que NO se crearon cuotas ni pagos
	if len(mockFeeRepo.Fees) > 0 {
		t.Errorf("Expected 0 fees created, but got %d", len(mockFeeRepo.Fees))
	}

	if len(mockPaymentRepo.CreatedPayments) > 0 {
		t.Errorf("Expected 0 payments created, but got %d", len(mockPaymentRepo.CreatedPayments))
	}
}

// TestGenerateAnnualFees_NegativeAmount verifica que no se permiten montos negativos
func TestGenerateAnnualFees_NegativeAmount(t *testing.T) {
	// Arrange
	ctx := context.Background()
	currentYear := time.Now().Year()

	mockMemberRepo := &MockMemberRepository{}
	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	mockFamilyRepo := &MockFamilyRepository{}
	mockCashFlowRepo := &MockCashFlowRepository{}

	paymentService := services.NewPaymentService(
		mockPaymentRepo,
		mockFeeRepo,
		mockMemberRepo,
		mockFamilyRepo,
		mockCashFlowRepo,
		nil,
	)

	req := &input.GenerateAnnualFeesRequest{
		Year:           currentYear,
		BaseFeeAmount:  -100.00,
		FamilyFeeExtra: 50.00,
	}

	// Act
	result, err := paymentService.GenerateAnnualFees(ctx, req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for negative amount, but got nil")
	}

	if result != nil {
		t.Error("Expected nil result, but got a result object")
	}

	// Verificar que NO se crearon cuotas ni pagos
	if len(mockFeeRepo.Fees) > 0 {
		t.Errorf("Expected 0 fees created, but got %d", len(mockFeeRepo.Fees))
	}
}

// TestGenerateAnnualFees_Idempotency verifica que no se crean duplicados
func TestGenerateAnnualFees_Idempotency(t *testing.T) {
	// Arrange
	ctx := context.Background()
	currentYear := time.Now().Year()

	member1 := &models.Member{
		ID:               1,
		MembershipNumber: "B00001",
		Name:             "Juan",
		Surnames:         "Pérez",
		MembershipType:   models.TipoMembresiaPIndividual,
		State:            models.EstadoActivo,
	}

	mockMemberRepo := &MockMemberRepository{
		ActiveMembers: []*models.Member{member1},
	}

	// Pre-crear una cuota de membresía
	existingFee := &models.MembershipFee{
		Year:           currentYear,
		BaseFeeAmount:  80.00, // Monto diferente
		FamilyFeeExtra: 40.00,
	}
	existingFee.ID = 1

	// Pre-crear un pago existente
	feeID := existingFee.ID
	existingPayment := &models.Payment{
		MemberID:        member1.ID,
		Amount:          80.00,
		Status:          models.PaymentStatusPending,
		MembershipFeeID: &feeID,
	}
	existingPayment.ID = 1

	mockPaymentRepo := &MockPaymentRepository{
		CreatedPayments: []*models.Payment{existingPayment},
	}

	mockFeeRepo := &MockMembershipFeeRepository{
		Fees: []*models.MembershipFee{existingFee},
	}

	mockFamilyRepo := &MockFamilyRepository{}
	mockCashFlowRepo := &MockCashFlowRepository{}

	paymentService := services.NewPaymentService(
		mockPaymentRepo,
		mockFeeRepo,
		mockMemberRepo,
		mockFamilyRepo,
		mockCashFlowRepo,
		nil,
	)

	req := &input.GenerateAnnualFeesRequest{
		Year:           currentYear,
		BaseFeeAmount:  100.00, // Nuevo monto
		FamilyFeeExtra: 50.00,
	}

	// Act
	result, err := paymentService.GenerateAnnualFees(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Verificar que se actualizó la cuota (no se creó una nueva)
	if len(mockFeeRepo.Fees) != 1 {
		t.Errorf("Expected 1 fee (updated, not created), but got %d", len(mockFeeRepo.Fees))
	}

	updatedFee := mockFeeRepo.Fees[0]
	if updatedFee.BaseFeeAmount != 100.00 {
		t.Errorf("Expected updated base fee amount 100.00, but got %.2f", updatedFee.BaseFeeAmount)
	}

	// Verificar que NO se creó un nuevo pago
	if result.PaymentsGenerated != 0 {
		t.Errorf("Expected 0 new payments (idempotency), but got %d", result.PaymentsGenerated)
	}

	if result.PaymentsExisting != 1 {
		t.Errorf("Expected 1 existing payment, but got %d", result.PaymentsExisting)
	}

	// Verificar que el total de pagos no cambió
	if len(mockPaymentRepo.CreatedPayments) != 1 {
		t.Errorf("Expected 1 total payment (no duplicates), but got %d", len(mockPaymentRepo.CreatedPayments))
	}

	// Verificar detalles
	if len(result.Details) != 1 {
		t.Fatalf("Expected 1 detail, but got %d", len(result.Details))
	}

	if result.Details[0].WasCreated {
		t.Error("Expected WasCreated=false for existing payment")
	}
}

// TestGenerateAnnualFees_NoActiveMembers verifica el caso sin miembros activos
func TestGenerateAnnualFees_NoActiveMembers(t *testing.T) {
	// Arrange
	ctx := context.Background()
	currentYear := time.Now().Year()

	mockMemberRepo := &MockMemberRepository{
		ActiveMembers: []*models.Member{}, // Sin miembros activos
	}

	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	mockFamilyRepo := &MockFamilyRepository{}
	mockCashFlowRepo := &MockCashFlowRepository{}

	paymentService := services.NewPaymentService(
		mockPaymentRepo,
		mockFeeRepo,
		mockMemberRepo,
		mockFamilyRepo,
		mockCashFlowRepo,
		nil,
	)

	req := &input.GenerateAnnualFeesRequest{
		Year:           currentYear,
		BaseFeeAmount:  100.00,
		FamilyFeeExtra: 50.00,
	}

	// Act
	result, err := paymentService.GenerateAnnualFees(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Verificar que se creó la cuota de membresía aunque no haya miembros
	if len(mockFeeRepo.Fees) != 1 {
		t.Errorf("Expected 1 membership fee created, but got %d", len(mockFeeRepo.Fees))
	}

	// Verificar estadísticas
	if result.TotalMembers != 0 {
		t.Errorf("Expected 0 total members, but got %d", result.TotalMembers)
	}

	if result.PaymentsGenerated != 0 {
		t.Errorf("Expected 0 payments generated, but got %d", result.PaymentsGenerated)
	}

	if len(mockPaymentRepo.CreatedPayments) != 0 {
		t.Errorf("Expected 0 payments created, but got %d", len(mockPaymentRepo.CreatedPayments))
	}
}

// MockCashFlowRepository es un mock simple para testing
type MockCashFlowRepository struct{}

func (m *MockCashFlowRepository) Create(ctx context.Context, cashFlow *models.CashFlow) error {
	return nil
}

func (m *MockCashFlowRepository) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) List(ctx context.Context, filters output.CashFlowFilter) ([]*models.CashFlow, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) Count(ctx context.Context, filters output.CashFlowFilter) (int64, error) {
	return 0, nil
}

func (m *MockCashFlowRepository) Update(ctx context.Context, cashFlow *models.CashFlow) error {
	return nil
}

func (m *MockCashFlowRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *MockCashFlowRepository) GetBalance(ctx context.Context, memberID *uint) (*output.CashFlowBalance, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) GetStats(ctx context.Context, from, to time.Time, memberID *uint) (*output.CashFlowStats, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) GetRunningBalance(ctx context.Context) (float64, error) {
	return 0, nil
}

func (m *MockCashFlowRepository) RecalculateRunningBalances(ctx context.Context) error {
	return nil
}

func (m *MockCashFlowRepository) ExistsByPaymentID(ctx context.Context, paymentID uint) (bool, error) {
	return false, nil
}

func (m *MockCashFlowRepository) GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) ListWithRunningBalance(ctx context.Context, filters output.CashFlowFilter) ([]*models.CashFlow, error) {
	return nil, nil
}

func (m *MockCashFlowRepository) UpdateCashFlowAndSyncPayment(ctx context.Context, cashFlow *models.CashFlow) error {
	return nil
}

func init() {
	// Ensure testutils package is used
	_ = testutils.TimePtr(time.Now())
}
