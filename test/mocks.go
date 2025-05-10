// Package test proporciona utilidades y mocks para las pruebas unitarias y de integración
// del sistema, facilitando la simulación de servicios y repositorios.
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

// Create crea un nuevo miembro en el repositorio.
// Implementa la funcionalidad correspondiente del repositorio simulando la creación de miembros.
func (m *MockMemberRepository) Create(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

// GetByID obtiene un miembro por su ID.
// Simula la búsqueda en base de datos y permite controlar el comportamiento en pruebas.
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

// Update actualiza la información de un miembro existente.
// Permite simular éxito o fallos en la operación de actualización durante las pruebas.
func (m *MockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

// List obtiene una lista de miembros aplicando los filtros especificados.
// Permite personalizar los resultados que se devuelven según las necesidades de las pruebas.
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

// GetByNumeroSocio obtiene un miembro por su número de socio.
// Simula la consulta en la base de datos para pruebas de búsqueda por número de socio.
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

// Delete elimina un miembro del repositorio por su ID.
// Simula la operación de eliminación para pruebas.
func (m *MockMemberRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockLogger es un mock de Logger
type MockLogger struct{}

// Info registra mensajes en el nivel Info.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Info(string, ...zap.Field) {}

// Error registra mensajes en el nivel Error.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Error(string, ...zap.Field) {}

// Warn registra mensajes en el nivel Warn.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Warn(string, ...zap.Field) {}

// Debug registra mensajes en el nivel Debug.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Debug(string, ...zap.Field) {}

// Panic registra mensajes en el nivel Panic.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Panic(string, ...zap.Field) {}

// Fatal registra mensajes en el nivel Fatal.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Fatal(string, ...zap.Field) {}

// Sync sincroniza y vacía cualquier entrada de log almacenada en búfer.
// Implementa la interfaz de logger para pruebas sin realizar acciones reales.
func (m *MockLogger) Sync() error { return nil }

// MockAuditLogger es un mock de audit.Logger
type MockAuditLogger struct{}

// LogAction registra una acción en el log de auditoría.
// Simula el registro de acciones para pruebas sin realizar registro real.
func (m *MockAuditLogger) LogAction(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _ string) {
	// Simulación de logging
}

// LogChange registra un cambio en el log de auditoría.
// Simula el registro de cambios para pruebas sin realizar registro real.
func (m *MockAuditLogger) LogChange(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _, _ any, _ string) {
	// Simulación de logging
}

// LogError registra un error en el log de auditoría.
// Simula el registro de errores para pruebas sin realizar registro real.
func (m *MockAuditLogger) LogError(_ context.Context, _ audit.ActionType, _ audit.EntityType, _ string, _ string, _ error) {
	// Simulación de logging
}

// MockFamilyRepository es un mock de FamilyRepository
type MockFamilyRepository struct {
	mock.Mock
}

// Create crea una nueva familia en el repositorio.
// Simula la creación de una familia para pruebas sin usar base de datos real.
func (m *MockFamilyRepository) Create(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

// GetByID obtiene una familia por su ID.
// Simula la búsqueda en base de datos y permite controlar el comportamiento en pruebas.
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

// Update actualiza la información de una familia existente.
// Permite simular éxito o fallos en la operación de actualización durante las pruebas.
func (m *MockFamilyRepository) Update(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

// Delete elimina una familia del repositorio por su ID.
// Simula la operación de eliminación para pruebas.
func (m *MockFamilyRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByNumeroSocio obtiene una familia por su número de socio.
// Simula la consulta en la base de datos para pruebas de búsqueda por número de socio.
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

// List obtiene una lista de familias aplicando paginación, búsqueda y ordenación.
// Permite personalizar los resultados que se devuelven según las necesidades de las pruebas.
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

// AddFamiliar añade un familiar a una familia existente.
// Simula la adición de relaciones entre entidades para pruebas.
func (m *MockFamilyRepository) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	args := m.Called(ctx, familyID, familiar)
	return args.Error(0)
}

// UpdateFamiliar actualiza la información de un familiar existente.
// Simula la actualización de datos para pruebas.
func (m *MockFamilyRepository) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	args := m.Called(ctx, familiar)
	return args.Error(0)
}

// RemoveFamiliar elimina un familiar de una familia.
// Simula la eliminación de relaciones entre entidades para pruebas.
func (m *MockFamilyRepository) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	args := m.Called(ctx, familiarID)
	return args.Error(0)
}

// GetFamiliares obtiene todos los familiares asociados a una familia.
// Simula la recuperación de relaciones entre entidades para pruebas.
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

// Create registra un nuevo pago en el repositorio.
// Simula la creación de pagos para pruebas sin utilizar la base de datos real.
func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

// Update actualiza la información de un pago existente.
// Simula la actualización de pagos para pruebas.
func (m *MockPaymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

// Delete elimina un pago del repositorio por su ID.
// Simula la eliminación de pagos para pruebas.
func (m *MockPaymentRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindByID busca un pago por su ID.
// Simula la consulta de pagos individuales para pruebas.
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

// FindByMember busca pagos realizados por un miembro en un rango de fechas.
// Simula la consulta de pagos filtrados por miembro para pruebas.
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

// FindByFamily busca pagos realizados por una familia en un rango de fechas.
// Simula la consulta de pagos filtrados por familia para pruebas.
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

// Create crea una nueva cuota de membresía en el repositorio.
// Simula la creación de cuotas para pruebas sin usar la base de datos real.
func (m *MockMembershipFeeRepository) Create(ctx context.Context, fee *models.MembershipFee) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

// Update actualiza la información de una cuota de membresía existente.
// Simula la actualización de cuotas para pruebas.
func (m *MockMembershipFeeRepository) Update(ctx context.Context, fee *models.MembershipFee) error {
	args := m.Called(ctx, fee)
	return args.Error(0)
}

// FindByYearMonth busca una cuota de membresía por año y mes.
// Simula la consulta de cuotas filtradas por período para pruebas.
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

// FindPendingByMember busca cuotas de membresía pendientes de pago para un miembro.
// Simula la consulta de cuotas pendientes para pruebas.
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

// FindByID busca una cuota de membresía por su ID.
// Simula la consulta de cuotas individuales para pruebas.
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

// Create registra un nuevo movimiento de caja en el repositorio.
// Simula la creación de movimientos para pruebas sin usar la base de datos real.
func (m *MockCashFlowRepository) Create(ctx context.Context, cashFlow *models.CashFlow) error {
	args := m.Called(ctx, cashFlow)
	return args.Error(0)
}

// GetByID obtiene un movimiento de caja por su ID.
// Simula la consulta de movimientos individuales para pruebas.
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

// GetByPaymentID obtiene un movimiento de caja asociado a un pago específico.
// Simula la consulta de movimientos por relación con pagos para pruebas.
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

// Update actualiza la información de un movimiento de caja existente.
// Simula la actualización de movimientos para pruebas.
func (m *MockCashFlowRepository) Update(ctx context.Context, cashFlow *models.CashFlow) error {
	args := m.Called(ctx, cashFlow)
	return args.Error(0)
}

// Delete elimina un movimiento de caja del repositorio por su ID.
// Simula la eliminación de movimientos para pruebas.
func (m *MockCashFlowRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List obtiene una lista de movimientos de caja aplicando filtros.
// Permite personalizar los resultados devueltos según las necesidades de las pruebas.
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

// GetBalance obtiene el saldo actual de caja.
// Simula el cálculo del saldo para pruebas.
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
