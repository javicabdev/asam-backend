package services

import (
	"github.com/javicabdev/asam-backend/internal/ports/input"
	"math"
	"time"
)

// feeCalculator implementa la interfaz input.FeeCalculator
type feeCalculator struct {
	baseFeeAmount  float64
	familyFeeExtra float64
	lateFeeRate    float64 // porcentaje diario de recargo
	maxLateFee     float64 // porcentaje máximo de recargo
}

func NewFeeCalculator(
	baseFeeAmount float64,
	familyFeeExtra float64,
	lateFeeRate float64,
	maxLateFee float64,
) input.FeeCalculator {
	return &feeCalculator{
		baseFeeAmount:  baseFeeAmount,
		familyFeeExtra: familyFeeExtra,
		lateFeeRate:    lateFeeRate,
		maxLateFee:     maxLateFee,
	}
}

func (c *feeCalculator) CalculateBaseFee(year, month int) float64 {
	// Por ahora retornamos el monto base fijo
	// Aquí se podrían implementar ajustes según la fecha
	return c.baseFeeAmount
}

func (c *feeCalculator) CalculateFamilyFee(year, month int) float64 {
	// La cuota familiar es la cuota base más el extra familiar
	return c.baseFeeAmount + c.familyFeeExtra
}

func (c *feeCalculator) CalculateLateFee(daysLate int) float64 {
	if daysLate <= 0 {
		return 0
	}

	// Calcular el recargo como porcentaje de la cuota base
	lateFeePercentage := float64(daysLate) * c.lateFeeRate

	// Limitar al máximo permitido
	lateFeePercentage = math.Min(lateFeePercentage, c.maxLateFee)

	return c.baseFeeAmount * (lateFeePercentage / 100)
}

// Métodos auxiliares que podrían ser útiles

func (c *feeCalculator) CalculateTotalWithLateFee(baseFee float64, daysLate int) float64 {
	return baseFee + c.CalculateLateFee(daysLate)
}

func (c *feeCalculator) IsLatePayment(dueDate time.Time) bool {
	return time.Now().After(dueDate)
}

func (c *feeCalculator) DaysLate(dueDate time.Time) int {
	if !c.IsLatePayment(dueDate) {
		return 0
	}
	return int(time.Since(dueDate).Hours() / 24)
}
