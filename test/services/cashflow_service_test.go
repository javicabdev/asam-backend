package services

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/javicabdev/asam-backend/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// createValidCashFlow crea un movimiento de flujo de efectivo válido para las pruebas.
func createValidCashFlow() *models.CashFlow {
	return &models.CashFlow{
		OperationType: models.OperationTypeOtherIncome,
		Amount:        100.0,
		Date:          time.Now(),
		Detail:        "Test income",
	}
}

func TestRegisterMovement(t *testing.T) {
	tests := []struct {
		name      string
		movement  *models.CashFlow
		setupMock func(*test.MockCashFlowRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name:     "successful movement registration",
			movement: createValidCashFlow(),
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("Create", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err, "No debería haber error en un registro exitoso")
			},
		},
		{
			name: "invalid amount",
			movement: &models.CashFlow{
				OperationType: models.OperationTypeOtherIncome,
				Amount:        0, // Monto inválido
				Date:          time.Now(),
				Detail:        "Test",
			},
			setupMock: func(cr *test.MockCashFlowRepository) {},
			wantErr:   true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err, "Debería haber error con un monto inválido")
				assert.True(t, errors.IsValidationError(err), "Debería ser un error de validación")
			},
		},
		{
			name:     "repository error",
			movement: createValidCashFlow(),
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("Create", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(errors.NewDatabaseError("database failure", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err, "Debería haber error cuando falla el repositorio")
				assert.True(t, errors.IsDatabaseError(err), "Debería ser un error de base de datos")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			err := service.RegisterMovement(context.Background(), tt.movement)

			tt.checkErr(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetCurrentBalance(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*test.MockCashFlowRepository)
		wantBalance float64
		wantErr     bool
		checkErr    func(t *testing.T, err error)
	}{
		{
			name: "successful balance calculation",
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("GetBalance", mock.Anything).Return(1000.0, nil)
				cr.On("List", mock.Anything, mock.Anything).Return(
					[]*models.CashFlow{
						{Amount: 500, OperationType: models.OperationTypeOtherIncome},
						{Amount: -200, OperationType: models.OperationTypeCurrentExpense},
					}, nil)
			},
			wantBalance: 1000.0,
			wantErr:     false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "repository error on GetBalance",
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("GetBalance", mock.Anything).Return(0.0, errors.NewDatabaseError("database failure", nil))
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
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			report, err := service.GetCurrentBalance(context.Background())

			if tt.wantErr {
				tt.checkErr(t, err)
				assert.Nil(t, report)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, report)
				assert.Equal(t, tt.wantBalance, report.CurrentBalance)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestGetCashFlowTrends(t *testing.T) {
	tests := []struct {
		name      string
		period    input.Period
		setupMock func(*test.MockCashFlowRepository)
		wantErr   bool
		checkErr  func(t *testing.T, err error)
	}{
		{
			name: "successful trends calculation",
			period: input.Period{
				StartDate: time.Now().AddDate(0, -6, 0),
				EndDate:   time.Now(),
			},
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("List", mock.Anything, mock.Anything).Return(
					[]*models.CashFlow{
						{Amount: 1000, OperationType: models.OperationTypeOtherIncome, Date: time.Now().AddDate(0, -1, 0)},
						{Amount: -500, OperationType: models.OperationTypeCurrentExpense, Date: time.Now()},
					}, nil)
			},
			wantErr: false,
			checkErr: func(t *testing.T, err error) {
				assert.NoError(t, err, "No debería haber error en un cálculo exitoso de tendencias")
			},
		},
		{
			name: "repository failure",
			period: input.Period{
				StartDate: time.Now().AddDate(0, -6, 0),
				EndDate:   time.Now(),
			},
			setupMock: func(cr *test.MockCashFlowRepository) {
				// Es importante devolver nil como primer parámetro cuando hay un error
				cr.On("List", mock.Anything, mock.Anything).Return(nil, errors.NewDatabaseError("failed to fetch trends", nil))
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Error(t, err, "Debería haber error cuando falla el repositorio")
				assert.True(t, errors.IsDatabaseError(err), "Debería ser un error de base de datos")
				// No verificamos mensajes de error específicos, solo el tipo de error
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			trends, err := service.GetCashFlowTrends(context.Background(), tt.period)

			if tt.wantErr {
				tt.checkErr(t, err)
				assert.Nil(t, trends)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trends)
				assert.NotEmpty(t, trends.MonthlyTrends)
			}

			repo.AssertExpectations(t)
		})
	}
}
