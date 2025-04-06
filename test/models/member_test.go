package models_test

import (
	"github.com/javicabdev/asam-backend/test"
	"testing"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/models"
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
	assert.Contains(t, err.Error(), "El número de socio es obligatorio")
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
	assert.Contains(t, err.Error(), "la fecha de baja no puede ser anterior a la fecha de alta")

	// Caso: FechaBaja igual a FechaAlta
	member.FechaBaja = &member.FechaAlta
	err = member.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "la fecha de baja debe ser posterior a la fecha de alta")
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
	assert.Equal(t, "Juan Pérez", member.NombreCompleto())
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
	assert.Contains(t, err.Error(), "un miembro inactivo debe tener fecha de baja")
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
