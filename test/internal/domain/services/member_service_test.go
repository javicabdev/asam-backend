package services_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/javicabdev/asam-backend/internal/adapters/gql/model"
	"github.com/javicabdev/asam-backend/internal/adapters/gql/resolvers"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"github.com/javicabdev/asam-backend/pkg/auth"
	"github.com/javicabdev/asam-backend/pkg/constants"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
)

// Mock del repositorio
type mockMemberRepository struct {
	mock.Mock
}

func (m *mockMemberRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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

// Tests
func TestCreateMember(t *testing.T) {
	tests := []struct {
		name      string
		input     model.CreateMemberInput
		setupMock func(*test.MockMemberService)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "successful create member - individual",
			input: model.CreateMemberInput{
				NumeroSocio:     test.GenerateValidNumeroSocio(1),
				TipoMembresia:   model.MembershipTypeIndividual,
				Nombre:          "Juan",
				Apellidos:       "García",
				CalleNumeroPiso: "Calle Test 1",
				CodigoPostal:    "08001",
				Poblacion:       "Barcelona",
				Provincia:       test.StringPtr("Barcelona"),
				Pais:            test.StringPtr("España"),
			},
			setupMock: func(ms *test.MockMemberService) {
				ms.On("CreateMember", mock.Anything, mock.MatchedBy(func(m *models.Member) bool {
					return m.MembershipType == models.TipoMembresiaPIndividual &&
						m.MembershipNumber == test.GenerateValidNumeroSocio(1)
				})).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "validation failed - empty numero socio",
			input: model.CreateMemberInput{
				NumeroSocio:   "",
				TipoMembresia: model.MembershipTypeIndividual,
				Nombre:        "Juan",
				Apellidos:     "García",
			},
			setupMock: func(_ *test.MockMemberService) {
				// No se llama al servicio porque falla la validación
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name: "database error",
			input: model.CreateMemberInput{
				NumeroSocio:     test.GenerateValidNumeroSocio(1),
				TipoMembresia:   model.MembershipTypeIndividual,
				Nombre:          "Juan",
				Apellidos:       "García",
				CalleNumeroPiso: "Calle Test 1",
				CodigoPostal:    "08001",
				Poblacion:       "Barcelona",
				Provincia:       test.StringPtr("Barcelona"),
				Pais:            test.StringPtr("España"),
			},
			setupMock: func(ms *test.MockMemberService) {
				ms.On("CreateMember", mock.Anything, mock.AnythingOfType("*models.Member")).
					Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memberService := new(test.MockMemberService)
			familyService := new(test.MockFamilyService)
			paymentService := new(test.MockPaymentService)
			cashFlowService := new(test.MockCashFlowService)
			authService := new(test.MockAuthService)
			userService := new(test.MockUserService)

			// Crear un mockUser para autenticación
			mockUser := &models.User{
				Role: models.RoleAdmin,
			}

			// Configurar el contexto con el usuario
			ctx := context.WithValue(context.Background(), constants.UserContextKey, mockUser)

			tt.setupMock(memberService)

			// Crear mock logger para el rate limiter
			mockLogger := &test.MockLogger{}
			loginRateLimiter := auth.NewLoginRateLimiter(mockLogger)

			resolver := resolvers.NewResolver(
				memberService,
				familyService,
				paymentService,
				cashFlowService,
				authService,
				userService,
				loginRateLimiter,
			)

			got, err := resolver.Mutation().CreateMember(ctx, tt.input)

			tt.checkErr(t, err)
			if !tt.wantErr {
				assert.NotNil(t, got)
				assert.Equal(t, models.TipoMembresiaPIndividual, got.MembershipType)
			} else {
				assert.Nil(t, got)
			}
			memberService.AssertExpectations(t)
		})
	}
}

func TestDeactivateMember(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	tests := []struct {
		name      string
		memberID  uint
		fechaBaja *time.Time
		setupRepo func(*mockMemberRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "successful deactivation",
			memberID: 1,
			setupRepo: func(repo *mockMemberRepository) {
				member := test.CreateValidMember()
				member.ID = 1
				repo.On("GetByID", mock.Anything, uint(1)).Return(member, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "member not found",
			memberID: 999,
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("member"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name:     "already inactive",
			memberID: 2,
			setupRepo: func(repo *mockMemberRepository) {
				member := test.CreateValidMember()
				member.ID = 2
				member.State = models.EstadoInactivo
				repo.On("GetByID", mock.Anything, uint(2)).Return(member, nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), "ya está dado de baja")
			},
		},
		{
			name:     "database error",
			memberID: 1,
			setupRepo: func(repo *mockMemberRepository) {
				// Database error when getting member
				repo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockMemberRepository)
			tt.setupRepo(repo)

			service := services.NewMemberService(repo, logger, auditLogger)
			err := service.DeactivateMember(context.Background(), tt.memberID, tt.fechaBaja)

			tt.checkErr(t, err)
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
		setupRepo func(*mockMemberRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "successful update",
			member: func() *models.Member {
				m := test.CreateValidMember()
				m.ID = 1
				m.Name = "Juan Actualizado"
				m.Profession = test.StringPtr("Ingeniero")
				return m
			}(),
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "member not found",
			member: func() *models.Member {
				m := test.CreateValidMember()
				m.ID = 999
				return m
			}(),
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("member"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name: "validation failed - missing required fields",
			member: func() *models.Member {
				m := test.CreateValidMember()
				m.ID = 1
				m.Address = ""
				return m
			}(),
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name: "validation failed - invalid membership type",
			member: func() *models.Member {
				m := test.CreateValidMember()
				m.ID = 1
				m.MembershipType = "invalid_type"
				return m
			}(),
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsValidationError(err), "debería ser un error de validación")
			},
		},
		{
			name: "database error",
			member: func() *models.Member {
				m := test.CreateValidMember()
				m.ID = 1
				m.Name = "Juan Actualizado"
				return m
			}(),
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).
					Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockMemberRepository)
			tt.setupRepo(repo)

			service := services.NewMemberService(repo, logger, auditLogger)
			err := service.UpdateMember(context.Background(), tt.member)

			tt.checkErr(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMemberByID(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	tests := []struct {
		name      string
		memberID  uint
		setupRepo func(*mockMemberRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "successful retrieval",
			memberID: 1,
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(test.CreateValidMember(), nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "member not found",
			memberID: 999,
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.NewNotFoundError("member"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name:     "database error",
			memberID: 1,
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByID", mock.Anything, uint(1)).Return(nil, errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear un nuevo mock para cada test
			repo := new(mockMemberRepository)
			tt.setupRepo(repo)

			service := services.NewMemberService(repo, logger, auditLogger)
			member, err := service.GetMemberByID(context.Background(), tt.memberID)

			tt.checkErr(t, err)
			if !tt.wantErr {
				assert.NotNil(t, member)
			} else {
				assert.Nil(t, member)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMemberByNumeroSocio(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	tests := []struct {
		name        string
		numeroSocio string
		setupRepo   func(*mockMemberRepository)
		wantErr     bool
		checkErr    func(t *testing.T, err error)
	}{
		{
			name:        "successful retrieval",
			numeroSocio: "B0001",
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByNumeroSocio", mock.Anything, "B0001").Return(test.CreateValidMember(), nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:        "member not found",
			numeroSocio: "999",
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("GetByNumeroSocio", mock.Anything, "999").Return(nil, errors.NewNotFoundError("member"))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFoundError(err), "debería ser un error de no encontrado")
			},
		},
		{
			name:        "database error",
			numeroSocio: "B0001",
			setupRepo: func(repo *mockMemberRepository) {
				// Database error
				repo.On("GetByNumeroSocio", mock.Anything, "B0001").Return(nil, errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear un nuevo mock para cada test
			repo := new(mockMemberRepository)
			tt.setupRepo(repo)

			service := services.NewMemberService(repo, logger, auditLogger)
			member, err := service.GetMemberByNumeroSocio(context.Background(), tt.numeroSocio)

			tt.checkErr(t, err)
			if !tt.wantErr {
				assert.NotNil(t, member)
			} else {
				assert.Nil(t, member)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestListMembers(t *testing.T) {
	logger := &test.MockLogger{}
	auditLogger := &test.MockAuditLogger{}

	tests := []struct {
		name      string
		filters   input.MemberFilters
		setupRepo func(*mockMemberRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "successful listing",
			filters: input.MemberFilters{
				State:          test.StringPtr(models.EstadoActivo),
				MembershipType: test.StringPtr(models.TipoMembresiaPIndividual),
			},
			setupRepo: func(repo *mockMemberRepository) {
				repo.On("List", mock.Anything, mock.AnythingOfType("output.MemberFilters")).
					Return([]models.Member{*test.CreateValidMember()}, nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "repository error",
			filters: input.MemberFilters{
				State: test.StringPtr(models.EstadoActivo),
			},
			setupRepo: func(repo *mockMemberRepository) {
				// Database error
				repo.On("List", mock.Anything, mock.AnythingOfType("output.MemberFilters")).
					Return([]models.Member{}, errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.True(t, errors.IsDatabaseError(err), "debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crear un nuevo mock para cada test
			repo := new(mockMemberRepository)
			tt.setupRepo(repo)

			service := services.NewMemberService(repo, logger, auditLogger)
			members, err := service.ListMembers(context.Background(), tt.filters)

			tt.checkErr(t, err)
			if !tt.wantErr {
				assert.NotEmpty(t, members)
			} else {
				assert.Empty(t, members)
			}
			repo.AssertExpectations(t)
		})
	}
}
