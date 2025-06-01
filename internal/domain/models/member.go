package models

import (
	"time"

	"gorm.io/gorm"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// Member representa un miembro individual de ASAM
type Member struct {
	ID               uint       `gorm:"primaryKey"`
	MembershipNumber string     `gorm:"unique;not null"`
	MembershipType   string     `gorm:"not null"`
	Name             string     `gorm:"not null"`
	Surnames         string     `gorm:"not null"`
	Address          string     `gorm:"not null"`
	Postcode         string     `gorm:"not null"`
	City             string     `gorm:"not null"`
	Province         string     `gorm:"not null;default:Barcelona"`
	Country          string     `gorm:"not null;default:España"`
	State            string     `gorm:"not null"`
	RegistrationDate time.Time  `gorm:"not null;type:date"`
	LeavingDate      *time.Time `gorm:"type:date"`
	BirthDate        *time.Time `gorm:"type:date"`
	IdentityCard     *string
	Email            *string
	Profession       *string
	Nationality      string `gorm:"default:Senegal"`
	Remarks          *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Tipos de membresía disponibles
const (
	TipoMembresiaPIndividual = "individual"
	TipoMembresiaPFamiliar   = "familiar"
)

// Estados posibles de un miembro
const (
	EstadoActivo   = "active"
	EstadoInactivo = "inactive"
)

// IsActive retorna true si el miembro está activo
func (m *Member) IsActive() bool {
	return m.State == EstadoActivo
}

// IsFamiliar retorna true si el miembro es de tipo familiar
func (m *Member) IsFamiliar() bool {
	return m.MembershipType == TipoMembresiaPFamiliar
}

// NombreCompleto retorna el nombre completo del miembro
func (m *Member) NombreCompleto() string {
	return m.Name + " " + m.Surnames
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
	if m.MembershipNumber == "" {
		errDetails["membershipNumber"] = "Member number is required"
	}

	if m.MembershipType != TipoMembresiaPIndividual && m.MembershipType != TipoMembresiaPFamiliar {
		errDetails["membershipType"] = "Must be INDIVIDUAL or FAMILY"
	}

	if m.Name == "" {
		errDetails["name"] = "Name is required"
	}

	if m.Surnames == "" {
		errDetails["surnames"] = "Last name is required"
	}

	if m.Address == "" {
		errDetails["address"] = "Address is required"
	}

	if m.Postcode == "" {
		errDetails["postcode"] = "Postal code is required"
	}

	if m.City == "" {
		errDetails["city"] = "City is required"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError("Error de validación en campos del miembro", errDetails)
	}

	return nil
}

// validateDates valida las fechas del miembro
func (m *Member) validateDates() error {
	errDetails := make(map[string]string)

	if m.RegistrationDate.IsZero() {
		errDetails["registrationDate"] = "Registration date is required"
	}

	if m.LeavingDate != nil {
		if m.LeavingDate.Before(m.RegistrationDate) || !m.LeavingDate.After(m.RegistrationDate) {
			errDetails["leavingDate"] = "Termination date must be after registration date"
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

	if m.State != EstadoActivo && m.State != EstadoInactivo {
		errDetails["state"] = "Status must be 'active' or 'inactive'"
	}

	if m.State == EstadoInactivo && m.LeavingDate == nil {
		errDetails["leavingDate"] = "Inactive member must have a termination date"
	}

	if len(errDetails) > 0 {
		return appErrors.NewValidationError("Error de validación en el estado", errDetails)
	}

	return nil
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un miembro
func (m *Member) BeforeCreate(*gorm.DB) error {
	if m.State == "" {
		m.State = EstadoActivo
	}
	return m.Validate()
}

// BeforeUpdate hook de GORM que se ejecuta antes de actualizar un miembro
func (m *Member) BeforeUpdate(*gorm.DB) error {
	return m.Validate()
}
