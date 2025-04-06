package models_test

import (
	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Helper para crear un Telefono válido
func createValidTelefono() *models.Telefono {
	return &models.Telefono{
		NumeroTelefono:  "123456789",
		ContactableID:   1,
		ContactableType: "Member",
	}
}

// Tests de validaciones básicas
func TestTelefonoValidation(t *testing.T) {
	telefono := createValidTelefono()

	// Caso válido
	assert.NoError(t, telefono.Validate())

	// Caso: Número de teléfono faltante
	telefono.NumeroTelefono = ""
	err := telefono.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "número de teléfono es requerido")
}

// Tests de hooks de GORM
func TestTelefono_BeforeCreate(t *testing.T) {
	telefono := createValidTelefono()

	// Caso válido
	assert.NoError(t, telefono.BeforeCreate(nil))

	// Caso inválido
	telefono.NumeroTelefono = ""
	assert.Error(t, telefono.BeforeCreate(nil))
}
