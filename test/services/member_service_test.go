package services_test

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/output"
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
		NumeroSocio:     "001",
		TipoMembresia:   models.TipoMembresiaPIndividual,
		Nombre:          "Juan",
		Apellidos:       "García",
		CalleNumeroPiso: "Calle Test 1, 1º",
		CodigoPostal:    "08224",
		Poblacion:       "Terrassa",
		Provincia:       "Barcelona",
		Pais:            "España",
		Estado:          models.EstadoActivo,
		FechaAlta:       time.Now().Add(-24 * time.Hour), // 1 día antes
		Nacionalidad:    "Senegal",
	}
}

// Tests
func TestCreateMember(t *testing.T) {
	tests := []struct {
		name    string
		member  *models.Member
		mockFn  func(*mockMemberRepository)
		wantErr bool
	}{
		{
			name:   "successful creation",
			member: createValidMember(),
			mockFn: func(repo *mockMemberRepository) {
				repo.On("GetByNumeroSocio", mock.Anything, "001").Return(nil, nil)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "duplicate numero socio",
			member: createValidMember(),
			mockFn: func(repo *mockMemberRepository) {
				repo.On("GetByNumeroSocio", mock.Anything, "001").Return(createValidMember(), nil)
			},
			wantErr: true,
		},
		{
			name: "invalid member data",
			member: &models.Member{
				NumeroSocio: "", // Invalid: empty numero socio
				Nombre:      "Juan",
			},
			mockFn: func(repo *mockMemberRepository) {
				repo.On("GetByNumeroSocio", mock.Anything, "").Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			repo := new(mockMemberRepository)
			tt.mockFn(repo)
			service := services.NewMemberService(repo)

			// Execute
			err := service.CreateMember(context.Background(), tt.member)

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

func TestDeactivateMember(t *testing.T) {
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
				repo.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)
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
			service := services.NewMemberService(repo)

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
