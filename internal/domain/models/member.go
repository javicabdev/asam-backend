package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
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
	if m.NumeroSocio == "" {
		return errors.New("el número de socio es obligatorio")
	}

	if m.TipoMembresia != TipoMembresiaPIndividual && m.TipoMembresia != TipoMembresiaPFamiliar {
		return errors.New("tipo de membresía no válido")
	}

	if m.Nombre == "" {
		return errors.New("el nombre es obligatorio")
	}

	if m.Apellidos == "" {
		return errors.New("los apellidos son obligatorios")
	}

	if m.CalleNumeroPiso == "" {
		return errors.New("la dirección es obligatoria")
	}

	if m.CodigoPostal == "" {
		return errors.New("el código postal es obligatorio")
	}

	if m.Poblacion == "" {
		return errors.New("la población es obligatoria")
	}

	return nil
}

// validateDates valida las fechas del miembro
func (m *Member) validateDates() error {
	if m.FechaAlta.IsZero() {
		return errors.New("la fecha de alta es obligatoria")
	}

	if m.FechaBaja != nil {
		if m.FechaBaja.Before(m.FechaAlta) {
			return errors.New("la fecha de baja no puede ser anterior a la fecha de alta")
		}
		if !m.FechaBaja.After(m.FechaAlta) {
			return errors.New("la fecha de baja debe ser posterior a la fecha de alta")
		}
	}

	return nil
}

// validateStatus valida el estado del miembro
func (m *Member) validateStatus() error {
	if m.Estado != EstadoActivo && m.Estado != EstadoInactivo {
		return errors.New("estado no válido")
	}

	if m.Estado == EstadoInactivo && m.FechaBaja == nil {
		return errors.New("un miembro inactivo debe tener fecha de baja")
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
