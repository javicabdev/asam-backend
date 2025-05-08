package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// Helper para crear un User válido
func createValidUser() *models.User {
	return &models.User{
		Username: "johndoe",
		Password: "password123", // Esto será hasheado por SetPassword
		Role:     models.RoleUser,
	}
}

// Tests de lógica de negocio
func TestUser_SetPassword(t *testing.T) {
	user := createValidUser()

	// Probar que la contraseña se hashea correctamente
	err := user.SetPassword("password123")
	assert.NoError(t, err)
	assert.NotEqual(t, "password123", user.Password) // La contraseña no debe estar en texto plano
	assert.NotEmpty(t, user.Password)
}

func TestUser_CheckPassword(t *testing.T) {
	user := createValidUser()
	_ = user.SetPassword("password123") // Hashear la contraseña

	// Contraseña correcta
	assert.True(t, user.CheckPassword("password123"))

	// Contraseña incorrecta
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUser_IsAdmin(t *testing.T) {
	user := createValidUser()

	// Caso: No es admin
	assert.False(t, user.IsAdmin())

	// Caso: Es admin
	user.Role = models.RoleAdmin
	assert.True(t, user.IsAdmin())
}

// Tests de hooks de GORM
func TestUser_BeforeCreate(t *testing.T) {
	user := createValidUser()

	// Caso: Rol no definido
	user.Role = ""
	assert.NoError(t, user.BeforeCreate(nil))
	assert.Equal(t, models.RoleUser, user.Role) // Rol por defecto debe ser 'user'

	// Caso: Rol definido
	user.Role = models.RoleAdmin
	assert.NoError(t, user.BeforeCreate(nil))
	assert.Equal(t, models.RoleAdmin, user.Role)
}
