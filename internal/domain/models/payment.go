package models

import (
	"time"

	"gorm.io/gorm"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

// PaymentStatus representa el estado de un pago
type PaymentStatus string

const (
	// PaymentStatusPending estado de un pago pendiente
	PaymentStatusPending   PaymentStatus = "pending"
	// PaymentStatusPaid estado de un pago completado
	PaymentStatusPaid      PaymentStatus = "paid"
	// PaymentStatusCancelled estado de un pago cancelado
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment representa un pago realizado por un miembro o familia
type Payment struct {
	gorm.Model
	MemberID        uint
	Member          Member `gorm:"foreignKey:MemberID"`
	FamilyID        *uint
	Family          *Family `gorm:"foreignKey:FamilyID"`
	Amount          float64
	PaymentDate     time.Time
	Status          PaymentStatus
	PaymentMethod   string
	Notes           string
	MembershipFeeID *uint
	MembershipFee   *MembershipFee `gorm:"foreignKey:MembershipFeeID"`
	CashFlow        *CashFlow      `gorm:"foreignKey:PaymentID"` // Relación inversa
}

// MembershipFee representa una cuota de membresía
type MembershipFee struct {
	gorm.Model
	Year           int
	Month          int
	BaseFeeAmount  float64
	FamilyFeeExtra float64 // Additional amount for family memberships
	Status         PaymentStatus
	DueDate        time.Time
	PaymentID      *uint
	Payment        *Payment `gorm:"foreignKey:PaymentID"`
}

// Validate verifica que el pago cumpla con las reglas de negocio
func (p *Payment) Validate() error {
	if p.MemberID == 0 && p.FamilyID == nil {
		return appErrors.NewValidationError(
			"payment must be associated with either a member or family",
			map[string]string{
				"MemberID": "cannot be 0 if FamilyID is nil",
				"FamilyID": "cannot be nil if MemberID is 0",
			},
		)
	}

	if p.Amount <= 0 {
		return appErrors.NewValidationError(
			"payment amount must be greater than 0",
			map[string]string{
				"Amount": "must be > 0",
			},
		)
	}

	return nil
}

// Calculate calcula el monto de la cuota según si es familiar o individual
func (mf *MembershipFee) Calculate(isFamily bool) float64 {
	amount := mf.BaseFeeAmount
	if isFamily {
		amount += mf.FamilyFeeExtra
	}
	return amount
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un registro
func (p *Payment) BeforeCreate(_ *gorm.DB) error {
	return p.Validate()
}

// BeforeUpdate hook de GORM que se ejecuta antes de actualizar un registro
func (p *Payment) BeforeUpdate(_ *gorm.DB) error {
	return p.Validate()
}
