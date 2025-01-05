package models

import (
	"gorm.io/gorm"
	"time"

	appErrors "github.com/javicabdev/asam-backend/pkg/errors"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

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

func (mf *MembershipFee) Calculate(isFamily bool) float64 {
	amount := mf.BaseFeeAmount
	if isFamily {
		amount += mf.FamilyFeeExtra
	}
	return amount
}

func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	return p.Validate()
}

func (p *Payment) BeforeUpdate(tx *gorm.DB) error {
	return p.Validate()
}
