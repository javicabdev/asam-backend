package models_test

import (
	"github.com/javicabdev/asam-backend/test"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
	"github.com/javicabdev/asam-backend/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// Tests de validaciones básicas
func TestValidateBasicFields(t *testing.T) {
	member := test.CreateValidMember()

	// Caso válido
	assert.NoError(t, member.Validate())

	// Caso: falta NumeroSocio
	member.NumeroSocio = ""
	err := member.Validate()
	assert.Error(t, err)
	// Verificamos que sea un error de validación
	assert.Contains(t, err.Error(), "VALIDATION_FAILED")
	// Verificamos que contenga el campo numeroSocio
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.NotNil(t, appErr.Fields)
	assert.Contains(t, appErr.Fields, "numeroSocio")
}

// Tests de validaciones de fechas
func TestValidateDates(t *testing.T) {
	member := test.CreateValidMember()

	// Caso válido
	assert.NoError(t, member.Validate())

	// Caso: FechaBaja antes de FechaAlta
	fechaBaja := member.FechaAlta.Add(-24 * time.Hour) // Crear valor de tiempo
	member.FechaBaja = &fechaBaja                      // Asignar puntero
	err := member.Validate()
	assert.Error(t, err)

	// Verificamos que sea un error de validación de fechas
	assert.Contains(t, err.Error(), "VALIDATION_FAILED")

	// Verificamos que contenga el campo fechaBaja
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.NotNil(t, appErr.Fields)
	assert.Contains(t, appErr.Fields, "fechaBaja")

	// Caso: FechaBaja igual a FechaAlta
	member.FechaBaja = &member.FechaAlta
	err = member.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VALIDATION_FAILED")
	appErr, ok = err.(*errors.AppError)
	assert.True(t, ok)
	assert.Contains(t, appErr.Fields, "fechaBaja")
}

// Tests de lógica de negocio
func TestIsActive(t *testing.T) {
	member := test.CreateValidMember()

	// Caso activo
	assert.True(t, member.IsActive())

	// Caso inactivo
	member.Estado = models.EstadoInactivo
	assert.False(t, member.IsActive())
}

func TestIsFamiliar(t *testing.T) {
	member := test.CreateValidMember()

	// Caso no familiar
	assert.False(t, member.IsFamiliar())

	// Caso familiar
	member.TipoMembresia = models.TipoMembresiaPFamiliar
	assert.True(t, member.IsFamiliar())
}

func TestNombreCompleto(t *testing.T) {
	member := test.CreateValidMember()

	// Validar nombre completo
	expected := member.Nombre + " " + member.Apellidos
	actual := member.NombreCompleto()

	assert.Equal(t, expected, actual)
}

// Tests de validación de estado
func TestValidateStatus(t *testing.T) {
	member := test.CreateValidMember()

	// Caso válido
	assert.NoError(t, member.Validate())

	// Caso: estado inactivo sin FechaBaja
	member.Estado = models.EstadoInactivo
	member.FechaBaja = nil
	err := member.Validate()
	assert.Error(t, err)

	// Verificamos que sea un error de validación
	assert.Contains(t, err.Error(), "VALIDATION_FAILED")

	// Verificamos que contenga el campo correcto
	appErr, ok := err.(*errors.AppError)
	assert.True(t, ok)
	assert.NotNil(t, appErr.Fields)
	assert.Contains(t, appErr.Fields, "fechaBaja")
	assert.Contains(t, appErr.Fields["fechaBaja"], "Inactive member")
}

// Tests de hooks de GORM
func TestBeforeCreate(t *testing.T) {
	member := test.CreateValidMember()

	// Caso válido
	err := member.BeforeCreate(nil)
	assert.NoError(t, err)

	// Caso inválido
	member.NumeroSocio = ""
	err = member.BeforeCreate(nil)
	assert.Error(t, err)
}

func TestBeforeUpdate(t *testing.T) {
	member := test.CreateValidMember()

	// Caso válido
	err := member.BeforeUpdate(nil)
	assert.NoError(t, err)

	// Caso inválido
	member.Nombre = ""
	err = member.BeforeUpdate(nil)
	assert.Error(t, err)
}
