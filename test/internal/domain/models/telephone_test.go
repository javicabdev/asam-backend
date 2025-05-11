package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// Helper para crear un Telephone válido
func createValidTelefono() *models.Telephone {
	return &models.Telephone{
		NumeroTelefono:  "123456789",
		ContactableID:   1,
		ContactableType: "Member",
	}
}

// Tests de validaciones básicas
func TestTelefonoValidation(t *testing.T) {
	tests := []struct {
		name         string
		telefono     *models.Telephone
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid phone number",
			telefono:    createValidTelefono(),
			expectError: false,
		},
		{
			name: "empty phone number",
			telefono: &models.Telephone{
				NumeroTelefono:  "",
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError:  true,
			errorMessage: "número de teléfono es requerido",
		},
		{
			name: "very long phone number is still valid",
			telefono: &models.Telephone{
				NumeroTelefono:  "123456789012345678901234567890", // Very long number
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError: false,
		},
		{
			name: "special characters in phone number",
			telefono: &models.Telephone{
				NumeroTelefono:  "+34 (123) 456-789",
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.telefono.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Tests de hooks de GORM
func TestTelefono_BeforeCreate(t *testing.T) {
	tests := []struct {
		name         string
		telefono     *models.Telephone
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid telephone",
			telefono:    createValidTelefono(),
			expectError: false,
		},
		{
			name: "empty phone number",
			telefono: &models.Telephone{
				NumeroTelefono:  "",
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError:  true,
			errorMessage: "número de teléfono es requerido",
		},
		{
			name: "valid phone with formatting",
			telefono: &models.Telephone{
				NumeroTelefono:  "+34-123-456-789",
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError: false,
		},
		{
			name: "valid phone with spaces",
			telefono: &models.Telephone{
				NumeroTelefono:  "+34 123 456 789",
				ContactableID:   1,
				ContactableType: "Member",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.telefono.BeforeCreate(nil)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test para validar que los campos ContactableID y ContactableType se guardan correctamente
func TestTelefono_Contactable(t *testing.T) {
	tests := []struct {
		name            string
		contactableID   uint
		contactableType string
	}{
		{
			name:            "member contact",
			contactableID:   1,
			contactableType: "Member",
		},
		{
			name:            "family contact",
			contactableID:   2,
			contactableType: "Family",
		},
		{
			name:            "familiar contact",
			contactableID:   3,
			contactableType: "Familiar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telefono := &models.Telephone{
				NumeroTelefono:  "123456789",
				ContactableID:   tt.contactableID,
				ContactableType: tt.contactableType,
			}

			// Verificar que los campos se guardan correctamente
			assert.Equal(t, tt.contactableID, telefono.ContactableID)
			assert.Equal(t, tt.contactableType, telefono.ContactableType)

			// Verificar que la validación pasa
			assert.NoError(t, telefono.Validate())
		})
	}
}
