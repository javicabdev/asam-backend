package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Telephone representa un número de teléfono de contacto
type Telephone struct {
	ID             uint   `gorm:"primaryKey"`
	NumeroTelefono string `gorm:"size:20;not null"`

	// Campos para relación polimórfica
	ContactableID   uint   `gorm:"not null"`
	ContactableType string `gorm:"not null"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Validate verifica la validez del número de teléfono
func (t *Telephone) Validate() error {
	if t.NumeroTelefono == "" {
		return errors.New("número de teléfono es requerido")
	}
	// Aquí podrías añadir más validaciones específicas del formato del teléfono
	return nil
}

// BeforeCreate hook de GORM para validaciones antes de crear
func (t *Telephone) BeforeCreate(tx *gorm.DB) error {
	return t.Validate()
}
