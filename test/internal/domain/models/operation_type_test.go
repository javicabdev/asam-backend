package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

func TestOperationType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		opType      models.OperationType
		expectValid bool
	}{
		{
			name:        "membership fee is valid",
			opType:      models.OperationTypeMembershipFee,
			expectValid: true,
		},
		{
			name:        "current expense is valid",
			opType:      models.OperationTypeCurrentExpense,
			expectValid: true,
		},
		{
			name:        "fund delivery is valid",
			opType:      models.OperationTypeFundDelivery,
			expectValid: true,
		},
		{
			name:        "other income is valid",
			opType:      models.OperationTypeOtherIncome,
			expectValid: true,
		},
		{
			name:        "empty operation type is invalid",
			opType:      "",
			expectValid: false,
		},
		{
			name:        "unknown operation type is invalid",
			opType:      "operacion_desconocida",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectValid, tt.opType.IsValid())
		})
	}
}

func TestOperationType_IsIncome(t *testing.T) {
	tests := []struct {
		name         string
		opType       models.OperationType
		expectIncome bool
	}{
		{
			name:         "membership fee is income",
			opType:       models.OperationTypeMembershipFee,
			expectIncome: true,
		},
		{
			name:         "other income is income",
			opType:       models.OperationTypeOtherIncome,
			expectIncome: true,
		},
		{
			name:         "current expense is not income",
			opType:       models.OperationTypeCurrentExpense,
			expectIncome: false,
		},
		{
			name:         "fund delivery is not income",
			opType:       models.OperationTypeFundDelivery,
			expectIncome: false,
		},
		{
			name:         "empty operation type is not income",
			opType:       "",
			expectIncome: false,
		},
		{
			name:         "unknown operation type is not income",
			opType:       "operacion_desconocida",
			expectIncome: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectIncome, tt.opType.IsIncome())
		})
	}
}

func TestOperationType_IsExpense(t *testing.T) {
	tests := []struct {
		name          string
		opType        models.OperationType
		expectExpense bool
	}{
		{
			name:          "current expense is expense",
			opType:        models.OperationTypeCurrentExpense,
			expectExpense: true,
		},
		{
			name:          "fund delivery is expense",
			opType:        models.OperationTypeFundDelivery,
			expectExpense: true,
		},
		{
			name:          "membership fee is not expense",
			opType:        models.OperationTypeMembershipFee,
			expectExpense: false,
		},
		{
			name:          "other income is not expense",
			opType:        models.OperationTypeOtherIncome,
			expectExpense: false,
		},
		{
			name:          "empty operation type is not expense",
			opType:        "",
			expectExpense: false,
		},
		{
			name:          "unknown operation type is not expense",
			opType:        "operacion_desconocida",
			expectExpense: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectExpense, tt.opType.IsExpense())
		})
	}
}
