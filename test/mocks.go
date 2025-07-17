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

// CreateMember creates a new member in the system.
// It simulates the creation process without accessing real database.
func (m *MockMemberService) CreateMember(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

// GetMemberByID retrieves a member by their unique identifier.
// Returns the member object or an error if not found or other issues occur.
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

// GetMemberByNumeroSocio retrieves a member by their membership number.
// This method simulates database lookup without actual database access.
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

// UpdateMember updates an existing member's information.
// Returns an error if the update operation fails.
func (m *MockMemberService) UpdateMember(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

// DeactivateMember sets a member as inactive with an optional deactivation date.
// Returns an error if the deactivation process fails.
func (m *MockMemberService) DeactivateMember(ctx context.Context, id uint, fechaBaja *time.Time) error {
	args := m.Called(ctx, id, fechaBaja)
	return args.Error(0)
}

// ListMembers retrieves a list of members based on the provided filters.
// Returns the filtered list and an error if the operation fails.
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

// Create creates a new family record in the system.
// Returns an error if the operation fails.
func (m *MockFamilyService) Create(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

// Update updates an existing family's information.
// Returns an error if the update operation fails.
func (m *MockFamilyService) Update(ctx context.Context, family *models.Family) error {
	args := m.Called(ctx, family)
	return args.Error(0)
}

// Delete removes a family record by its ID.
// Returns an error if the deletion fails.
func (m *MockFamilyService) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByID retrieves a family record by its unique identifier.
// Returns the family object or an error if not found or other issues occur.
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

// GetByNumeroSocio retrieves a family by their membership number.
// Returns the family object or an error if not found or other issues occur.
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

// List retrieves families with pagination, search and sorting options.
// Returns the list of families, the total count, and an error if the operation fails.
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

// AddFamiliar adds a new family member to an existing family.
// Returns an error if the operation fails.
func (m *MockFamilyService) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	args := m.Called(ctx, familyID, familiar)
	return args.Error(0)
}

// UpdateFamiliar updates information of an existing family member.
// Returns an error if the update operation fails.
func (m *MockFamilyService) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	args := m.Called(ctx, familiar)
	return args.Error(0)
}

// RemoveFamiliar removes a family member from a family by their ID.
// Returns an error if the removal fails.
func (m *MockFamilyService) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	args := m.Called(ctx, familiarID)
	return args.Error(0)
}

// GetFamiliares retrieves all family members for a specific family.
// Returns the list of family members and an error if the retrieval fails.
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

// RegisterPayment records a new payment in the system.
// Returns an error if the registration fails.
func (m *MockPaymentService) RegisterPayment(ctx context.Context, payment *models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

// CancelPayment cancels an existing payment with a specified reason.
// Returns an error if the cancellation process fails.
func (m *MockPaymentService) CancelPayment(ctx context.Context, paymentID uint, reason string) error {
	args := m.Called(ctx, paymentID, reason)
	return args.Error(0)
}

// GetPayment retrieves a payment record by its ID.
// Returns the payment object or an error if not found or other issues occur.
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

// GetMemberPayments retrieves all payments made by a specific member.
// Returns the list of payments and an error if the retrieval fails.
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

// GetFamilyPayments retrieves all payments made by a specific family.
// Returns the list of payments and an error if the retrieval fails.
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

// GenerateMonthlyFees creates monthly membership fees for all active members.
// Returns an error if the generation process fails.
func (m *MockPaymentService) GenerateMonthlyFees(ctx context.Context, year, month int, baseAmount float64) error {
	args := m.Called(ctx, year, month, baseAmount)
	return args.Error(0)
}

// GetMembershipFee retrieves the membership fee for a specific month and year.
// Returns the fee object or an error if not found or other issues occur.
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

// UpdateFeeAmount changes the amount of a specific membership fee.
// Returns an error if the update operation fails.
func (m *MockPaymentService) UpdateFeeAmount(ctx context.Context, feeID uint, newAmount float64) error {
	args := m.Called(ctx, feeID, newAmount)
	return args.Error(0)
}

// GetMemberStatement generates an account statement for a specific member.
// Returns the statement with balance and payment history or an error if the process fails.
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

// GetFamilyStatement generates an account statement for a specific family.
// Returns the statement with balance and payment history or an error if the process fails.
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

// GetDefaulters retrieves a list of members who are behind on payments.
// Returns the list of defaulters with their account statements or an error if the retrieval fails.
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

// SendPaymentReminder sends a payment reminder notification to a specific member.
// Returns an error if the notification sending process fails.
func (m *MockPaymentService) SendPaymentReminder(ctx context.Context, memberID uint) error {
	args := m.Called(ctx, memberID)
	return args.Error(0)
}

// SendPaymentConfirmation sends a payment confirmation notification for a specific payment.
// Returns an error if the notification sending process fails.
func (m *MockPaymentService) SendPaymentConfirmation(ctx context.Context, paymentID uint) error {
	args := m.Called(ctx, paymentID)
	return args.Error(0)
}

// SendDefaulterNotification sends a notification to a member who is behind on payments.
// The days parameter specifies how many days the member is late on payment.
// Returns an error if the notification sending process fails.
func (m *MockPaymentService) SendDefaulterNotification(ctx context.Context, memberID uint, days int) error {
	args := m.Called(ctx, memberID, days)
	return args.Error(0)
}

// MockCashFlowService es un mock de input.CashFlowService
type MockCashFlowService struct {
	mock.Mock
}

// RegisterMovement records a new cash flow movement in the system.
// Returns an error if the registration fails.
func (m *MockCashFlowService) RegisterMovement(ctx context.Context, movement *models.CashFlow) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

// GetMovement retrieves a cash flow movement by its ID.
// Returns the movement object or an error if not found or other issues occur.
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

// UpdateMovement updates an existing cash flow movement.
// Returns an error if the update operation fails.
func (m *MockCashFlowService) UpdateMovement(ctx context.Context, movement *models.CashFlow) error {
	args := m.Called(ctx, movement)
	return args.Error(0)
}

// DeleteMovement removes a cash flow movement by its ID.
// Returns an error if the deletion fails.
func (m *MockCashFlowService) DeleteMovement(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetMovementsByPeriod retrieves cash flow movements for a specific time period with filtering options.
// Returns the list of movements and an error if the retrieval fails.
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

// GetCurrentBalance calculates and returns the current balance of all accounts.
// Returns a balance report or an error if the calculation fails.
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

// GetBalanceByPeriod calculates and returns the balance for a specific time period.
// Returns a balance report or an error if the calculation fails.
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

// ValidateBalance performs a validation of the current balance against expected values.
// Returns a validation report or an error if the validation process fails.
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

// GetFinancialReport obtiene un informe financiero basado en el tipo de reporte y periodo especificados.
// Retorna el informe con datos financieros o un error si la operación falla.
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

// GetCashFlowTrends analiza tendencias en el flujo de caja para un periodo determinado.
// Retorna un análisis de tendencias o un error si la operación falla.
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

// GetProjections genera proyecciones financieras para el número de meses especificado.
// Retorna las proyecciones estimadas o un error si la operación falla.
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

// GetFinancialAlerts obtiene alertas financieras basadas en reglas de negocio predefinidas.
// Retorna una lista de alertas relevantes o un error si la operación falla.
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

// MockUserService es un mock de input.UserService
type MockUserService struct {
	mock.Mock
}

// CreateUser creates a new user with the given details.
// Returns the created user or an error if the creation fails.
func (m *MockUserService) CreateUser(ctx context.Context, username, password string, role models.Role) (*models.User, error) {
	args := m.Called(ctx, username, password, role)
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

// UpdateUser updates an existing user's details.
// Returns the updated user or an error if the update fails.
func (m *MockUserService) UpdateUser(ctx context.Context, id uint, updates map[string]interface{}) (*models.User, error) {
	args := m.Called(ctx, id, updates)
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

// DeleteUser deletes a user by ID (soft delete).
// Returns an error if the deletion fails.
func (m *MockUserService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetUser retrieves a user by ID.
// Returns the user or an error if not found or other issues occur.
func (m *MockUserService) GetUser(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
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

// GetUserByEmail retrieves a user by email address.
// Returns the user or an error if not found or other issues occur.
func (m *MockUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
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

// ListUsers retrieves a paginated list of users.
// Returns the list of users or an error if the retrieval fails.
func (m *MockUserService) ListUsers(ctx context.Context, page, pageSize int) ([]*models.User, error) {
	args := m.Called(ctx, page, pageSize)
	err := args.Error(1)
	var users []*models.User
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		users, ok = ret0.([]*models.User)
		if !ok {
			return nil, err
		}
	}
	return users, err
}

// ChangePassword changes a user's password.
// Returns an error if the password change fails.
func (m *MockUserService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	args := m.Called(ctx, userID, currentPassword, newPassword)
	return args.Error(0)
}

// ResetPassword resets a user's password (admin function).
// Returns an error if the password reset fails.
func (m *MockUserService) ResetPassword(ctx context.Context, userID uint, newPassword string) error {
	args := m.Called(ctx, userID, newPassword)
	return args.Error(0)
}

// SendVerificationEmail sends a verification email to the user.
// Returns an error if the email sending fails.
func (m *MockUserService) SendVerificationEmail(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// VerifyEmail verifies a user's email using the provided token.
// Returns an error if the verification fails.
func (m *MockUserService) VerifyEmail(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// RequestPasswordReset initiates a password reset request for the given email.
// Returns an error if the request fails.
func (m *MockUserService) RequestPasswordReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

// ResetPasswordWithToken resets a user's password using a reset token.
// Returns an error if the reset fails.
func (m *MockUserService) ResetPasswordWithToken(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

// ResendVerificationEmail resends the verification email to the user.
// Returns an error if the email sending fails.
func (m *MockUserService) ResendVerificationEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

// MockAuthService es un mock de input.AuthService
type MockAuthService struct {
	mock.Mock
}

// Login autentica a un usuario con las credenciales proporcionadas.
// Retorna detalles del token si la autenticación es exitosa o un error si falla.
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

// Logout cierra la sesión del usuario invalidando el token de acceso.
// Retorna un error si la operación falla.
func (m *MockAuthService) Logout(ctx context.Context, accessToken string) error {
	args := m.Called(ctx, accessToken)
	return args.Error(0)
}

// RefreshToken renueva un token de acceso usando un token de refresco válido.
// Retorna nuevos detalles de token o un error si la operación falla.
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

// ValidateToken verifica la validez de un token de acceso y retorna el usuario asociado.
// Retorna el usuario si el token es válido o un error si la validación falla.
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

// ResetPasswordWithToken restablece la contraseña de un usuario usando un token de restablecimiento válido.
// Retorna un error si la operación falla.
func (m *MockAuthService) ResetPasswordWithToken(ctx context.Context, token string, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

// MockEmailVerificationService es un mock de input.EmailVerificationService
type MockEmailVerificationService struct {
	mock.Mock
}

// GenerateVerificationToken genera un token de verificación de email
func (m *MockEmailVerificationService) GenerateVerificationToken(ctx context.Context, userID uint) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

// GeneratePasswordResetToken genera un token de restablecimiento de contraseña
func (m *MockEmailVerificationService) GeneratePasswordResetToken(ctx context.Context, userID uint) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

// VerifyEmailToken verifica un token de email
func (m *MockEmailVerificationService) VerifyEmailToken(ctx context.Context, token string) (*models.User, error) {
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

// VerifyPasswordResetToken verifica un token de restablecimiento de contraseña
func (m *MockEmailVerificationService) VerifyPasswordResetToken(ctx context.Context, token string) (*models.VerificationToken, error) {
	args := m.Called(ctx, token)
	err := args.Error(1)
	var verToken *models.VerificationToken
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		verToken, ok = ret0.(*models.VerificationToken)
		if !ok {
			return nil, err
		}
	}
	return verToken, err
}

// SendVerificationEmailToUser envía un email de verificación a un usuario
func (m *MockEmailVerificationService) SendVerificationEmailToUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// SendPasswordResetEmailToUser envía un email de restablecimiento de contraseña a un usuario
func (m *MockEmailVerificationService) SendPasswordResetEmailToUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// CleanupExpiredTokens limpia los tokens expirados
func (m *MockEmailVerificationService) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockEmailNotificationService es un mock de input.EmailNotificationService
type MockEmailNotificationService struct {
	mock.Mock
}

// SendVerificationEmail envía un email de verificación
func (m *MockEmailNotificationService) SendVerificationEmail(ctx context.Context, user *models.User, verificationURL string) error {
	args := m.Called(ctx, user, verificationURL)
	return args.Error(0)
}

// SendPasswordResetEmail envía un email de restablecimiento de contraseña
func (m *MockEmailNotificationService) SendPasswordResetEmail(ctx context.Context, user *models.User, resetURL string) error {
	args := m.Called(ctx, user, resetURL)
	return args.Error(0)
}

// SendWelcomeEmail envía un email de bienvenida
func (m *MockEmailNotificationService) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// SendPasswordChangedEmail envía un email de notificación de cambio de contraseña
func (m *MockEmailNotificationService) SendPasswordChangedEmail(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// MockVerificationTokenRepository es un mock de output.VerificationTokenRepository
type MockVerificationTokenRepository struct {
	mock.Mock
}

// Create crea un nuevo token de verificación
func (m *MockVerificationTokenRepository) Create(ctx context.Context, token *models.VerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// GetByToken obtiene un token por su valor
func (m *MockVerificationTokenRepository) GetByToken(ctx context.Context, token string) (*models.VerificationToken, error) {
	args := m.Called(ctx, token)
	err := args.Error(1)
	var verToken *models.VerificationToken
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		verToken, ok = ret0.(*models.VerificationToken)
		if !ok {
			return nil, err
		}
	}
	return verToken, err
}

// Update actualiza un token existente
func (m *MockVerificationTokenRepository) Update(ctx context.Context, token *models.VerificationToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// DeleteExpired elimina todos los tokens expirados
func (m *MockVerificationTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// InvalidateUserTokens invalida todos los tokens de un usuario de un tipo específico
func (m *MockVerificationTokenRepository) InvalidateUserTokens(ctx context.Context, userID uint, tokenType string) error {
	args := m.Called(ctx, userID, tokenType)
	return args.Error(0)
}

// CountActiveTokensByUser cuenta los tokens activos de un usuario
func (m *MockVerificationTokenRepository) CountActiveTokensByUser(ctx context.Context, userID uint, tokenType string) (int64, error) {
	args := m.Called(ctx, userID, tokenType)
	err := args.Error(1)

	// Get the count safely
	ret0 := args.Get(0)
	if ret0 == nil {
		return 0, err
	}

	count, ok := ret0.(int64)
	if !ok {
		// If type assertion fails, return 0 and the error
		return 0, err
	}

	return count, err
}

// Delete elimina un token por su ID
func (m *MockVerificationTokenRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// GetByUserIDAndType obtiene tokens por ID de usuario y tipo
func (m *MockVerificationTokenRepository) GetByUserIDAndType(ctx context.Context, userID uint, tokenType string) ([]*models.VerificationToken, error) {
	args := m.Called(ctx, userID, tokenType)
	err := args.Error(1)
	var tokens []*models.VerificationToken
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		tokens, ok = ret0.([]*models.VerificationToken)
		if !ok {
			return nil, err
		}
	}
	return tokens, err
}

// MockUserRepository es un mock de output.UserRepository
type MockUserRepository struct {
	mock.Mock
}

// Create crea un nuevo usuario
func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Update actualiza un usuario existente
func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

// Delete elimina un usuario por su ID
func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindByID busca un usuario por su ID
func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
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

// FindByEmail busca un usuario por su email
func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
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

// List lista usuarios con paginación
func (m *MockUserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, offset, limit)
	err := args.Error(1)
	var users []*models.User
	ret0 := args.Get(0)
	if ret0 != nil {
		var ok bool
		users, ok = ret0.([]*models.User)
		if !ok {
			return nil, err
		}
	}
	return users, err
}
