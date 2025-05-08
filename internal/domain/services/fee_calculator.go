package services

import (
	"math"
	"time"

	"github.com/javicabdev/asam-backend/internal/ports/input"
)

// feeCalculator implementa la interfaz input.FeeCalculator
type feeCalculator struct {
	baseFeeAmount  float64
	familyFeeExtra float64
	lateFeeRate    float64 // porcentaje diario de recargo
	maxLateFee     float64 // porcentaje máximo de recargo
}

// NewFeeCalculator crea una nueva instancia del calculador de cuotas
// Nota: Dado que la interfaz no permite devolver error, aseguramos valores válidos
func NewFeeCalculator(
	baseFeeAmount float64,
	familyFeeExtra float64,
	lateFeeRate float64,
	maxLateFee float64,
) input.FeeCalculator {
	// Garantizar valores válidos
	if baseFeeAmount <= 0 {
		// Valor por defecto seguro (10 unidades)
		baseFeeAmount = 10.0
	}

	// Garantizar que los recargos sean positivos
	if lateFeeRate < 0 {
		lateFeeRate = 0.0
	}

	if maxLateFee < 0 {
		maxLateFee = 0.0
	}

	return &feeCalculator{
		baseFeeAmount:  baseFeeAmount,
		familyFeeExtra: math.Max(0, familyFeeExtra), // Asegurar valor no negativo
		lateFeeRate:    lateFeeRate,
		maxLateFee:     maxLateFee,
	}
}

func (c *feeCalculator) CalculateBaseFee(year, month int) float64 {
	// Validar entrada (año y mes)
	if year < 2000 || month < 1 || month > 12 {
		// Para entradas inválidas, devolver el valor base sin ajustes
		return c.baseFeeAmount
	}

	// Por ahora retornamos el monto base fijo
	// Aquí se podrían implementar ajustes según la fecha
	return c.baseFeeAmount
}

func (c *feeCalculator) CalculateFamilyFee(year, month int) float64 {
	// Validar entrada (año y mes)
	if year < 2000 || month < 1 || month > 12 {
		// Para entradas inválidas, devolver la suma simple
		return c.baseFeeAmount + c.familyFeeExtra
	}

	// La cuota familiar es la cuota base más el extra familiar
	// En el futuro podría incluir lógica para ajustes estacionales, etc.
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
	if baseFee < 0 || daysLate < 0 {
		// Manejar entradas inválidas
		return math.Max(0, baseFee)
	}
	return baseFee + c.CalculateLateFee(daysLate)
}

func (c *feeCalculator) IsLatePayment(dueDate time.Time) bool {
	// Verificar si es una fecha válida (no cero)
	if dueDate.IsZero() {
		return false
	}
	return time.Now().After(dueDate)
}

func (c *feeCalculator) DaysLate(dueDate time.Time) int {
	if dueDate.IsZero() || !c.IsLatePayment(dueDate) {
		return 0
	}
	return int(time.Since(dueDate).Hours() / 24)
}
