package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal Métrica para total de requests HTTP
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asam_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration Métrica para duración de requests HTTP
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asam_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// DBOperationsTotal Métrica para total de operaciones de BD
	DBOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "asam_db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "table", "status"},
	)

	// DBOperationDuration Métrica para duración de operaciones de BD
	DBOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "asam_db_operation_duration_seconds",
			Help:    "Duration of database operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)
)
