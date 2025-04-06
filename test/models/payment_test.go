package models_test

import (
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/assert"
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
	payment := createValidPayment()

	// Caso válido
	assert.NoError(t, payment.Validate())

	// Caso: MemberID y FamilyID son nulos
	payment.MemberID = 0
	payment.FamilyID = nil
	err := payment.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment must be associated with either a member or family")

	// Caso: Monto inválido
	payment.MemberID = 1 // Restaurar un MemberID válido
	payment.Amount = 0
	err = payment.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "payment amount must be greater than 0")
}

// Tests de lógica de negocio en MembershipFee
func TestMembershipFee_Calculate(t *testing.T) {
	fee := createValidMembershipFee()

	// Caso individual
	amount := fee.Calculate(false)
	assert.Equal(t, 50.0, amount)

	// Caso familiar
	amount = fee.Calculate(true)
	assert.Equal(t, 70.0, amount)
}

// Tests de hooks de GORM
func TestPayment_BeforeCreate(t *testing.T) {
	payment := createValidPayment()

	// Caso válido
	assert.NoError(t, payment.BeforeCreate(nil))

	// Caso inválido
	payment.MemberID = 0
	payment.FamilyID = nil
	assert.Error(t, payment.BeforeCreate(nil))
}

func TestPayment_BeforeUpdate(t *testing.T) {
	payment := createValidPayment()

	// Caso válido
	assert.NoError(t, payment.BeforeUpdate(nil))

	// Caso inválido
	payment.Amount = 0
	assert.Error(t, payment.BeforeUpdate(nil))
}
