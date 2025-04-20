package models

import (
	"gorm.io/gorm"
	"time"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// Member representa un miembro individual de ASAM
type Member struct {
	ID                 uint       `gorm:"primaryKey;column:miembro_id"`
	NumeroSocio        string     `gorm:"unique;not null;column:numero_socio"`
	TipoMembresia      string     `gorm:"not null;column:tipo_membresia"`
	Nombre             string     `gorm:"not null"`
	Apellidos          string     `gorm:"not null"`
	CalleNumeroPiso    string     `gorm:"not null;column:calle_numero_piso"`
	CodigoPostal       string     `gorm:"not null;column:codigo_postal"`
	Poblacion          string     `gorm:"not null"`
	Provincia          string     `gorm:"not null;default:Barcelona"`
	Pais               string     `gorm:"not null;default:España"`
	Estado             string     `gorm:"not null"`
	FechaAlta          time.Time  `gorm:"not null;type:date;column:fecha_alta"`
	FechaBaja          *time.Time `gorm:"type:date;column:fecha_baja"`
	FechaNacimiento    *time.Time `gorm:"type:date;column:fecha_nacimiento"`
	DocumentoIdentidad *string    `gorm:"column:documento_identidad"`
	CorreoElectronico  *string    `gorm:"column:correo_electronico"`
	Profesion          *string
	Nacionalidad       string `gorm:"default:Senegal"`
	Observaciones      *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Tipos de membresía disponibles
const (
	TipoMembresiaPIndividual = "individual"
	TipoMembresiaPFamiliar   = "familiar"
)

// Estados posibles de un miembro
const (
	EstadoActivo   = "activo"
	EstadoInactivo = "inactivo"
)

// IsActive retorna true si el miembro está activo
func (m *Member) IsActive() bool {
	return m.Estado == EstadoActivo
}

// IsFamiliar retorna true si el miembro es de tipo familiar
func (m *Member) IsFamiliar() bool {
	return m.TipoMembresia == TipoMembresiaPFamiliar
}

// NombreCompleto retorna el nombre completo del miembro
func (m *Member) NombreCompleto() string {
	return m.Nombre + " " + m.Apellidos
}

// Validate realiza las validaciones de negocio del miembro
func (m *Member) Validate() error {
	if err := m.validateBasicFields(); err != nil {
		return err
	}

	if err := m.validateDates(); err != nil {
		return err
	}

	if err := m.validateStatus(); err != nil {
		return err
	}

	return nil
}

// validateBasicFields valida los campos básicos obligatorios
func (m *Member) validateBasicFields() error {
	errDetails := make(map[string]string)

	// Validamos todos los campos requeridos primero
	if m.NumeroSocio == "" {
		errDetails["numeroSocio"] = "Member number is required"
	}

	if m.TipoMembresia != TipoMembresiaPIndividual && m.TipoMembresia != TipoMembresiaPFamiliar {
		errDetails["tipoMembresia"] = "Must be INDIVIDUAL or FAMILY"
	}

	if m.Nombre == "" {
		errDetails["nombre"] = "Name is required"
	}

	if m.Apellidos == "" {
		errDetails["apellidos"] = "Last name is required"
	}

	if m.CalleNumeroPiso == "" {
		errDetails["calleNumeroPiso"] = "Address is required"
	}

	if m.CodigoPostal == "" {
		errDetails["codigoPostal"] = "Postal code is required"
	}

	if m.Poblacion == "" {
		errDetails["poblacion"] = "City is required"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError("Error de validación en campos del miembro", errDetails)
	}

	return nil
}

// validateDates valida las fechas del miembro
func (m *Member) validateDates() error {
	errDetails := make(map[string]string)

	if m.FechaAlta.IsZero() {
		errDetails["fechaAlta"] = "Registration date is required"
	}

	if m.FechaBaja != nil {
		if m.FechaBaja.Before(m.FechaAlta) || !m.FechaBaja.After(m.FechaAlta) {
			errDetails["fechaBaja"] = "Termination date must be after registration date"
		}
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError("Error de validación en las fechas", errDetails)
	}

	return nil
}

// validateStatus valida el estado del miembro
func (m *Member) validateStatus() error {
	errDetails := make(map[string]string)

	if m.Estado != EstadoActivo && m.Estado != EstadoInactivo {
		errDetails["estado"] = "Status must be 'activo' or 'inactivo'"
	}

	if m.Estado == EstadoInactivo && m.FechaBaja == nil {
		errDetails["fechaBaja"] = "Inactive member must have a termination date"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError("Error de validación en el estado", errDetails)
	}

	return nil
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un miembro
func (m *Member) BeforeCreate(*gorm.DB) error {
	if m.Estado == "" {
		m.Estado = EstadoActivo
	}
	return m.Validate()
}

// BeforeUpdate hook de GORM que se ejecuta antes de actualizar un miembro
func (m *Member) BeforeUpdate(*gorm.DB) error {
	return m.Validate()
}

// TableName especifica el nombre de la tabla en la base de datos
func (m *Member) TableName() string {
	return "miembros"
}
