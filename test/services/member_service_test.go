package services_test

import (
	"context"
	"errors"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Mock del repositorio
type mockMemberRepository struct {
	mock.Mock
}

func (m *mockMemberRepository) Delete(_ context.Context, _ uint) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockMemberRepository) Create(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockMemberRepository) GetByID(ctx context.Context, id uint) (*models.Member, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *mockMemberRepository) GetByNumeroSocio(ctx context.Context, numeroSocio string) (*models.Member, error) {
	args := m.Called(ctx, numeroSocio)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Member), args.Error(1)
}

func (m *mockMemberRepository) Update(ctx context.Context, member *models.Member) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *mockMemberRepository) List(ctx context.Context, filters output.MemberFilters) ([]models.Member, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]models.Member), args.Error(1)
}

// Helper functions
func createValidMember() *models.Member {
	return &models.Member{
		NumeroSocio:   "B0001",
		TipoMembresia: models.TipoMembresiaPIndividual,
		Nombre:        "Juan",
		Apellidos:     "García",
		Direccion:     "Calle Test 1, 1º",
		CodigoPostal:  "08224",
		Poblacion:     "Terrassa",
		Provincia:     "Barcelona",
		Pais:          "España",
		Estado:        models.EstadoActivo,
		FechaAlta:     time.Now().Add(-24 * time.Hour), // 1 día antes
		Nacionalidad:  "Senegal",
	}
}

// Tests
func TestCreateMember(t *testing.T) {
	tests := []struct {
		name      string
		input     model.CreateMemberInput
		setupMock func(*test.MockMemberService)
		wantErr   bool
	}{
		{
			name: "successful create member - individual",
			input: model.CreateMemberInput{
				NumeroSocio:   test.GenerateValidNumeroSocio(1),
				TipoMembresia: model.MembershipTypeIndividual,
				Nombre:        "Juan",
				Apellidos:     "García",
				Direccion:     "Calle Test 1",
				CodigoPostal:  "08001",
				Poblacion:     "Barcelona",
				Provincia:     test.StringPtr("Barcelona"),
				Pais:          test.StringPtr("España"),
			},
			setupMock: func(ms *test.MockMemberService) {
				ms.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.TipoMembresia == models.TipoMembresiaPIndividual &&
						m.NumeroSocio == test.GenerateValidNumeroSocio(1)
				})).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			memberService := new(test.MockMemberService)
			familyService := new(test.MockFamilyService)
			paymentService := new(test.MockPaymentService)
			cashFlowService := new(test.MockCashFlowService)
			authService := new(test.MockAuthService)

			tt.setupMock(memberService)

			resolver := resolvers.NewResolver(
				memberService,
				familyService,
				paymentService,
				cashFlowService,
				authService,
			)

			// Execute
			got, err := resolver.Mutation().CreateMember(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				// Verifica que el tipo de membresía se mapeó correctamente
				if tt.input.TipoMembresia == model.MembershipTypeIndividual {
					assert.Equal(t, models.TipoMembresiaPIndividual, got.TipoMembresia)
				} else {
					assert.Equal(t, models.TipoMembresiaPFamiliar, got.TipoMembresia)
				}
			}

			memberService.AssertExpectations(t)
		})
	}
}

func TestDeactivateMember(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	// Crear una fecha de alta en el pasado
	fechaAlta := time.Now().Add(-24 * time.Hour) // 1 día antes

	tests := []struct {
		name      string
		memberID  uint
		fechaBaja *time.Time
		mockFn    func(*mockMemberRepository)
		wantErr   bool
	}{
		{
			name:     "successful deactivation",
			memberID: 1,
			mockFn: func(repo *mockMemberRepository) {
				member := createValidMember()
				member.ID = 1
				member.FechaAlta = fechaAlta // Usar la fecha de alta anterior
				repo.On("GetByID", mock.Anything, uint(1)).Return(member, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "member not found",
			memberID: 999,
			mockFn: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("miembro no encontrado"))

			},
			wantErr: true,
		},
		{
			name:     "already inactive",
			memberID: 2,
			mockFn: func(repo *mockMemberRepository) {
				member := createValidMember()
				member.ID = 2
				member.Estado = models.EstadoInactivo
				repo.On("GetByID", mock.Anything, uint(2)).Return(member, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			repo := new(mockMemberRepository)
			tt.mockFn(repo)
			service := services.NewMemberService(repo, logger, auditLogger)

			// Execute
			// Si se proporciona una fecha específica, usarla
			err := service.DeactivateMember(context.Background(), tt.memberID, tt.fechaBaja)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestUpdateMember(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	tests := []struct {
		name      string
		member    *models.Member
		setupMock func(*mockMemberRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful update",
			member: func() *models.Member {
				m := createValidMember()
				m.ID = 1
				m.Nombre = "Juan Actualizado"
				m.Profesion = test.StringPtr("Ingeniero")
				return m
			}(),
			setupMock: func(repo *mockMemberRepository) {
				// Mock para obtener el miembro existente
				repo.On("GetByID", mock.Anything, uint(1)).Return(createValidMember(), nil)

				// Mock para actualizar el miembro
				repo.On("Update", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.ID == 1 &&
						m.Nombre == "Juan Actualizado" &&
						m.Direccion == "Calle Test 1, 1º" &&
						m.CodigoPostal == "08224" &&
						m.Poblacion == "Terrassa" &&
						*m.Profesion == "Ingeniero"
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "member not found",
			member: func() *models.Member {
				m := createValidMember()
				m.ID = 999
				return m
			}(),
			setupMock: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("miembro no encontrado"))
			},
			wantErr: true,
			errMsg:  "error checking existing member: miembro no encontrado",
		},
		{
			name: "validation failed - missing required fields",
			member: func() *models.Member {
				m := createValidMember()
				m.ID = 1
				m.Direccion = "" // Campo obligatorio faltante
				return m
			}(),
			setupMock: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(createValidMember(), nil)
			},
			wantErr: true,
			errMsg:  "la dirección es obligatoria",
		},
		{
			name: "validation failed - invalid membership type",
			member: func() *models.Member {
				m := createValidMember()
				m.ID = 1
				m.TipoMembresia = "invalid_type"
				return m
			}(),
			setupMock: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(createValidMember(), nil)
			},
			wantErr: true,
			errMsg:  "tipo de membresía no válido",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			repo := new(mockMemberRepository)
			tt.setupMock(repo)
			service := services.NewMemberService(repo, logger, auditLogger)

			// Execute
			err := service.UpdateMember(context.Background(), tt.member)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMemberByID(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}
	repo := new(mockMemberRepository)

	service := services.NewMemberService(repo, logger, auditLogger)

	// Caso: Miembro encontrado
	repo.On("GetByID", mock.Anything, uint(1)).Return(createValidMember(), nil)

	member, err := service.GetMemberByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, member)

	// Caso: Miembro no encontrado
	repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("miembro no encontrado"))

	member, err = service.GetMemberByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Nil(t, member)
	assert.Contains(t, err.Error(), "miembro no encontrado")
}

func TestGetMemberByNumeroSocio(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}
	repo := new(mockMemberRepository)

	service := services.NewMemberService(repo, logger, auditLogger)

	// Caso: Miembro encontrado
	repo.On("GetByNumeroSocio", mock.Anything, "B0001").Return(createValidMember(), nil)

	member, err := service.GetMemberByNumeroSocio(context.Background(), "B0001")
	assert.NoError(t, err)
	assert.NotNil(t, member)

	// Caso: Miembro no encontrado
	repo.On("GetByNumeroSocio", mock.Anything, "999").Return(nil, errors.New("miembro no encontrado"))

	member, err = service.GetMemberByNumeroSocio(context.Background(), "999")
	assert.Error(t, err)
	assert.Nil(t, member)
	assert.Contains(t, err.Error(), "miembro no encontrado")
}

func TestListMembers(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}
	repo := new(mockMemberRepository)

	service := services.NewMemberService(repo, logger, auditLogger)

	// Caso: Listado exitoso
	repo.On("List", mock.Anything, mock.AnythingOfType("output.MemberFilters")).
		Return([]models.Member{*createValidMember()}, nil)

	filters := input.MemberFilters{
		Estado:        test.StringPtr(models.EstadoActivo),
		TipoMembresia: test.StringPtr(models.TipoMembresiaPIndividual),
	}
	members, err := service.ListMembers(context.Background(), filters)
	assert.NoError(t, err)
	assert.NotEmpty(t, members)

	// Limpia las expectativas anteriores
	repo.ExpectedCalls = nil

	// Caso: Error en el repositorio
	repo.On("List", mock.Anything, mock.AnythingOfType("output.MemberFilters")).
		Return([]models.Member(nil), errors.New("error listing members")) // Asegúrate de usar []models.Member(nil)

	members, err = service.ListMembers(context.Background(), filters)
	assert.Error(t, err)
	assert.Nil(t, members) // Verifica que el resultado sea nil
	assert.Contains(t, err.Error(), "error listing members")
}
