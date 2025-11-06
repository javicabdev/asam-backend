package services

import (
	"context"
	"errors"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/ports/output"
)

// MockTransaction simulates a database transaction for testing
type MockTransaction struct {
	CommitFunc   func() error
	RollbackFunc func() error
	Committed    bool
	RolledBack   bool
}

func (m *MockTransaction) Commit() error {
	m.Committed = true
	if m.CommitFunc != nil {
		return m.CommitFunc()
	}
	return nil
}

func (m *MockTransaction) Rollback() error {
	m.RolledBack = true
	if m.RollbackFunc != nil {
		return m.RollbackFunc()
	}
	return nil
}

// MockFamilyRepository simulates a family repository for testing
type MockFamilyRepository struct {
	CreateFunc              func(ctx context.Context, family *models.Family) error
	GetByIDFunc             func(ctx context.Context, id uint) (*models.Family, error)
	GetByOriginMemberIDFunc func(ctx context.Context, memberID uint) (*models.Family, error)
	BeginTransactionFunc    func(ctx context.Context) (output.Transaction, error)
	CreateWithTxFunc        func(ctx context.Context, tx output.Transaction, family *models.Family) error
	AddFamiliarWithTxFunc   func(ctx context.Context, tx output.Transaction, familyID uint, familiar *models.Familiar) error

	// Tracking
	CreatedFamilies    []*models.Family
	CreatedFamiliares  []*models.Familiar
	TransactionStarted bool
}

func (m *MockFamilyRepository) Create(ctx context.Context, family *models.Family) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, family)
	}
	m.CreatedFamilies = append(m.CreatedFamilies, family)
	return nil
}

func (m *MockFamilyRepository) GetByID(ctx context.Context, id uint) (*models.Family, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	for _, family := range m.CreatedFamilies {
		if family.ID == id {
			return family, nil
		}
	}
	return nil, nil
}

func (m *MockFamilyRepository) GetByOriginMemberID(ctx context.Context, memberID uint) (*models.Family, error) {
	if m.GetByOriginMemberIDFunc != nil {
		return m.GetByOriginMemberIDFunc(ctx, memberID)
	}
	for _, family := range m.CreatedFamilies {
		if family.MiembroOrigenID != nil && *family.MiembroOrigenID == memberID {
			return family, nil
		}
	}
	return nil, nil
}

func (m *MockFamilyRepository) BeginTransaction(ctx context.Context) (output.Transaction, error) {
	m.TransactionStarted = true
	if m.BeginTransactionFunc != nil {
		return m.BeginTransactionFunc(ctx)
	}
	return &MockTransaction{}, nil
}

func (m *MockFamilyRepository) CreateWithTx(ctx context.Context, tx output.Transaction, family *models.Family) error {
	if m.CreateWithTxFunc != nil {
		return m.CreateWithTxFunc(ctx, tx, family)
	}
	// Simulate auto-increment ID with safe conversion
	length := len(m.CreatedFamilies)
	if length < 0 {
		return errors.New("invalid CreatedFamilies length")
	}
	family.ID = uint(length) + 1
	m.CreatedFamilies = append(m.CreatedFamilies, family)
	return nil
}

func (m *MockFamilyRepository) AddFamiliarWithTx(ctx context.Context, tx output.Transaction, familyID uint, familiar *models.Familiar) error {
	if m.AddFamiliarWithTxFunc != nil {
		return m.AddFamiliarWithTxFunc(ctx, tx, familyID, familiar)
	}
	// Simulate auto-increment ID with safe conversion
	length := len(m.CreatedFamiliares)
	if length < 0 {
		return errors.New("invalid CreatedFamiliares length")
	}
	familiar.ID = uint(length) + 1
	familiar.FamiliaID = familyID
	m.CreatedFamiliares = append(m.CreatedFamiliares, familiar)
	return nil
}

// Unimplemented methods (not needed for our tests)
func (m *MockFamilyRepository) Update(ctx context.Context, family *models.Family) error {
	return errors.New("not implemented")
}

func (m *MockFamilyRepository) Delete(ctx context.Context, id uint) error {
	return errors.New("not implemented")
}

func (m *MockFamilyRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Family, error) {
	for _, family := range m.CreatedFamilies {
		if family.NumeroSocio == numeroSocio {
			return family, nil
		}
	}
	return nil, nil
}

func (m *MockFamilyRepository) List(ctx context.Context, page, pageSize int, searchTerm *string, orderBy string) ([]*models.Family, int, error) {
	return nil, 0, errors.New("not implemented")
}

func (m *MockFamilyRepository) AddFamiliar(ctx context.Context, familyID uint, familiar *models.Familiar) error {
	return errors.New("not implemented")
}

func (m *MockFamilyRepository) UpdateFamiliar(ctx context.Context, familiar *models.Familiar) error {
	return errors.New("not implemented")
}

func (m *MockFamilyRepository) RemoveFamiliar(ctx context.Context, familiarID uint) error {
	return errors.New("not implemented")
}

func (m *MockFamilyRepository) GetFamiliares(ctx context.Context, familyID uint) ([]*models.Familiar, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFamilyRepository) GetByIDWithTx(ctx context.Context, tx output.Transaction, id uint) (*models.Family, error) {
	return nil, errors.New("not implemented")
}

// MockMemberRepository simulates a member repository for testing
type MockMemberRepository struct {
	CreateWithTxFunc            func(ctx context.Context, tx output.Transaction, member *models.Member) error
	GetByIDWithTxFunc           func(ctx context.Context, tx output.Transaction, id uint) (*models.Member, error)
	GetByNumeroSocioWithTxFunc  func(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error)
	GetByIdentityCardWithTxFunc func(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error)

	// Tracking
	CreatedMembers []*models.Member
}

func (m *MockMemberRepository) CreateWithTx(ctx context.Context, tx output.Transaction, member *models.Member) error {
	if m.CreateWithTxFunc != nil {
		return m.CreateWithTxFunc(ctx, tx, member)
	}
	// Simulate auto-increment ID with safe conversion
	length := len(m.CreatedMembers)
	if length < 0 {
		return errors.New("invalid CreatedMembers length")
	}
	member.ID = uint(length) + 1
	m.CreatedMembers = append(m.CreatedMembers, member)
	return nil
}

func (m *MockMemberRepository) GetByIDWithTx(ctx context.Context, tx output.Transaction, id uint) (*models.Member, error) {
	if m.GetByIDWithTxFunc != nil {
		return m.GetByIDWithTxFunc(ctx, tx, id)
	}
	for _, member := range m.CreatedMembers {
		if member.ID == id {
			return member, nil
		}
	}
	return nil, nil
}

func (m *MockMemberRepository) GetByNumeroSocioWithTx(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error) {
	if m.GetByNumeroSocioWithTxFunc != nil {
		return m.GetByNumeroSocioWithTxFunc(ctx, tx, numeroSocio)
	}
	for _, member := range m.CreatedMembers {
		if member.MembershipNumber == numeroSocio {
			return member, nil
		}
	}
	return nil, nil
}

func (m *MockMemberRepository) GetByIdentityCardWithTx(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error) {
	if m.GetByIdentityCardWithTxFunc != nil {
		return m.GetByIdentityCardWithTxFunc(ctx, tx, identityCard)
	}
	for _, member := range m.CreatedMembers {
		if member.IdentityCard != nil && *member.IdentityCard == identityCard {
			return member, nil
		}
	}
	return nil, nil
}

// Unimplemented methods (not needed for our tests)
func (m *MockMemberRepository) Create(ctx context.Context, member *models.Member) error {
	return errors.New("not implemented")
}

func (m *MockMemberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) GetByIdentityCard(ctx context.Context, identityCard string) (*models.Member, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	return errors.New("not implemented")
}

func (m *MockMemberRepository) Delete(ctx context.Context, id uint) error {
	return errors.New("not implemented")
}

func (m *MockMemberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) GetLastMemberNumberByPrefix(ctx context.Context, prefix string) (string, error) {
	return "", errors.New("not implemented")
}

func (m *MockMemberRepository) SearchWithoutUser(ctx context.Context, criteria string) ([]models.Member, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) BeginTransaction(ctx context.Context) (output.Transaction, error) {
	return &MockTransaction{}, nil
}

// MockPaymentRepository simulates a payment repository for testing
type MockPaymentRepository struct {
	CreateWithTxFunc func(ctx context.Context, tx output.Transaction, payment *models.Payment) error
	// Tracking
	CreatedPayments []*models.Payment
}

func (m *MockPaymentRepository) CreateWithTx(ctx context.Context, tx output.Transaction, payment *models.Payment) error {
	if m.CreateWithTxFunc != nil {
		return m.CreateWithTxFunc(ctx, tx, payment)
	}
	// Simulate auto-increment ID
	length := len(m.CreatedPayments)
	if length < 0 {
		return errors.New("invalid CreatedPayments length")
	}
	payment.ID = uint(length) + 1
	m.CreatedPayments = append(m.CreatedPayments, payment)
	return nil
}

// Unimplemented methods (not needed for our tests)
func (m *MockPaymentRepository) Create(ctx context.Context, payment *models.Payment) error {
	return errors.New("not implemented")
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *models.Payment) error {
	return errors.New("not implemented")
}

func (m *MockPaymentRepository) Delete(ctx context.Context, id uint) error {
	return errors.New("not implemented")
}

func (m *MockPaymentRepository) GetByMember(ctx context.Context, memberID uint) ([]*models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) GetByFamily(ctx context.Context, familyID uint) ([]*models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) FindByMember(ctx context.Context, memberID uint, from, to time.Time) ([]models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) FindByFamily(ctx context.Context, familyID uint, from, to time.Time) ([]models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) GetDefaultersData(ctx context.Context) ([]output.DefaulterData, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) FindByID(ctx context.Context, id uint) (*models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) HasInitialPayment(ctx context.Context, memberID *uint, familyID *uint) (bool, error) {
	return false, errors.New("not implemented")
}

func (m *MockPaymentRepository) FindAll(ctx context.Context, filters *output.PaymentRepositoryFilters) ([]models.Payment, error) {
	return nil, errors.New("not implemented")
}

func (m *MockPaymentRepository) CountAll(ctx context.Context, filters *output.PaymentRepositoryFilters) (int64, error) {
	return 0, errors.New("not implemented")
}

// MockMembershipFeeRepository simulates a membership fee repository for testing
type MockMembershipFeeRepository struct {
	FindCurrentYearFunc func(ctx context.Context) (*models.MembershipFee, error)
	// Tracking
	Fees []*models.MembershipFee
}

func (m *MockMembershipFeeRepository) FindCurrentYear(ctx context.Context) (*models.MembershipFee, error) {
	if m.FindCurrentYearFunc != nil {
		return m.FindCurrentYearFunc(ctx)
	}
	// Return a default fee for testing
	return &models.MembershipFee{
		Year:           time.Now().Year(),
		BaseFeeAmount:  30.0,
		FamilyFeeExtra: 10.0,
	}, nil
}

// Unimplemented methods (not needed for our tests)
func (m *MockMembershipFeeRepository) Create(ctx context.Context, fee *models.MembershipFee) error {
	return errors.New("not implemented")
}

func (m *MockMembershipFeeRepository) FindByID(ctx context.Context, id uint) (*models.MembershipFee, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMembershipFeeRepository) FindByYear(ctx context.Context, year int) (*models.MembershipFee, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMembershipFeeRepository) FindPendingByMember(ctx context.Context, memberID uint) ([]models.MembershipFee, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMembershipFeeRepository) Update(ctx context.Context, fee *models.MembershipFee) error {
	return errors.New("not implemented")
}

func (m *MockMembershipFeeRepository) FindByYearWithTx(ctx context.Context, tx output.Transaction, year int) (*models.MembershipFee, error) {
	// Return a default fee for testing
	fee := &models.MembershipFee{
		Year:           year,
		BaseFeeAmount:  30.0,
		FamilyFeeExtra: 10.0,
	}
	fee.ID = 1
	return fee, nil
}

func (m *MockMembershipFeeRepository) CreateWithTx(ctx context.Context, tx output.Transaction, fee *models.MembershipFee) error {
	return errors.New("not implemented")
}
