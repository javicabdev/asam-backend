package models_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// Helper para crear un Payment válido
func createValidPayment() *models.Payment {
	return &models.Payment{
		MemberID:      1,
		Amount:        100.0,
		PaymentDate:   time.Now(),
		Status:        models.PaymentStatusPending,
		PaymentMethod: "credit_card",
	}
}

// Helper para crear una MembershipFee válida
func createValidMembershipFee() *models.MembershipFee {
	return &models.MembershipFee{
		Year:           2023,
		Month:          1,
		BaseFeeAmount:  50.0,
		FamilyFeeExtra: 20.0,
		Status:         models.PaymentStatusPending,
		DueDate:        time.Now().AddDate(0, 1, 0),
	}
}

// Tests de validaciones básicas
func TestPaymentValidation(t *testing.T) {
	tests := []struct {
		name         string
		payment      *models.Payment
		expectError  bool
		errorMessage string
		errorFields  map[string]bool
	}{
		{
			name:        "valid payment",
			payment:     createValidPayment(),
			expectError: false,
		},
		{
			name: "missing both member and family",
			payment: &models.Payment{
				MemberID:      0,
				FamilyID:      nil,
				Amount:        100.0,
				PaymentDate:   time.Now(),
				Status:        models.PaymentStatusPending,
				PaymentMethod: "credit_card",
			},
			expectError:  true,
			errorMessage: "payment must be associated with either a member or family",
			errorFields:  map[string]bool{"MemberID": true, "FamilyID": true},
		},
		{
			name: "zero amount",
			payment: &models.Payment{
				MemberID:      1,
				Amount:        0,
				PaymentDate:   time.Now(),
				Status:        models.PaymentStatusPending,
				PaymentMethod: "credit_card",
			},
			expectError:  true,
			errorMessage: "payment amount must be greater than 0",
			errorFields:  map[string]bool{"Amount": true},
		},
		{
			name: "negative amount",
			payment: &models.Payment{
				MemberID:      1,
				Amount:        -50.0,
				PaymentDate:   time.Now(),
				Status:        models.PaymentStatusPending,
				PaymentMethod: "credit_card",
			},
			expectError:  true,
			errorMessage: "payment amount must be greater than 0",
			errorFields:  map[string]bool{"Amount": true},
		},
		{
			name: "valid with family ID",
			payment: func() *models.Payment {
				familyID := uint(1)
				return &models.Payment{
					MemberID:      0, // No member ID
					FamilyID:      &familyID,
					Amount:        100.0,
					PaymentDate:   time.Now(),
					Status:        models.PaymentStatusPending,
					PaymentMethod: "credit_card",
				}
			}(),
			expectError: false,
		},
		{
			name: "valid with both member and family ID",
			payment: func() *models.Payment {
				familyID := uint(1)
				return &models.Payment{
					MemberID:      1,
					FamilyID:      &familyID,
					Amount:        100.0,
					PaymentDate:   time.Now(),
					Status:        models.PaymentStatusPending,
					PaymentMethod: "credit_card",
				}
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}

				// Verificar campos específicos con error
				if len(tt.errorFields) > 0 {
					var appErr *appErrors.AppError
					if assert.ErrorAs(t, err, &appErr) && appErr.Fields != nil {
						for field := range tt.errorFields {
							assert.Contains(t, appErr.Fields, field, "el campo %s debería estar en los errores", field)
						}
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Tests de lógica de negocio en MembershipFee
func TestMembershipFee_Calculate(t *testing.T) {
	tests := []struct {
		name           string
		fee            *models.MembershipFee
		isFamily       bool
		expectedAmount float64
	}{
		{
			name:           "base fee for individual",
			fee:            createValidMembershipFee(),
			isFamily:       false,
			expectedAmount: 50.0,
		},
		{
			name:           "base fee plus family extra for family",
			fee:            createValidMembershipFee(),
			isFamily:       true,
			expectedAmount: 70.0,
		},
		{
			name: "custom base fee for individual",
			fee: &models.MembershipFee{
				Year:           2023,
				Month:          1,
				BaseFeeAmount:  75.0,
				FamilyFeeExtra: 25.0,
				Status:         models.PaymentStatusPending,
				DueDate:        time.Now().AddDate(0, 1, 0),
			},
			isFamily:       false,
			expectedAmount: 75.0,
		},
		{
			name: "custom base fee plus family extra for family",
			fee: &models.MembershipFee{
				Year:           2023,
				Month:          1,
				BaseFeeAmount:  75.0,
				FamilyFeeExtra: 25.0,
				Status:         models.PaymentStatusPending,
				DueDate:        time.Now().AddDate(0, 1, 0),
			},
			isFamily:       true,
			expectedAmount: 100.0,
		},
		{
			name: "zero family extra",
			fee: &models.MembershipFee{
				Year:           2023,
				Month:          1,
				BaseFeeAmount:  80.0,
				FamilyFeeExtra: 0.0,
				Status:         models.PaymentStatusPending,
				DueDate:        time.Now().AddDate(0, 1, 0),
			},
			isFamily:       true,
			expectedAmount: 80.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := tt.fee.Calculate(tt.isFamily)
			assert.Equal(t, tt.expectedAmount, amount)
		})
	}
}

// Tests de hooks de GORM
func TestPayment_BeforeCreate(t *testing.T) {
	tests := []struct {
		name        string
		payment     *models.Payment
		expectError bool
		errorField  string
	}{
		{
			name:        "valid payment",
			payment:     createValidPayment(),
			expectError: false,
		},
		{
			name: "missing member and family",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.MemberID = 0
				p.FamilyID = nil
				return p
			}(),
			expectError: true,
			errorField:  "MemberID",
		},
		{
			name: "zero amount",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.Amount = 0
				return p
			}(),
			expectError: true,
			errorField:  "Amount",
		},
		{
			name: "negative amount",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.Amount = -10.0
				return p
			}(),
			expectError: true,
			errorField:  "Amount",
		},
		{
			name: "valid with family ID",
			payment: func() *models.Payment {
				p := createValidPayment()
				familyID := uint(1)
				p.FamilyID = &familyID
				return p
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.BeforeCreate(nil)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					var appErr *appErrors.AppError
					if assert.ErrorAs(t, err, &appErr) {
						assert.Contains(t, appErr.Fields, tt.errorField)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPayment_BeforeUpdate(t *testing.T) {
	tests := []struct {
		name        string
		payment     *models.Payment
		expectError bool
		errorField  string
	}{
		{
			name:        "valid payment",
			payment:     createValidPayment(),
			expectError: false,
		},
		{
			name: "zero amount",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.Amount = 0
				return p
			}(),
			expectError: true,
			errorField:  "Amount",
		},
		{
			name: "update to remove member and family",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.MemberID = 0
				p.FamilyID = nil
				return p
			}(),
			expectError: true,
			errorField:  "MemberID",
		},
		{
			name: "update with valid family ID",
			payment: func() *models.Payment {
				p := createValidPayment()
				p.MemberID = 0 // Remove member ID
				familyID := uint(1)
				p.FamilyID = &familyID // Add family ID
				return p
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payment.BeforeUpdate(nil)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					var appErr *appErrors.AppError
					if assert.ErrorAs(t, err, &appErr) {
						assert.Contains(t, appErr.Fields, tt.errorField)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test para diferentes estados de pago
func TestPayment_Status(t *testing.T) {
	tests := []struct {
		name   string
		status models.PaymentStatus
	}{
		{
			name:   "pending status",
			status: models.PaymentStatusPending,
		},
		{
			name:   "paid status",
			status: models.PaymentStatusPaid,
		},
		{
			name:   "cancelled status",
			status: models.PaymentStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := createValidPayment()
			payment.Status = tt.status

			// El test solo verifica que los estados se pueden asignar y son válidos
			assert.Equal(t, tt.status, payment.Status)
			assert.NoError(t, payment.Validate())
		})
	}
}
