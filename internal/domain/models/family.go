package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/javicabdev/asam-backend/internal/domain/services/validation"
)

// Family representa una familia en el sistema ASAM
//
// Convención de numeración:
// - NumeroSocio debe comenzar con 'A' para familias (ej: A00001)
// - Los miembros individuales asociados a esta familia también usan prefijo 'A'
type Family struct {
	ID              uint   `gorm:"primaryKey"`
	NumeroSocio     string `gorm:"unique;not null"` // Formato: AXXXXX
	MiembroOrigenID *uint  `gorm:"index"`           // Referencia al miembro que origina la familia
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
	MiembroOrigen *Member     `gorm:"foreignKey:MiembroOrigenID"`
	Familiares    []Familiar  `gorm:"foreignKey:FamiliaID"`
	Telefonos     []Telephone `gorm:"polymorphic:Contactable"`

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

	// Validar datos de cónyuges (flexible: esposa opcional)
	if err := validator.ValidateConyugesFlexible(
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

// BeforeCreate hook de GORM para transformaciones antes de crear
// NOTA: Las validaciones deben ocurrir ANTES de la transacción en la capa de servicio
// Este hook solo normaliza datos para evitar problemas durante la transacción
func (f *Family) BeforeCreate(_ *gorm.DB) error {
	// Normalizar documentos de identidad si se proporcionan (DNI, NIE, pasaporte, etc.)
	if f.EsposoDocumentoIdentidad != "" {
		f.EsposoDocumentoIdentidad = validation.NormalizeIdentityDocument(f.EsposoDocumentoIdentidad)
	}
	if f.EsposaDocumentoIdentidad != "" {
		f.EsposaDocumentoIdentidad = validation.NormalizeIdentityDocument(f.EsposaDocumentoIdentidad)
	}
	// No validamos aquí para evitar fallos dentro de transacciones
	// La validación se hace en validateFamilyAtomicRequest() antes de iniciar la transacción
	return nil
}

// BeforeUpdate hook de GORM para transformaciones antes de actualizar
// NOTA: Las validaciones deben ocurrir ANTES de la transacción en la capa de servicio
// Este hook solo normaliza datos para evitar problemas durante la transacción
func (f *Family) BeforeUpdate(_ *gorm.DB) error {
	// Normalizar documentos de identidad si se proporcionan (DNI, NIE, pasaporte, etc.)
	if f.EsposoDocumentoIdentidad != "" {
		f.EsposoDocumentoIdentidad = validation.NormalizeIdentityDocument(f.EsposoDocumentoIdentidad)
	}
	if f.EsposaDocumentoIdentidad != "" {
		f.EsposaDocumentoIdentidad = validation.NormalizeIdentityDocument(f.EsposaDocumentoIdentidad)
	}
	// No validamos aquí para evitar fallos dentro de transacciones
	// La validación debe hacerse en la capa de servicio antes de la transacción
	return nil
}

// NombreCompletoEsposo retorna el nombre completo del esposo
func (f *Family) NombreCompletoEsposo() string {
	return fmt.Sprintf("%s %s", f.EsposoNombre, f.EsposoApellidos)
}

// NombreCompletoEsposa retorna el nombre completo de la esposa
func (f *Family) NombreCompletoEsposa() string {
	return fmt.Sprintf("%s %s", f.EsposaNombre, f.EsposaApellidos)
}
