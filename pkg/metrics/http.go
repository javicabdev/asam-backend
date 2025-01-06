package metrics

import (
	"net/http"
	"strconv"
	"time"
)

type MetricsMiddleware struct {
	next http.Handler
}

func NewMetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &MetricsMiddleware{next: next}
	}
}

func (m *MetricsMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Wrap ResponseWriter to capture status code
	wrapped := wrapResponseWriter(w)

	// Call next handler
	m.next.ServeHTTP(wrapped, r)

	// Record metrics
	duration := time.Since(start).Seconds()

	HttpRequestsTotal.WithLabelValues(
		r.Method,
		r.URL.Path,
		strconv.Itoa(wrapped.status),
	).Inc()

	HttpRequestDuration.WithLabelValues(
		r.Method,
		r.URL.Path,
	).Observe(duration)
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
