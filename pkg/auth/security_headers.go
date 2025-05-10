package auth

import (
	"net/http"
)

// SecurityHeadersMiddleware implements HTTP middleware to add security-related
// headers to all responses to improve application security.
type SecurityHeadersMiddleware struct {
	// Podríamos añadir opciones configurables si fuera necesario
}

// NewSecurityHeadersMiddleware creates a new instance of SecurityHeadersMiddleware
// with default settings.
func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

// Middleware returns an HTTP middleware function that adds security headers
// to all HTTP responses to protect against common web vulnerabilities.
func (m *SecurityHeadersMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security Headers
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
