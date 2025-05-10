package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/javicabdev/asam-backend/pkg/errors"
)

// Tipo para keys de contexto
type validationContextKey string

// Clave para sanitización
const sanitizeKey validationContextKey = "sanitize"

type ValidationMiddleware struct {
	next http.Handler
}

func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

func (m *ValidationMiddleware) Handler(next http.Handler) http.Handler {
	m.next = next
	return m
}

func (m *ValidationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Validar method HTTP
	if r.Method != http.MethodPost && r.Method != http.MethodOptions {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 2. Validar Content-Type
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
	}

	// 3. Validar tamaño del request
	if r.ContentLength > 1024*1024 { // 1MB máximo
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	// 4. Validar caracteres especiales en inputs (sanitización básica)
	r = r.WithContext(context.WithValue(r.Context(), sanitizeKey, true))

	// Continuar con el siguiente handler
	m.next.ServeHTTP(w, r)
}

// ValidateAndSanitize función helper para sanitizar inputs
func ValidateAndSanitize(input string) (string, error) {
	// Validar longitud
	if len(input) > 255 {
		return "", errors.NewValidationError("Input too long", map[string]string{
			"max_length": "255",
		})
	}

	// Remover caracteres peligrosos
	sanitized := strings.Map(func(r rune) rune {
		switch {
		case r == '<', r == '>', r == '\'', r == '"', r == ';':
			return -1 // eliminar estos caracteres
		default:
			return r
		}
	}, input)

	return sanitized, nil
}
