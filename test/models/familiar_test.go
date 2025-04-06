package models_test

import (
	"github.com/javicabdev/asam-backend/test"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

// Helper para crear un Familiar válido
func createValidFamiliar() *models.Familiar {
	return &models.Familiar{
		FamiliaID:         1,
		Nombre:            "Juan",
		Apellidos:         "Pérez",
		FechaNacimiento:   test.TimePtr(time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)),
		DNINIE:            "12345678A",
		CorreoElectronico: "juan.perez@example.com",
		Parentesco:        "Hijo",
	}
}

// Tests de validaciones básicas
func TestFamiliarValidation(t *testing.T) {
	familiar := createValidFamiliar()

	// Caso válido
	assert.NoError(t, familiar.Validate())

	// Caso: FamiliaID faltante
	familiar.FamiliaID = 0
	err := familiar.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID de familia es requerido")

	// Caso: Nombre faltante
	familiar.FamiliaID = 1 // Restaurar FamiliaID válido
	familiar.Nombre = ""
	err = familiar.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nombre es requerido")

	// Caso: Apellidos faltantes
	familiar.Nombre = "Juan" // Restaurar Nombre válido
	familiar.Apellidos = ""
	err = familiar.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "apellidos son requeridos")

	// Caso: Parentesco faltante
	familiar.Apellidos = "Pérez" // Restaurar Apellidos válidos
	familiar.Parentesco = ""
	err = familiar.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parentesco es requerido")
}

// Tests de hooks de GORM
func TestFamiliar_BeforeCreate(t *testing.T) {
	familiar := createValidFamiliar()

	// Caso válido
	assert.NoError(t, familiar.BeforeCreate(nil))

	// Caso inválido
	familiar.Nombre = ""
	assert.Error(t, familiar.BeforeCreate(nil))
}

func TestFamiliar_BeforeSave(t *testing.T) {
	familiar := createValidFamiliar()

	// Caso válido
	assert.NoError(t, familiar.BeforeSave(nil))

	// Caso inválido
	familiar.Parentesco = ""
	assert.Error(t, familiar.BeforeSave(nil))
}
