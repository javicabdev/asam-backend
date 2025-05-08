package test

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
)

// UintPtr devuelve un puntero a un uint
func UintPtr(u uint) *uint {
	return &u
}

// StringPtr devuelve un puntero a un string
func StringPtr(s string) *string {
	return &s
}

// TimePtr convierte un "time.Time" en un puntero a "time.Time"
func TimePtr(t time.Time) *time.Time {
	return &t
}

// MockMemberRepository es un mock de MemberRepository
type MockMemberRepository struct {
	mock.Mock
}

func (m *MockMemberRepository) Create(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)

	var member *models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		member, ok = ret0.(*models.Member)
		if !ok {
			return nil, err // Return zero value for member if type assertion fails
		}
	}
	return member, err
}

func (m *MockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	args := m.Called(ctx, filters)
	err := args.Error(1)

	var members []models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		members, ok = ret0.([]models.Member)
		if !ok {
			return nil, err // Return zero value for members if type assertion fails
		}
	}
	return members, err
}

func (m *MockMemberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	args := m.Called(ctx, numeroSocio)
	err := args.Error(1)

	var member *models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		member, ok = ret0.(*models.Member)
		if !ok {
			return nil, err
		}
	}
	return member, err
}

func (m *MockMemberRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockLogger es un mock de Logger
type MockLogger struct{}

func (m *MockLogger) Info(string, ...zap.Field)  {}
func (m *MockLogger) Error(string, ...zap.Field) {}
func (m *MockLogger) Warn(string, ...zap.Field)  {}
func (m *MockLogger) Debug(string, ...zap.Field) {}
func (m *MockLogger) Panic(string, ...zap.Field) {}
func (m *MockLogger) Fatal(string, ...zap.Field) {}
func (m *MockLogger) Sync() error                { return nil }

// MockAuditLogger es un mock de audit.Logger
type MockAuditLogger struct{}

func (m *MockAuditLogger) LogAction(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _ string) {
	// Simulación de logging
}

func (m *MockAuditLogger) LogChange(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _, _ any, _ string) {
	// Simulación de logging
}

func (m *MockAuditLogger) LogError(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _ string, _ error) {
	// Simulación de logging
}

// MockFamilyRepository es un mock de FamilyRepository
type MockFamilyRepository struct {
	mock.Mock
}

func (m *MockFamilyRepository) Create(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockFamilyRepository) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)

	var family *models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		family, ok = ret0.(*models.Family)
		if !ok {
			return nil, err
		}
	}
	return family, err
}

func (m *MockFamilyRepository) Update(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockFamilyRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFamilyRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	args := m.Called(ctx, numeroSocio)
	err := args.Error(1)

	var family *models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		family, ok = ret0.(*models.Family)
		if !ok {
			return nil, err
		}
	}
	return family, err
}

func (m *MockFamilyRepository) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	args := m.Called(ctx, page, pageSize, searchTerm, orderBy)
	err := args.Error(2) // Error is the third return value (index 2)

	var families []*models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		families, ok = ret0.([]*models.Family)
		if !ok {
			// If type assertion for families fails, return zero values for both families and count
			return nil, 0, err
		}
	}

	// Get the count, args.Int is safe and returns 0 if not set or wrong type
	count := args.Int(1)

	return families, count, err
}

func (m *MockFamilyRepository) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	args := m.Called(ctx, familyID, familiar)
	return args.Error(0)
}

func (m *MockFamilyRepository) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	args := m.Called(ctx, familiar)
	return args.Error(0)
}

func (m *MockFamilyRepository) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	args := m.Called(ctx, familiarID)
	return args.Error(0)
}

func (m *MockFamilyRepository) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	args := m.Called(ctx, familyID)
	err := args.Error(1) // Error is the second return value (index 1)

	var familiares []*models.Familiar
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		familiares, ok = ret0.([]*models.Familiar)
		if !ok {
			return nil, err // Return zero value for slice if type assertion fails
		}
	}
	return familiares, err
}

// MockPaymentRepository es un mock de PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(ctx context.Context, id uint) (*models.Payment, error) {
	args := m.Called(ctx, id)
	err := args.Error(1) // Error is the second return value (index 1)

	var payment *models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payment, ok = ret0.(*models.Payment)
		if !ok {
			return nil, err // Return zero value for pointer if type assertion fails
		}
	}
	return payment, err
}

func (m *MockPaymentRepository) FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
	args := m.Called(ctx, memberID, from, to)
	err := args.Error(1)

	var payments []models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payments, ok = ret0.([]models.Payment)
		if !ok {
			return nil, err
		}
	}
	return payments, err
}

func (m *MockPaymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	args := m.Called(ctx, familyID, from, to)
	err := args.Error(1)

	var payments []models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payments, ok = ret0.([]models.Payment)
		if !ok {
			return nil, err
		}
	}
	return payments, err
}

// MockMembershipFeeRepository es un mock de MembershipFeeRepository
type MockMembershipFeeRepository struct {
	mock.Mock
}

func (m *MockMembershipFeeRepository) Create(ctx context.Context, fee *models.MembershipFee) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

func (m *MockMembershipFeeRepository) Update(ctx context.Context, fee *models.MembershipFee) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

func (m *MockMembershipFeeRepository) FindByYearMonth(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	args := m.Called(ctx, year, month)
	err := args.Error(1)

	var fee *models.MembershipFee
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		fee, ok = ret0.(*models.MembershipFee)
		if !ok {
			return nil, err
		}
	}
	return fee, err
}

func (m *MockMembershipFeeRepository) FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error) {
	args := m.Called(ctx, memberID)
	err := args.Error(1)

	var fees []models.MembershipFee
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		fees, ok = ret0.([]models.MembershipFee)
		if !ok {
			return nil, err
		}
	}
	return fees, err
}

func (m *MockMembershipFeeRepository) FindByID(ctx context.Context, id uint) (*models.MembershipFee, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)

	var fee *models.MembershipFee
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		fee, ok = ret0.(*models.MembershipFee)
		if !ok {
			return nil, err
		}
	}
	return fee, err
}

// MockCashFlowRepository es un mock de CashFlowRepository
type MockCashFlowRepository struct {
	mock.Mock
}

func (m *MockCashFlowRepository) Create(ctx context.Context, cashFlow *models.CashFlow) error {
	args := m.Called(ctx, cashFlow)
	return args.Error(0)
}

func (m *MockCashFlowRepository) GetByID(ctx context.Context, id uint) (*models.CashFlow, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)

	var cashFlow *models.CashFlow
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		cashFlow, ok = ret0.(*models.CashFlow)
		if !ok {
			return nil, err
		}
	}
	return cashFlow, err
}

func (m *MockCashFlowRepository) GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error) {
	args := m.Called(ctx, paymentID)
	err := args.Error(1)

	var cashFlow *models.CashFlow
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		cashFlow, ok = ret0.(*models.CashFlow)
		if !ok {
			return nil, err
		}
	}
	return cashFlow, err
}

func (m *MockCashFlowRepository) Update(ctx context.Context, cashFlow *models.CashFlow) error {
	args := m.Called(ctx, cashFlow)
	return args.Error(0)
}

func (m *MockCashFlowRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCashFlowRepository) List(ctx context.Context, filter output.CashFlowFilter) ([]*models.CashFlow, error) {
	args := m.Called(ctx, filter)
	err := args.Error(1)

	var cashFlows []*models.CashFlow
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		cashFlows, ok = ret0.([]*models.CashFlow)
		if !ok {
			return nil, err
		}
	}
	return cashFlows, err
}

func (m *MockCashFlowRepository) GetBalance(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	// For simple types like float64, args.Float() is safer as it handles type assertion
	// and defaults to 0.0 if not configured or wrong type.
	balance := args.Get(0) // Get raw value
	if balance == nil {
		return 0.0, err // Return 0.0 if not configured
	}

	floatBalance, ok := balance.(float64)
	if !ok {
		// If it's not a float64, it's a mock setup issue. Return 0.0 and the error.
		return 0.0, err
	}
	return floatBalance, err
}

// MockMemberService es un mock de MemberService
type MockMemberService struct {
	mock.Mock
}

func (m *MockMemberService) CreateMember(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberService) GetMemberByID(ctx context.Context, id uint) (*models.Member, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)
	var member *models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		member, ok = ret0.(*models.Member)
		if !ok {
			return nil, err
		}
	}
	return member, err
}

func (m *MockMemberService) GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	args := m.Called(ctx, numeroSocio)
	err := args.Error(1)
	var member *models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		member, ok = ret0.(*models.Member)
		if !ok {
			return nil, err
		}
	}
	return member, err
}

func (m *MockMemberService) UpdateMember(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberService) DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error {
	args := m.Called(ctx, id, fechaBaja)
	return args.Error(0)
}

func (m *MockMemberService) ListMembers(ctx context.Context, filters input.MemberFilters) ([]*models.Member, error) {
	args := m.Called(ctx, filters)
	err := args.Error(1)
	var members []*models.Member
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		members, ok = ret0.([]*models.Member)
		if !ok {
			return nil, err
		}
	}
	return members, err
}

// MockFamilyService es un mock de FamilyService
type MockFamilyService struct {
	mock.Mock
}

func (m *MockFamilyService) Create(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockFamilyService) Update(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

func (m *MockFamilyService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFamilyService) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)
	var family *models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		family, ok = ret0.(*models.Family)
		if !ok {
			return nil, err
		}
	}
	return family, err
}

func (m *MockFamilyService) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	args := m.Called(ctx, numeroSocio)
	err := args.Error(1)
	var family *models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		family, ok = ret0.(*models.Family)
		if !ok {
			return nil, err
		}
	}
	return family, err
}

func (m *MockFamilyService) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	args := m.Called(ctx, page, pageSize, searchTerm, orderBy)
	err := args.Error(2)
	var families []*models.Family
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		families, ok = ret0.([]*models.Family)
		if !ok {
			return nil, 0, err
		}
	}
	count := args.Int(1)
	return families, count, err
}

func (m *MockFamilyService) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	args := m.Called(ctx, familyID, familiar)
	return args.Error(0)
}

func (m *MockFamilyService) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	args := m.Called(ctx, familiar)
	return args.Error(0)
}

func (m *MockFamilyService) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	args := m.Called(ctx, familiarID)
	return args.Error(0)
}

func (m *MockFamilyService) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	args := m.Called(ctx, familyID)
	err := args.Error(1)
	var familiares []*models.Familiar
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		familiares, ok = ret0.([]*models.Familiar)
		if !ok {
			return nil, err
		}
	}
	return familiares, err
}

// MockPaymentService es un mock de input.PaymentService
type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) RegisterPayment(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentService) CancelPayment(ctx context.Context, paymentID uint, reason string) error {
	args := m.Called(ctx, paymentID, reason)
	return args.Error(0)
}

func (m *MockPaymentService) GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error) {
	args := m.Called(ctx, paymentID)
	err := args.Error(1)
	var payment *models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payment, ok = ret0.(*models.Payment)
		if !ok {
			return nil, err
		}
	}
	return payment, err
}

func (m *MockPaymentService) GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error) {
	args := m.Called(ctx, memberID)
	err := args.Error(1)
	var payments []*models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payments, ok = ret0.([]*models.Payment)
		if !ok {
			return nil, err
		}
	}
	return payments, err
}

func (m *MockPaymentService) GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error) {
	args := m.Called(ctx, familyID)
	err := args.Error(1)
	var payments []*models.Payment
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		payments, ok = ret0.([]*models.Payment)
		if !ok {
			return nil, err
		}
	}
	return payments, err
}

func (m *MockPaymentService) GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error {
	args := m.Called(ctx, year, month, baseAmount)
	return args.Error(0)
}

func (m *MockPaymentService) GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	args := m.Called(ctx, year, month)
	err := args.Error(1)
	var fee *models.MembershipFee
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		fee, ok = ret0.(*models.MembershipFee)
		if !ok {
			return nil, err
		}
	}
	return fee, err
}

func (m *MockPaymentService) UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error {
	args := m.Called(ctx, feeID, newAmount)
	return args.Error(0)
}

func (m *MockPaymentService) GetMemberStatement(ctx context.Context, memberID uint) (*input.AccountStatement, error) {
	args := m.Called(ctx, memberID)
	err := args.Error(1)
	var statement *input.AccountStatement
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		statement, ok = ret0.(*input.AccountStatement)
		if !ok {
			return nil, err
		}
	}
	return statement, err
}

func (m *MockPaymentService) GetFamilyStatement(ctx context.Context, familyID uint) (*input.AccountStatement, error) {
	args := m.Called(ctx, familyID)
	err := args.Error(1)
	var statement *input.AccountStatement
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		statement, ok = ret0.(*input.AccountStatement)
		if !ok {
			return nil, err
		}
	}
	return statement, err
}

func (m *MockPaymentService) GetDefaulters(ctx context.Context) ([]input.AccountStatement, error) {
	args := m.Called(ctx)
	err := args.Error(1) // Assuming error is the second return value (index 1)
	var statements []input.AccountStatement
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		statements, ok = ret0.([]input.AccountStatement)
		if !ok {
			return nil, err
		}
	}
	return statements, err
}

func (m *MockPaymentService) SendPaymentReminder(ctx context.Context, memberID uint) error {
	args := m.Called(ctx, memberID)
	return args.Error(0)
}

func (m *MockPaymentService) SendPaymentConfirmation(ctx context.Context, paymentID uint) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

func (m *MockPaymentService) SendDefaulterNotification(ctx context.Context, memberID uint, days int) error {
	args := m.Called(ctx, memberID, days)
	return args.Error(0)
}

// MockCashFlowService es un mock de input.CashFlowService
type MockCashFlowService struct {
	mock.Mock
}

func (m *MockCashFlowService) RegisterMovement(ctx context.Context, movement *models.CashFlow) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockCashFlowService) GetMovement(ctx context.Context, id uint) (*models.CashFlow, error) {
	args := m.Called(ctx, id)
	err := args.Error(1)
	var movement *models.CashFlow
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		movement, ok = ret0.(*models.CashFlow)
		if !ok {
			return nil, err
		}
	}
	return movement, err
}

func (m *MockCashFlowService) UpdateMovement(ctx context.Context, movement *models.CashFlow) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

func (m *MockCashFlowService) DeleteMovement(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCashFlowService) GetMovementsByPeriod(ctx context.Context, filter input.CashFlowFilter) ([]*models.CashFlow, error) {
	args := m.Called(ctx, filter)
	err := args.Error(1) // Assuming error is the second return value (index 1)
	var movements []*models.CashFlow
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		movements, ok = ret0.([]*models.CashFlow)
		if !ok {
			return nil, err
		}
	}
	return movements, err
}

func (m *MockCashFlowService) GetCurrentBalance(ctx context.Context) (*input.BalanceReport, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	var report *input.BalanceReport
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		report, ok = ret0.(*input.BalanceReport)
		if !ok {
			return nil, err
		}
	}
	return report, err
}

func (m *MockCashFlowService) GetBalanceByPeriod(ctx context.Context, startDate, endDate time.Time) (*input.BalanceReport, error) {
	args := m.Called(ctx, startDate, endDate)
	err := args.Error(1)
	var report *input.BalanceReport
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		report, ok = ret0.(*input.BalanceReport)
		if !ok {
			return nil, err
		}
	}
	return report, err
}

func (m *MockCashFlowService) ValidateBalance(ctx context.Context) (*input.BalanceValidation, error) {
	args := m.Called(ctx)
	err := args.Error(1)
	var validation *input.BalanceValidation
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		validation, ok = ret0.(*input.BalanceValidation)
		if !ok {
			return nil, err
		}
	}
	return validation, err
}

func (m *MockCashFlowService) GetFinancialReport(ctx context.Context, reportType input.ReportType, period input.Period) (*input.FinancialReport, error) {
	args := m.Called(ctx, reportType, period)
	err := args.Error(1)
	var report *input.FinancialReport
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		report, ok = ret0.(*input.FinancialReport)
		if !ok {
			return nil, err
		}
	}
	return report, err
}

func (m *MockCashFlowService) GetCashFlowTrends(ctx context.Context, period input.Period) (*input.TrendAnalysis, error) {
	args := m.Called(ctx, period)
	err := args.Error(1)
	var analysis *input.TrendAnalysis
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		analysis, ok = ret0.(*input.TrendAnalysis)
		if !ok {
			return nil, err
		}
	}
	return analysis, err
}

func (m *MockCashFlowService) GetProjections(ctx context.Context, months int) (*input.FinancialProjection, error) {
	args := m.Called(ctx, months)
	err := args.Error(1)
	var projection *input.FinancialProjection
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		projection, ok = ret0.(*input.FinancialProjection)
		if !ok {
			return nil, err
		}
	}
	return projection, err
}

func (m *MockCashFlowService) GetFinancialAlerts(ctx context.Context) ([]input.FinancialAlert, error) {
	args := m.Called(ctx)
	err := args.Error(1) // Assuming error is the second return value (index 1)
	var alerts []input.FinancialAlert
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		alerts, ok = ret0.([]input.FinancialAlert)
		if !ok {
			return nil, err
		}
	}
	return alerts, err
}

// MockAuthService es un mock de input.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*input.TokenDetails, error) {
	args := m.Called(ctx, username, password)
	err := args.Error(1)
	var details *input.TokenDetails
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		details, ok = ret0.(*input.TokenDetails)
		if !ok {
			return nil, err
		}
	}
	return details, err
}

func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*input.TokenDetails, error) {
	args := m.Called(ctx, refreshToken)
	err := args.Error(1)
	var details *input.TokenDetails
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		details, ok = ret0.(*input.TokenDetails)
		if !ok {
			return nil, err
		}
	}
	return details, err
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	err := args.Error(1)
	var user *models.User
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		user, ok = ret0.(*models.User)
		if !ok {
			return nil, err
		}
	}
	return user, err
}
