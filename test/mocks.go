package test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/logger/audit"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *MockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockMemberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Member), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
}

func (m *MockFamilyRepository) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	args := m.Called(ctx, page, pageSize, searchTerm, orderBy)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Family), args.Int(1), args.Error(2)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Familiar), args.Error(1)
}

func (m *MockMemberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	args := m.Called(ctx, numeroSocio)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *MockMemberRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
	args := m.Called(ctx, memberID, from, to)
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	args := m.Called(ctx, familyID, from, to)
	return args.Get(0).([]models.Payment), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MembershipFee), args.Error(1)
}

func (m *MockMembershipFeeRepository) FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error) {
	args := m.Called(ctx, memberID)
	return args.Get(0).([]models.MembershipFee), args.Error(1)
}

func (m *MockMembershipFeeRepository) FindByID(ctx context.Context, id uint) (*models.MembershipFee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MembershipFee), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CashFlow), args.Error(1)
}

func (m *MockCashFlowRepository) GetByPaymentID(ctx context.Context, paymentID uint) (*models.CashFlow, error) {
	args := m.Called(ctx, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CashFlow), args.Error(1)
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
	return args.Get(0).([]*models.CashFlow), args.Error(1)
}

func (m *MockCashFlowRepository) GetBalance(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *MockMemberService) GetMemberByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	args := m.Called(ctx, numeroSocio)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Convert the returned slice to []*models.Member
	members := args.Get(0).([]*models.Member)
	return members, args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
}

func (m *MockFamilyService) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	args := m.Called(ctx, numeroSocio)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Family), args.Error(1)
}

func (m *MockFamilyService) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	args := m.Called(ctx, page, pageSize, searchTerm, orderBy)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Family), args.Int(1), args.Error(2)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Familiar), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentService) GetMemberPayments(ctx context.Context, memberID uint) ([]*models.Payment, error) {
	args := m.Called(ctx, memberID)
	return args.Get(0).([]*models.Payment), args.Error(1)
}

func (m *MockPaymentService) GetFamilyPayments(ctx context.Context, familyID uint) ([]*models.Payment, error) {
	args := m.Called(ctx, familyID)
	return args.Get(0).([]*models.Payment), args.Error(1)
}

func (m *MockPaymentService) GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error {
	args := m.Called(ctx, year, month, baseAmount)
	return args.Error(0)
}

func (m *MockPaymentService) GetMembershipFee(ctx context.Context, year, month int) (*models.MembershipFee, error) {
	args := m.Called(ctx, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MembershipFee), args.Error(1)
}

func (m *MockPaymentService) UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error {
	args := m.Called(ctx, feeID, newAmount)
	return args.Error(0)
}

func (m *MockPaymentService) GetMemberStatement(ctx context.Context, memberID uint) (*input.AccountStatement, error) {
	args := m.Called(ctx, memberID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.AccountStatement), args.Error(1)
}

func (m *MockPaymentService) GetFamilyStatement(ctx context.Context, familyID uint) (*input.AccountStatement, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.AccountStatement), args.Error(1)
}

func (m *MockPaymentService) GetDefaulters(ctx context.Context) ([]input.AccountStatement, error) {
	args := m.Called(ctx)
	return args.Get(0).([]input.AccountStatement), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CashFlow), args.Error(1)
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
	return args.Get(0).([]*models.CashFlow), args.Error(1)
}

func (m *MockCashFlowService) GetCurrentBalance(ctx context.Context) (*input.BalanceReport, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.BalanceReport), args.Error(1)
}

func (m *MockCashFlowService) GetBalanceByPeriod(ctx context.Context, startDate, endDate time.Time) (*input.BalanceReport, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.BalanceReport), args.Error(1)
}

func (m *MockCashFlowService) ValidateBalance(ctx context.Context) (*input.BalanceValidation, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.BalanceValidation), args.Error(1)
}

func (m *MockCashFlowService) GetFinancialReport(ctx context.Context, reportType input.ReportType, period input.Period) (*input.FinancialReport, error) {
	args := m.Called(ctx, reportType, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.FinancialReport), args.Error(1)
}

func (m *MockCashFlowService) GetCashFlowTrends(ctx context.Context, period input.Period) (*input.TrendAnalysis, error) {
	args := m.Called(ctx, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.TrendAnalysis), args.Error(1)
}

func (m *MockCashFlowService) GetProjections(ctx context.Context, months int) (*input.FinancialProjection, error) {
	args := m.Called(ctx, months)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.FinancialProjection), args.Error(1)
}

func (m *MockCashFlowService) GetFinancialAlerts(ctx context.Context) ([]input.FinancialAlert, error) {
	args := m.Called(ctx)
	return args.Get(0).([]input.FinancialAlert), args.Error(1)
}

// MockAuthService es un mock de input.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, username, password string) (*input.TokenDetails, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.TokenDetails), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*input.TokenDetails, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*input.TokenDetails), args.Error(1)
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.User, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}
