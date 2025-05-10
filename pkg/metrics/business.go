package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Constantes para las métricas de morosidad
const (
	DefaulterBucket30     = "30"
	DefaulterBucket60     = "60"
	DefaulterBucket90     = "90"
	DefaulterBucket90Plus = ">90"
)

var (
	// MembersByStatus métrica que registra el número de miembros por estado
	MembersByStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asam_members_total",
			Help: "Total number of members by status",
		},
		[]string{"status", "type"}, // status: activo/inactivo, type: individual/familiar
	)

	// PaymentMetrics Pagos y Cuotas
	PaymentMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asam_payments_amount",
			Help: "Payment amounts by type and status",
		},
		[]string{"type", "status"}, // type: cuota/otros, status: pending/paid/cancelled
	)

	// PaymentLatency métrica que registra el tiempo de pago
PaymentLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asam_payment_latency_days",
			Help:    "Days between due date and payment date",
			Buckets: []float64{1, 7, 15, 30, 60, 90}, // Buckets relevantes para el negocio
		},
		[]string{"member_type"}, // individual/familiar
	)

	// CashFlowMetrics Flujo de Caja
	CashFlowMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asam_cashflow_amount",
			Help: "Cash flow amounts by operation type",
		},
		[]string{"operation_type"}, // ingreso_cuota, gasto_corriente, entrega_fondo, otros_ingresos
	)

	// DefaulterMetrics Morosidad
	DefaulterMetrics = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "asam_defaulters_total",
			Help: "Number of defaulting members",
		},
		[]string{"days_late"}, // 30, 60, 90, >90
	)
)

// UpdateMemberMetrics actualiza las métricas de miembros
func UpdateMemberMetrics(_, inactive, individualActive, familyActive int) {
	MembersByStatus.WithLabelValues("activo", "individual").Set(float64(individualActive))
	MembersByStatus.WithLabelValues("activo", "familiar").Set(float64(familyActive))
	MembersByStatus.WithLabelValues("inactivo", "total").Set(float64(inactive))
}

// UpdateCashFlowMetrics actualiza las métricas de flujo de caja
func UpdateCashFlowMetrics(totalBalance, income, expenses float64) {
	CashFlowMetrics.WithLabelValues("balance").Set(totalBalance)
	CashFlowMetrics.WithLabelValues("income").Set(income)
	CashFlowMetrics.WithLabelValues("expenses").Set(expenses)
}

// getDefaulterBucket retorna el bucket apropiado según los días de retraso
func getDefaulterBucket(daysLate int) string {
	switch {
	case daysLate <= 30:
		return DefaulterBucket30
	case daysLate <= 60:
		return DefaulterBucket60
	case daysLate <= 90:
		return DefaulterBucket90
	default:
		return DefaulterBucket90Plus
	}
}

// UpdateDefaulterMetrics actualiza las métricas de morosos según los días de retraso
func UpdateDefaulterMetrics(daysLate int, count int) {
	bucket := getDefaulterBucket(daysLate)
	DefaulterMetrics.WithLabelValues(bucket).Set(float64(count))
}
