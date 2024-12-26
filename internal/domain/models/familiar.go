package models

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// Familiar representa un familiar (hijo/a u otro dependiente) en una familia ASAM
type Familiar struct {
	ID                uint   `gorm:"primaryKey"`
	FamiliaID         uint   `gorm:"not null;index"`
	Nombre            string `gorm:"size:100;not null"`
	Apellidos         string `gorm:"size:100;not null"`
	FechaNacimiento   *time.Time
	DNINIE            string `gorm:"column:dni_nie;size:20"` // DNI o NIE del familiar
	CorreoElectronico string `gorm:"size:100"`
	Parentesco        string `gorm:"size:50;not null"` // Ejemplo: "Hijo", "Hija", "Otro"

	// Relaciones
	Familia Family `gorm:"foreignKey:FamiliaID"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Validate verifica la validez de los datos del familiar
func (f *Familiar) Validate() error {
	if f.FamiliaID == 0 {
		return errors.New("ID de familia es requerido")
	}

	if f.Nombre == "" {
		return errors.New("nombre es requerido")
	}

	if f.Apellidos == "" {
		return errors.New("apellidos son requeridos")
	}

	if f.Parentesco == "" {
		return errors.New("parentesco es requerido")
	}

	return nil
}

// BeforeCreate hook de GORM para validaciones antes de crear
func (f *Familiar) BeforeCreate(tx *gorm.DB) error {
	return f.Validate()
}

// BeforeSave hook de GORM para validaciones antes de guardar
func (f *Familiar) BeforeSave(tx *gorm.DB) error {
	return f.Validate()
}
