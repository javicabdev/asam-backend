package services

import (
	"context"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/mock"
)

// Helper types
type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) SendEmail(ctx context.Context, to string, subject string, body string) error {
	args := m.Called(ctx, to, subject, body)
	return args.Error(0)
}

func (m *mockNotificationService) SendSMS(ctx context.Context, to string, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

type mockFeeCalculator struct {
	mock.Mock
}

func (m *mockFeeCalculator) CalculateBaseFee(year, month int) float64 {
	args := m.Called(year, month)
	return args.Get(0).(float64)
}

func (m *mockFeeCalculator) CalculateFamilyFee(year, month int) float64 {
	args := m.Called(year, month)
	return args.Get(0).(float64)
}

func (m *mockFeeCalculator) CalculateLateFee(daysLate int) float64 {
	args := m.Called(daysLate)
	return args.Get(0).(float64)
}

func createValidPayment() *models.Payment {
	return &models.Payment{
		MemberID:      1,
		Amount:        100.0,
		PaymentDate:   time.Now(),
		PaymentMethod: "efectivo",
		Status:        models.PaymentStatusPending,
	}
}
