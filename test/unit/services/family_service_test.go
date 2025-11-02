package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test/unit/testutils"
)

// TestCreateFamilyAtomic_RollbackOnInvalidEmail verifies that if email validation fails,
// NO member, family, or familiar is created (complete rollback)
func TestCreateFamilyAtomic_RollbackOnInvalidEmail(t *testing.T) {
	// Arrange
	mockFamilyRepo := &MockFamilyRepository{}
	mockMemberRepo := &MockMemberRepository{}
	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}

	familyService := services.NewFamilyService(mockFamilyRepo, mockMemberRepo, mockPaymentRepo, mockFeeRepo)

	req := &input.CreateFamilyAtomicRequest{
		Family: &models.Family{
			NumeroSocio:              "A00123",
			EsposoNombre:             "Juan",
			EsposoApellidos:          "Pérez García",
			EsposoDocumentoIdentidad: "12345678Z",     // Valid DNI
			EsposoCorreoElectronico:  "invalid_email", // ❌ Email sin @ ni dominio
			EsposaNombre:             "María",
			EsposaApellidos:          "López Martínez",
		},
		CreateMemberIfNotExists: true,
		Familiares: []*models.Familiar{
			{
				Nombre:          "Pedro",
				Apellidos:       "Pérez López",
				FechaNacimiento: testutils.TimePtr(time.Now().AddDate(-10, 0, 0)),
				Parentesco:      "Hijo",
			},
		},
	}

	// Act
	family, err := familyService.CreateFamilyAtomic(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected validation error for invalid email, but got nil")
	}

	// Verify error contains email validation message
	if appErr, ok := appErrors.AsAppError(err); ok {
		if appErr.Code != appErrors.ErrValidationFailed {
			t.Errorf("Expected validation error, got: %v", appErr.Code)
		}
	} else {
		t.Error("Expected AppError with validation details")
	}

	if family != nil {
		t.Error("Expected nil family, but got a family object")
	}

	// CRITICAL: Verify that transaction was NOT started (validation happens before)
	if mockFamilyRepo.TransactionStarted {
		t.Error("Transaction should NOT have started due to validation failure")
	}

	// CRITICAL: Verify NO member was created
	if len(mockMemberRepo.CreatedMembers) > 0 {
		t.Errorf("Expected 0 members created, but found %d members", len(mockMemberRepo.CreatedMembers))
	}

	// CRITICAL: Verify NO family was created
	if len(mockFamilyRepo.CreatedFamilies) > 0 {
		t.Errorf("Expected 0 families created, but found %d families", len(mockFamilyRepo.CreatedFamilies))
	}

	// CRITICAL: Verify NO familiares were created
	if len(mockFamilyRepo.CreatedFamiliares) > 0 {
		t.Errorf("Expected 0 familiares created, but found %d familiares", len(mockFamilyRepo.CreatedFamiliares))
	}
}

// TestCreateFamilyAtomic_RollbackOnFamilyCreationError verifies that if family creation fails,
// the member creation is rolled back
func TestCreateFamilyAtomic_RollbackOnFamilyCreationError(t *testing.T) {
	// Arrange
	mockTx := &MockTransaction{}
	memberCreated := false

	mockFamilyRepo := &MockFamilyRepository{
		BeginTransactionFunc: func(ctx context.Context) (output.Transaction, error) {
			return mockTx, nil
		},
		CreateWithTxFunc: func(ctx context.Context, tx output.Transaction, family *models.Family) error {
			// Simulate error when creating family
			return errors.New("database error creating family")
		},
	}
	mockMemberRepo := &MockMemberRepository{
		CreateWithTxFunc: func(ctx context.Context, tx output.Transaction, member *models.Member) error {
			memberCreated = true
			member.ID = 1
			return nil
		},
		GetByNumeroSocioWithTxFunc: func(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error) {
			return nil, nil // No existing member
		},
		GetByIdentityCardWithTxFunc: func(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error) {
			return nil, nil // No duplicate DNI
		},
	}

	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	familyService := services.NewFamilyService(mockFamilyRepo, mockMemberRepo, mockPaymentRepo, mockFeeRepo)

	req := &input.CreateFamilyAtomicRequest{
		Family: &models.Family{
			NumeroSocio:              "A00124",
			EsposoNombre:             "Carlos",
			EsposoApellidos:          "González Ruiz",
			EsposoDocumentoIdentidad: "12345678Z", // Valid DNI
			EsposoCorreoElectronico:  "carlos@example.com",
			EsposaNombre:             "Ana",
			EsposaApellidos:          "Martín Sánchez",
		},
		CreateMemberIfNotExists: true,
	}

	// Act
	family, err := familyService.CreateFamilyAtomic(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error when creating family, but got nil")
	}

	if family != nil {
		t.Error("Expected nil family, but got a family object")
	}

	// CRITICAL: Verify transaction was rolled back
	if !mockTx.RolledBack {
		t.Error("Expected transaction to be rolled back after family creation error")
	}

	if mockTx.Committed {
		t.Error("Transaction should NOT have been committed after error")
	}

	// Verify member creation was attempted
	if !memberCreated {
		t.Error("Expected member creation to be attempted before family error")
	}

	// Verify family was NOT successfully created
	if len(mockFamilyRepo.CreatedFamilies) > 0 {
		t.Errorf("Expected 0 families created due to error, but found %d", len(mockFamilyRepo.CreatedFamilies))
	}
}

// TestCreateFamilyAtomic_RollbackOnFamiliarCreationError verifies that if familiar creation fails,
// both member and family creations are rolled back
func TestCreateFamilyAtomic_RollbackOnFamiliarCreationError(t *testing.T) {
	// Arrange
	mockTx := &MockTransaction{}
	memberCreated := false
	familyCreated := false

	mockFamilyRepo := &MockFamilyRepository{
		BeginTransactionFunc: func(ctx context.Context) (output.Transaction, error) {
			return mockTx, nil
		},
		CreateWithTxFunc: func(ctx context.Context, tx output.Transaction, family *models.Family) error {
			familyCreated = true
			family.ID = 1
			return nil
		},
		AddFamiliarWithTxFunc: func(ctx context.Context, tx output.Transaction, familyID uint, familiar *models.Familiar) error {
			// Simulate error when adding familiar
			return errors.New("database error creating familiar")
		},
	}
	mockMemberRepo := &MockMemberRepository{
		CreateWithTxFunc: func(ctx context.Context, tx output.Transaction, member *models.Member) error {
			memberCreated = true
			member.ID = 1
			return nil
		},
		GetByNumeroSocioWithTxFunc: func(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error) {
			return nil, nil
		},
		GetByIdentityCardWithTxFunc: func(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error) {
			return nil, nil
		},
	}

	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	familyService := services.NewFamilyService(mockFamilyRepo, mockMemberRepo, mockPaymentRepo, mockFeeRepo)

	req := &input.CreateFamilyAtomicRequest{
		Family: &models.Family{
			NumeroSocio:              "A00125",
			EsposoNombre:             "Luis",
			EsposoApellidos:          "Fernández Díaz",
			EsposoDocumentoIdentidad: "12345678Z", // Valid DNI
			EsposoCorreoElectronico:  "luis@example.com",
		},
		CreateMemberIfNotExists: true,
		Familiares: []*models.Familiar{
			{
				Nombre:          "Laura",
				Apellidos:       "Fernández García",
				FechaNacimiento: testutils.TimePtr(time.Now().AddDate(-8, 0, 0)),
				Parentesco:      "Hija",
			},
		},
	}

	// Act
	family, err := familyService.CreateFamilyAtomic(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error when creating familiar, but got nil")
	}

	if family != nil {
		t.Error("Expected nil family, but got a family object")
	}

	// CRITICAL: Verify transaction was rolled back
	if !mockTx.RolledBack {
		t.Error("Expected transaction to be rolled back after familiar creation error")
	}

	if mockTx.Committed {
		t.Error("Transaction should NOT have been committed after error")
	}

	// Verify both member and family were attempted to be created
	if !memberCreated {
		t.Error("Expected member creation to be attempted")
	}

	if !familyCreated {
		t.Error("Expected family creation to be attempted before familiar error")
	}

	// Verify familiar was NOT successfully created
	if len(mockFamilyRepo.CreatedFamiliares) > 0 {
		t.Errorf("Expected 0 familiares created due to error, but found %d", len(mockFamilyRepo.CreatedFamiliares))
	}
}

// TestCreateFamilyAtomic_SuccessfulCreation verifies that when everything is valid,
// the transaction is committed successfully
func TestCreateFamilyAtomic_SuccessfulCreation(t *testing.T) {
	// Arrange
	mockTx := &MockTransaction{}
	mockFamilyRepo := &MockFamilyRepository{
		BeginTransactionFunc: func(ctx context.Context) (output.Transaction, error) {
			return mockTx, nil
		},
	}
	mockMemberRepo := &MockMemberRepository{
		GetByNumeroSocioWithTxFunc: func(ctx context.Context, tx output.Transaction, numeroSocio string) (*models.Member, error) {
			return nil, nil
		},
		GetByIdentityCardWithTxFunc: func(ctx context.Context, tx output.Transaction, identityCard string) (*models.Member, error) {
			return nil, nil
		},
	}

	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}
	familyService := services.NewFamilyService(mockFamilyRepo, mockMemberRepo, mockPaymentRepo, mockFeeRepo)

	req := &input.CreateFamilyAtomicRequest{
		Family: &models.Family{
			NumeroSocio:              "A00126",
			EsposoNombre:             "Miguel",
			EsposoApellidos:          "Rodríguez Torres",
			EsposoDocumentoIdentidad: "12345678Z", // Valid DNI
			EsposoCorreoElectronico:  "miguel@example.com",
			EsposaNombre:             "Elena",
			EsposaApellidos:          "Vega Morales",
			EsposaDocumentoIdentidad: "87654321X", // Valid DNI
		},
		CreateMemberIfNotExists: true,
		Familiares: []*models.Familiar{
			{
				Nombre:          "Sofia",
				Apellidos:       "Rodríguez Vega",
				FechaNacimiento: testutils.TimePtr(time.Now().AddDate(-12, 0, 0)),
				Parentesco:      "Hija",
			},
			{
				Nombre:          "Javier",
				Apellidos:       "Rodríguez Vega",
				FechaNacimiento: testutils.TimePtr(time.Now().AddDate(-9, 0, 0)),
				Parentesco:      "Hijo",
			},
		},
	}

	// Act
	family, err := familyService.CreateFamilyAtomic(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	if family == nil {
		t.Fatal("Expected family to be created, but got nil")
	}

	// Verify transaction was committed
	if !mockTx.Committed {
		t.Error("Expected transaction to be committed after successful creation")
	}

	if mockTx.RolledBack {
		t.Error("Transaction should NOT have been rolled back on success")
	}

	// Verify all entities were created
	if len(mockMemberRepo.CreatedMembers) != 1 {
		t.Errorf("Expected 1 member created, got %d", len(mockMemberRepo.CreatedMembers))
	}

	if len(mockFamilyRepo.CreatedFamilies) != 1 {
		t.Errorf("Expected 1 family created, got %d", len(mockFamilyRepo.CreatedFamilies))
	}

	if len(mockFamilyRepo.CreatedFamiliares) != 2 {
		t.Errorf("Expected 2 familiares created, got %d", len(mockFamilyRepo.CreatedFamiliares))
	}

	// Verify family has correct data
	if family.NumeroSocio != "A00126" {
		t.Errorf("Expected NumeroSocio 'A00126', got '%s'", family.NumeroSocio)
	}

	if family.EsposoNombre != "Miguel" {
		t.Errorf("Expected EsposoNombre 'Miguel', got '%s'", family.EsposoNombre)
	}
}

// TestCreateFamilyAtomic_ValidationPreventsMemberCreation verifies that validation errors
// prevent any database operations from starting
func TestCreateFamilyAtomic_ValidationPreventsMemberCreation(t *testing.T) {
	// Arrange
	mockFamilyRepo := &MockFamilyRepository{}
	mockMemberRepo := &MockMemberRepository{}
	mockPaymentRepo := &MockPaymentRepository{}
	mockFeeRepo := &MockMembershipFeeRepository{}

	familyService := services.NewFamilyService(mockFamilyRepo, mockMemberRepo, mockPaymentRepo, mockFeeRepo)

	testCases := []struct {
		name        string
		request     *input.CreateFamilyAtomicRequest
		expectedErr bool
	}{
		{
			name: "Invalid email format",
			request: &input.CreateFamilyAtomicRequest{
				Family: &models.Family{
					NumeroSocio:              "A00127",
					EsposoNombre:             "Test",
					EsposoApellidos:          "User",
					EsposoDocumentoIdentidad: "12345678Z",
					EsposoCorreoElectronico:  "not-an-email", // Invalid
				},
				CreateMemberIfNotExists: true,
			},
			expectedErr: true,
		},
		{
			name: "Missing esposo nombre",
			request: &input.CreateFamilyAtomicRequest{
				Family: &models.Family{
					NumeroSocio:              "A00128",
					EsposoNombre:             "", // Missing
					EsposoApellidos:          "User",
					EsposoDocumentoIdentidad: "12345678Z",
				},
				CreateMemberIfNotExists: true,
			},
			expectedErr: true,
		},
		{
			name: "Invalid numero socio format",
			request: &input.CreateFamilyAtomicRequest{
				Family: &models.Family{
					NumeroSocio:              "INVALID", // Invalid format
					EsposoNombre:             "Test",
					EsposoApellidos:          "User",
					EsposoDocumentoIdentidad: "12345678Z",
				},
				CreateMemberIfNotExists: true,
			},
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset counters
			mockFamilyRepo.TransactionStarted = false
			mockFamilyRepo.CreatedFamilies = nil
			mockFamilyRepo.CreatedFamiliares = nil
			mockMemberRepo.CreatedMembers = nil

			// Act
			family, err := familyService.CreateFamilyAtomic(context.Background(), tc.request)

			// Assert
			if tc.expectedErr && err == nil {
				t.Errorf("Expected validation error, but got nil")
			}

			if !tc.expectedErr && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if family != nil {
				t.Error("Expected nil family on validation error")
			}

			// CRITICAL: Verify no transaction was started
			if mockFamilyRepo.TransactionStarted {
				t.Error("Transaction should NOT start when validation fails")
			}

			// CRITICAL: Verify no entities were created
			if len(mockMemberRepo.CreatedMembers) > 0 {
				t.Errorf("Expected 0 members on validation error, got %d", len(mockMemberRepo.CreatedMembers))
			}

			if len(mockFamilyRepo.CreatedFamilies) > 0 {
				t.Errorf("Expected 0 families on validation error, got %d", len(mockFamilyRepo.CreatedFamilies))
			}

			if len(mockFamilyRepo.CreatedFamiliares) > 0 {
				t.Errorf("Expected 0 familiares on validation error, got %d", len(mockFamilyRepo.CreatedFamiliares))
			}
		})
	}
}
