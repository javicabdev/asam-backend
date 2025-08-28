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
		// Enhanced Content Security Policy for production
		// Allows connections to frontend domain and external services
		cspPolicy := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"connect-src 'self' https://mutuaasam.org wss://mutuaasam.org https://fonts.googleapis.com http://localhost:3000 http://localhost:8080; " +
			"img-src 'self' data: https:; " +
			"frame-src 'none'; " +
			"object-src 'none'; " +
			"base-uri 'self';"

		w.Header().Set("Content-Security-Policy", cspPolicy)
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}
