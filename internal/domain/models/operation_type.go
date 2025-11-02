package models

// OperationType define los tipos de operaciones permitidas en el flujo de caja
type OperationType string

const (
	// INGRESOS (amount > 0)

	// OperationTypeMembershipFee representa un ingreso por cuota de membresía (generado automáticamente)
	OperationTypeMembershipFee OperationType = "INGRESO_CUOTA"
	// OperationTypeDonation representa un ingreso por donación (registro manual)
	OperationTypeDonation OperationType = "INGRESO_DONACION"
	// OperationTypeOtherIncome representa otros ingresos no categorizados (registro manual)
	OperationTypeOtherIncome OperationType = "INGRESO_OTRO"

	// GASTOS (amount < 0)

	// OperationTypeRepatriation representa un gasto de repatriación (requiere member_id)
	OperationTypeRepatriation OperationType = "GASTO_REPATRIACION"
	// OperationTypeAdministrative representa gastos administrativos (tasas, sellos, copistería)
	OperationTypeAdministrative OperationType = "GASTO_ADMINISTRATIVO"
	// OperationTypeBankFees representa gastos bancarios (comisiones)
	OperationTypeBankFees OperationType = "GASTO_BANCARIO"
	// OperationTypeSocialAid representa ayudas sociales
	OperationTypeSocialAid OperationType = "GASTO_AYUDA"
	// OperationTypeOtherExpense representa otros gastos no categorizados
	OperationTypeOtherExpense OperationType = "GASTO_OTRO"
)

// IsValid verifica si el tipo de operación es válido
func (ot OperationType) IsValid() bool {
	switch ot {
	case OperationTypeMembershipFee,
		OperationTypeDonation,
		OperationTypeOtherIncome,
		OperationTypeRepatriation,
		OperationTypeAdministrative,
		OperationTypeBankFees,
		OperationTypeSocialAid,
		OperationTypeOtherExpense:
		return true
	}
	return false
}

// IsIncome indica si el tipo de operación es un ingreso
func (ot OperationType) IsIncome() bool {
	return ot == OperationTypeMembershipFee ||
		ot == OperationTypeDonation ||
		ot == OperationTypeOtherIncome
}

// IsExpense indica si el tipo de operación es un gasto
func (ot OperationType) IsExpense() bool {
	return ot == OperationTypeRepatriation ||
		ot == OperationTypeAdministrative ||
		ot == OperationTypeBankFees ||
		ot == OperationTypeSocialAid ||
		ot == OperationTypeOtherExpense
}
