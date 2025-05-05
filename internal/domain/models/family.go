package models

import (
	"fmt"
	"time"

	"github.com/javicabdev/asam-backend/internal/domain/services/validation"
	"gorm.io/gorm"
)

// Family representa una familia en el sistema ASAM
type Family struct {
	ID              uint   `gorm:"primaryKey"`
	NumeroSocio     string `gorm:"unique;not null"`
	MiembroOrigenID *uint  `gorm:"index"` // Referencia al miembro que origina la familia
	EsposoNombre    string `gorm:"size:100"`
	EsposoApellidos string `gorm:"size:100"`
	EsposaNombre    string `gorm:"size:100"`
	EsposaApellidos string `gorm:"size:100"`

	// Datos adicionales
	EsposoFechaNacimiento    *time.Time
	EsposoDocumentoIdentidad string `gorm:"size:20"`
	EsposoCorreoElectronico  string `gorm:"size:100"`
	EsposaFechaNacimiento    *time.Time
	EsposaDocumentoIdentidad string `gorm:"size:20"`
	EsposaCorreoElectronico  string `gorm:"size:100"`

	// Relaciones
	MiembroOrigen *Member    `gorm:"foreignKey:MiembroOrigenID"`
	Familiares    []Familiar `gorm:"foreignKey:FamiliaID"`
	Telefonos     []Telefono `gorm:"polymorphic:Contactable"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// validator es una instancia del validador que usaremos
var validator = validation.NewFamilyValidator()

// Validate verifica la validez de los datos de la familia
func (f *Family) Validate() error {
	// Validar número de socio
	if err := validator.ValidateNumeroSocio(f.NumeroSocio); err != nil {
		return err
	}

	// Validar datos de cónyuges
	if err := validator.ValidateConyuges(
		f.EsposoNombre,
		f.EsposoApellidos,
		f.EsposaNombre,
		f.EsposaApellidos,
	); err != nil {
		return err
	}

	// Validar documentos de identidad
	if err := validator.ValidateDocumentIDs(
		f.EsposoDocumentoIdentidad,
		f.EsposaDocumentoIdentidad,
	); err != nil {
		return err
	}

	// Validar información de contacto
	if err := validator.ValidateContactInfo(
		f.EsposoCorreoElectronico,
		f.EsposaCorreoElectronico,
	); err != nil {
		return err
	}

	// Validar fechas
	if err := validator.ValidateDates(
		f.EsposoFechaNacimiento,
		f.EsposaFechaNacimiento,
	); err != nil {
		return err
	}

	return nil
}

// BeforeCreate hook de GORM para validaciones antes de crear
func (f *Family) BeforeCreate(tx *gorm.DB) error {
	return f.Validate()
}

// BeforeUpdate hook de GORM para validaciones antes de actualizar
func (f *Family) BeforeUpdate(tx *gorm.DB) error {
	return f.Validate()
}

// NombreCompletoEsposo retorna el nombre completo del esposo
func (f *Family) NombreCompletoEsposo() string {
	return fmt.Sprintf("%s %s", f.EsposoNombre, f.EsposoApellidos)
}

// NombreCompletoEsposa retorna el nombre completo de la esposa
func (f *Family) NombreCompletoEsposa() string {
	return fmt.Sprintf("%s %s", f.EsposaNombre, f.EsposaApellidos)
}
