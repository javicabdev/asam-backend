package models

// OperationType define los tipos de operaciones permitidas en el flujo de caja
type OperationType string

const (
	// OperationTypeMembershipFee representa un ingreso por cuota de membresía
	OperationTypeMembershipFee OperationType = "ingreso_cuota"
	// OperationTypeCurrentExpense representa un gasto corriente de la asociación
	OperationTypeCurrentExpense OperationType = "gasto_corriente"
	// OperationTypeFundDelivery representa una entrega de fondos (ej.: por fallecimiento)
	OperationTypeFundDelivery OperationType = "entrega_fondo"
	// OperationTypeOtherIncome representa otros ingresos no categorizados
	OperationTypeOtherIncome OperationType = "otros_ingresos"
)

// IsValid verifica si el tipo de operación es válido
func (ot OperationType) IsValid() bool {
	switch ot {
	case OperationTypeMembershipFee,
		OperationTypeCurrentExpense,
		OperationTypeFundDelivery,
		OperationTypeOtherIncome:
		return true
	}
	return false
}

// IsIncome indica si el tipo de operación es un ingreso
func (ot OperationType) IsIncome() bool {
	return ot == OperationTypeMembershipFee || ot == OperationTypeOtherIncome
}

// IsExpense indica si el tipo de operación es un gasto
func (ot OperationType) IsExpense() bool {
	return ot == OperationTypeCurrentExpense || ot == OperationTypeFundDelivery
}
