package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Métrica para total de requests HTTP
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asam_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// Métrica para duración de requests HTTP
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asam_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Métrica para total de operaciones de BD
	DBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asam_db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	// Métrica para duración de operaciones de BD
	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asam_db_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Métricas de negocio
	ActiveMembersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "asam_active_members_total",
			Help: "Total number of active members",
		},
	)

	PaymentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asam_payments_total",
			Help: "Total number of payments",
		},
		[]string{"status", "type"},
	)

	CashFlowBalance = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "asam_cashflow_balance",
			Help: "Current cash flow balance",
		},
	)
)
