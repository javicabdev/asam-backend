package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/internal/ports/output"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Member), args.Error(1)
}

func (m *mockMemberRepository) GetLastMemberNumberByPrefix(ctx context.Context, prefix string) (string, error) {
	args := m.Called(ctx, prefix)
	return args.String(0), args.Error(1)
}

// Test básico de creación de miembro
func TestCreateMember(t *testing.T) {
	// Logger de prueba
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("successful create member", func(t *testing.T) {
		repo := new(mockMemberRepository)

		// El servicio primero verifica si el miembro existe
		repo.On("GetByNumeroSocio", mock.Anything, "B00001").Return(nil, nil)
		// Luego crea el miembro
		repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)

		service := services.NewMemberService(repo, logger, auditLogger)

		member := &models.Member{
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			RegistrationDate: time.Now(),
			Nationality:      "España",
		}

		err := service.CreateMember(context.Background(), member)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// Test básico de obtención de miembro por ID
func TestGetMemberByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("successful retrieval", func(t *testing.T) {
		repo := new(mockMemberRepository)
		expectedMember := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			RegistrationDate: time.Now(),
			Nationality:      "España",
		}
		repo.On("GetByID", mock.Anything, uint(1)).Return(expectedMember, nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		member, err := service.GetMemberByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, member)
		assert.Equal(t, uint(1), member.ID)
		repo.AssertExpectations(t)
	})
}

// Test básico de desactivación de miembro
func TestDeactivateMember(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("successful deactivation", func(t *testing.T) {
		repo := new(mockMemberRepository)
		// Establecer fecha de registro hace un mes para evitar problemas de validación
		registrationDate := time.Now().AddDate(0, -1, 0)
		member := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			RegistrationDate: registrationDate,
			Nationality:      "España",
		}
		repo.On("GetByID", mock.Anything, uint(1)).Return(member, nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		err := service.DeactivateMember(context.Background(), 1, nil)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// Test básico de obtención del siguiente número de socio
func TestGetNextMemberNumber(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("first individual member", func(t *testing.T) {
		repo := new(mockMemberRepository)
		repo.On("GetLastMemberNumberByPrefix", mock.Anything, "B").Return("", nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		result, err := service.GetNextMemberNumber(context.Background(), false)

		assert.NoError(t, err)
		assert.Equal(t, "B00001", result)
		repo.AssertExpectations(t)
	})
}

// Test básico de verificación de existencia de número de socio
func TestCheckMemberNumberExists(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("existing member", func(t *testing.T) {
		repo := new(mockMemberRepository)
		member := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			RegistrationDate: time.Now(),
			Nationality:      "España",
		}
		repo.On("GetByNumeroSocio", mock.Anything, "B00001").Return(member, nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		exists, err := service.CheckMemberNumberExists(context.Background(), "B00001")

		assert.NoError(t, err)
		assert.True(t, exists)
		repo.AssertExpectations(t)
	})
}

// Test básico de actualización de miembro
func TestUpdateMember(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("successful update", func(t *testing.T) {
		repo := new(mockMemberRepository)
		existingMember := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			RegistrationDate: time.Now(),
			Nationality:      "España",
		}
		repo.On("GetByID", mock.Anything, uint(1)).Return(existingMember, nil)
		repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Member")).Return(nil)

		service := services.NewMemberService(repo, logger, auditLogger)

		updatedMember := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			Name:             "Juan Actualizado",
			Surnames:         "García",
			Address:          "Nueva Dirección",
			Postcode:         "08002",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoActivo,
			MembershipType:   models.TipoMembresiaPIndividual,
			RegistrationDate: time.Now(),
			Nationality:      "España",
		}

		err := service.UpdateMember(context.Background(), updatedMember)
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

// Test básico de listado de miembros
func TestListMembers(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("successful listing", func(t *testing.T) {
		repo := new(mockMemberRepository)
		// El repositorio devuelve []models.Member (valores)
		expectedMembers := []models.Member{
			{
				ID:               1,
				MembershipNumber: "B00001",
				MembershipType:   models.TipoMembresiaPIndividual,
				Name:             "Juan",
				Surnames:         "García",
				Address:          "Calle Test 1",
				Postcode:         "08001",
				City:             "Barcelona",
				Province:         "Barcelona",
				Country:          "España",
				State:            models.EstadoActivo,
				RegistrationDate: time.Now(),
				Nationality:      "España",
			},
			{
				ID:               2,
				MembershipNumber: "B00002",
				MembershipType:   models.TipoMembresiaPIndividual,
				Name:             "María",
				Surnames:         "López",
				Address:          "Calle Test 2",
				Postcode:         "08002",
				City:             "Barcelona",
				Province:         "Barcelona",
				Country:          "España",
				State:            models.EstadoActivo,
				RegistrationDate: time.Now(),
				Nationality:      "España",
			},
		}
		repo.On("List", mock.Anything, mock.AnythingOfType("output.MemberFilters")).Return(expectedMembers, nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		// El servicio devuelve []*models.Member (punteros)
		members, err := service.ListMembers(context.Background(), input.MemberFilters{})

		assert.NoError(t, err)
		assert.Len(t, members, 2)
		// Verificar que son punteros
		assert.NotNil(t, members[0])
		assert.NotNil(t, members[1])
		assert.Equal(t, "Juan", members[0].Name)
		assert.Equal(t, "María", members[1].Name)
		repo.AssertExpectations(t)
	})
}

// Test básico de errores
func TestServiceErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	auditLogger := &test.MockAuditLogger{}

	t.Run("member not found", func(t *testing.T) {
		repo := new(mockMemberRepository)
		repo.On("GetByID", mock.Anything, uint(999)).Return(nil, nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		member, err := service.GetMemberByID(context.Background(), 999)

		assert.Error(t, err)
		assert.Nil(t, member)
		assert.True(t, errors.IsNotFoundError(err))
		repo.AssertExpectations(t)
	})

	t.Run("already inactive member", func(t *testing.T) {
		repo := new(mockMemberRepository)
		member := &models.Member{
			ID:               1,
			MembershipNumber: "B00001",
			MembershipType:   models.TipoMembresiaPIndividual,
			Name:             "Juan",
			Surnames:         "García",
			Address:          "Calle Test 1",
			Postcode:         "08001",
			City:             "Barcelona",
			Province:         "Barcelona",
			Country:          "España",
			State:            models.EstadoInactivo,
			RegistrationDate: time.Now(),
			Nationality:      "España",
			LeavingDate:      test.TimePtr(time.Now()),
		}
		repo.On("GetByID", mock.Anything, uint(1)).Return(member, nil)

		service := services.NewMemberService(repo, logger, auditLogger)
		err := service.DeactivateMember(context.Background(), 1, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ya está dado de baja")
		repo.AssertExpectations(t)
	})
}
