package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
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
		return errors.New("payment must be associated with either a member or family")
	}
	if p.Amount <= 0 {
		return errors.New("payment amount must be greater than 0")
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
