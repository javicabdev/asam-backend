package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/javicabdev/asam-backend/pkg/constants"
)

// clientInfoMiddleware captures client information (IP, User-Agent, Device Name)
// and adds it to the request context before other handlers process it.
func clientInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract real client IP
		clientIP := getClientIP(r)

		// Extract User-Agent
		userAgent := r.Header.Get("User-Agent")

		// Determine device name from User-Agent
		deviceName := getDeviceName(userAgent)

		// Add information to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, constants.IPContextKey, clientIP)
		ctx = context.WithValue(ctx, constants.UserAgentContextKey, userAgent)
		ctx = context.WithValue(ctx, constants.DeviceNameContextKey, deviceName)

		// Continue with the enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClientIP extracts the real client IP considering proxies
func getClientIP(r *http.Request) string {
	// Check proxy headers in priority order
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// If no proxy headers, use RemoteAddr
	// RemoteAddr may include port, so we separate it
	ip := r.RemoteAddr
	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		ip = ip[:colonIdx]
	}

	return ip
}

// getDeviceName attempts to determine device type from User-Agent
func getDeviceName(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// Detect mobile devices
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") {
		if strings.Contains(ua, "tablet") {
			return "Android Tablet"
		}
		return "Android Mobile"
	}

	if strings.Contains(ua, "iphone") {
		return "iPhone"
	}

	if strings.Contains(ua, "ipad") {
		return "iPad"
	}

	// Detect desktop browsers
	if strings.Contains(ua, "windows") {
		return "Windows Desktop"
	}

	if strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os") {
		return "Mac Desktop"
	}

	if strings.Contains(ua, "linux") {
		return "Linux Desktop"
	}

	// Default
	return "Web Browser"
}
