package services

import (
	"context"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/internal/domain/services"
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"github.com/javicabdev/asam-backend/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
	}{
		{
			name:     "successful movement registration",
			movement: createValidCashFlow(),
			setupMock: func(cr *test.MockCashFlowRepository) {
				cr.On("Create", mock.Anything, mock.AnythingOfType("*models.CashFlow")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid amount",
			movement: &models.CashFlow{
				OperationType: models.OperationTypeOtherIncome,
				Amount:        0, // Invalid amount
				Date:          time.Now(),
				Detail:        "Test",
			},
			setupMock: func(cr *test.MockCashFlowRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			err := service.RegisterMovement(context.Background(), tt.movement)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			report, err := service.GetCurrentBalance(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(test.MockCashFlowRepository)
			tt.setupMock(repo)

			service := services.NewCashFlowService(repo)
			trends, err := service.GetCashFlowTrends(context.Background(), tt.period)

			if tt.wantErr {
				assert.Error(t, err)
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
