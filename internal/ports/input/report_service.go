package input

import (
	"context"
	"time"
)

// ReportService define las operaciones de negocio para reportes
type ReportService interface {
	GetDelinquentReport(ctx context.Context, input DelinquentReportInput) (*DelinquentReportResponse, error)
}

// DelinquentReportInput contiene los parámetros de entrada para el reporte de morosos
type DelinquentReportInput struct {
	CutoffDate  *time.Time
	MinAmount   *float64
	DebtorType  *string
	SortBy      *string
}

// DelinquentReportResponse contiene la respuesta del reporte de morosos
type DelinquentReportResponse struct {
	Debtors     []*Debtor
	Summary     DelinquentSummary
	GeneratedAt time.Time
}

// Debtor representa un deudor (socio o familia)
type Debtor struct {
	MemberID          *uint
	FamilyID          *uint
	Type              string
	Member            *DebtorMemberInfo
	Family            *DebtorFamilyInfo
	PendingPayments   []*PendingPayment
	TotalDebt         float64
	OldestDebtDays    int
	OldestDebtDate    time.Time
	LastPaymentDate   *time.Time
	LastPaymentAmount *float64
}

// DebtorMemberInfo contiene información básica del socio para el informe
type DebtorMemberInfo struct {
	ID           uint
	MemberNumber string
	FirstName    string
	LastName     string
	Email        *string
	Phone        *string
	Status       string
	Membership   string // "INDIVIDUAL" o "FAMILY"
}

// DebtorFamilyInfo contiene información básica de la familia para el informe
type DebtorFamilyInfo struct {
	ID            uint
	FamilyName    string
	PrimaryMember DebtorMemberInfo
	TotalMembers  int
}

// PendingPayment contiene información de un pago pendiente
type PendingPayment struct {
	ID          uint
	Amount      float64
	CreatedAt   time.Time
	DaysOverdue int
	Notes       *string
}

// DelinquentSummary contiene el resumen estadístico del informe
type DelinquentSummary struct {
	TotalDebtors          int
	IndividualDebtors     int
	FamilyDebtors         int
	TotalDebtAmount       float64
	AverageDaysOverdue    int
	AverageDebtPerDebtor  float64
}
