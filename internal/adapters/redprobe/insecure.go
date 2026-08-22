// Package redprobe is a THROWAWAY tripwire to prove the gosec gate can fail.
// It must never be merged. See PR verification for issue #142.
package redprobe

import (
	"crypto/tls"
	"net/http"
)

// InsecureClient returns an HTTP client with TLS verification disabled.
// This deliberately triggers gosec G402 (HIGH severity, HIGH confidence).
func InsecureClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}
