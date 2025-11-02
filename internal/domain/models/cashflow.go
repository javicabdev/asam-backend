// Package models defines the domain models for the ASAM backend.
// It contains the core business entities and their validation logic.
package models

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrInvalidOperationType error que se produce cuando el tipo de operación no es válido
	ErrInvalidOperationType = errors.New("tipo de operación inválido")
	// ErrInvalidAmount error que se produce cuando el monto no es válido
	ErrInvalidAmount = errors.New("monto inválido")
	// ErrInvalidDate error que se produce cuando la fecha no es válida
	ErrInvalidDate = errors.New("fecha inválida")
	// ErrMissingDetail error que se produce cuando falta el detalle
	ErrMissingDetail = errors.New("detalle requerido")
)

// CashFlow representa un movimiento en el flujo de caja de la asociación.
type CashFlow struct {
	ID            uint           `gorm:"primaryKey"`
	MemberID      *uint          `gorm:"index"`
	PaymentID     *uint          `gorm:"index"` // Referencia al pago que generó este movimiento
	OperationType OperationType  `gorm:"type:varchar(20);not null"`
	Amount        float64        `gorm:"type:decimal(10,2);not null"`
	Date          time.Time      `gorm:"type:timestamp;not null"`
	Detail        string         `gorm:"type:varchar(255)"`
	CreatedAt     time.Time      `gorm:"not null"`
	UpdatedAt     time.Time      `gorm:"not null"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	// Relaciones
	Member  *Member  `gorm:"foreignKey:MemberID"`
	Payment *Payment `gorm:"foreignKey:PaymentID"` // Relación con el pago
}

// BeforeCreate hook de GORM que se ejecuta antes de crear un registro.
func (cf *CashFlow) BeforeCreate(_ *gorm.DB) error {
	return cf.Validate()
}

// BeforeUpdate hook de GORM que se ejecuta antes de actualizar un registro.
func (cf *CashFlow) BeforeUpdate(_ *gorm.DB) error {
	return cf.Validate()
}

// Validate realiza todas las validaciones del modelo
func (cf *CashFlow) Validate() error {
	if !cf.OperationType.IsValid() {
		return ErrInvalidOperationType
	}

	if cf.Amount == 0 {
		return ErrInvalidAmount
	}

	if cf.Date.IsZero() {
		return ErrInvalidDate
	}

	if cf.Detail == "" {
		return ErrMissingDetail
	}

	// Validar que el monto sea positivo para ingresos y negativo para gastos
	if cf.OperationType.IsIncome() && cf.Amount < 0 {
		return errors.New("los ingresos deben tener monto positivo")
	}
	if cf.OperationType.IsExpense() && cf.Amount > 0 {
		return errors.New("los gastos deben tener monto negativo")
	}

	return nil
}

// NewFromPayment crea un nuevo movimiento de caja a partir de un pago
func NewFromPayment(payment *Payment) (*CashFlow, error) {
	if payment == nil {
		return nil, errors.New("payment no puede ser nil")
	}

	if payment.PaymentDate == nil {
		return nil, errors.New("payment date cannot be nil for cash flow creation")
	}

	paymentID := payment.ID

	// Determinar detalle según si es cuota anual
	detail := "Pago registrado"
	if payment.MembershipFee != nil {
		detail = fmt.Sprintf("Cuota anual %d", payment.MembershipFee.Year)
	} else if payment.Notes != "" {
		detail = payment.Notes
	}

	cashFlow := &CashFlow{
		MemberID:      &payment.MemberID,
		PaymentID:     &paymentID,
		OperationType: OperationTypeMembershipFee,
		Amount:        payment.Amount,
		Date:          *payment.PaymentDate,
		Detail:        detail,
	}

	return cashFlow, nil
}
