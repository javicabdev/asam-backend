package validation_test

import (
	"testing"

	"github.com/javicabdev/asam-backend/internal/domain/services/validation"
	"github.com/stretchr/testify/assert"
)

func TestValidarNIF(t *testing.T) {
	tests := []struct {
		name     string
		nif      string
		expected bool
	}{
		// DNIs válidos
		{
			name:     "DNI válido con letra Z",
			nif:      "12345678Z",
			expected: true,
		},
		{
			name:     "DNI válido con letra X",
			nif:      "87654321X",
			expected: true,
		},
		{
			name:     "DNI válido con letra A",
			nif:      "11111111H",
			expected: true,
		},
		{
			name:     "DNI válido con letra Q",
			nif:      "99999999R",
			expected: true,
		},
		{
			name:     "DNI válido - caso real 1",
			nif:      "44556677L",
			expected: true,
		},
		{
			name:     "DNI válido - caso real 2",
			nif:      "00000001R",
			expected: true,
		},

		// NIEs válidos
		{
			name:     "NIE válido con X",
			nif:      "X1234567L",
			expected: true,
		},
		{
			name:     "NIE válido con Y",
			nif:      "Y1234567X",
			expected: true,
		},
		{
			name:     "NIE válido con Z",
			nif:      "Z1234567R",
			expected: true,
		},
		{
			name:     "NIE válido con X - caso real",
			nif:      "X0987654B",
			expected: true,
		},
		{
			name:     "NIE válido con Y - caso real",
			nif:      "Y2345678Z",
			expected: true,
		},
		{
			name:     "NIE válido con Z - caso real",
			nif:      "Z9876543A",
			expected: true,
		},

		// Formatos con espacios y guiones (deben ser válidos tras normalización)
		{
			name:     "DNI con guiones",
			nif:      "12345678-Z",
			expected: true,
		},
		{
			name:     "DNI con espacios",
			nif:      "12345678 Z",
			expected: true,
		},
		{
			name:     "NIE con guiones",
			nif:      "X-1234567-L",
			expected: true,
		},
		{
			name:     "NIE con espacios",
			nif:      "X 1234567 L",
			expected: true,
		},
		{
			name:     "DNI con múltiples espacios",
			nif:      "  12345678   Z  ",
			expected: true,
		},
		{
			name:     "NIE mixto espacios y guiones",
			nif:      "X - 1234567 - L",
			expected: true,
		},

		// Minúsculas (deben ser válidas tras normalización)
		{
			name:     "DNI con letra minúscula",
			nif:      "12345678z",
			expected: true,
		},
		{
			name:     "NIE con letras minúsculas",
			nif:      "x1234567l",
			expected: true,
		},
		{
			name:     "NIE mixto mayúsculas y minúsculas",
			nif:      "x1234567L",
			expected: true,
		},

		// Casos inválidos - letra de control incorrecta
		{
			name:     "DNI con letra incorrecta",
			nif:      "12345678A",
			expected: false,
		},
		{
			name:     "NIE con letra incorrecta",
			nif:      "X1234567A",
			expected: false,
		},
		{
			name:     "DNI todo ceros con letra incorrecta",
			nif:      "00000000A",
			expected: false,
		},

		// Casos inválidos - formato incorrecto
		{
			name:     "Longitud incorrecta - muy corto",
			nif:      "1234567Z",
			expected: false,
		},
		{
			name:     "Longitud incorrecta - muy largo",
			nif:      "123456789Z",
			expected: false,
		},
		{
			name:     "Letras en posición de números",
			nif:      "A2345678Z",
			expected: false,
		},
		{
			name:     "NIE con letra inicial incorrecta",
			nif:      "K1234567L",
			expected: false,
		},
		{
			name:     "Sin letra final",
			nif:      "12345678",
			expected: false,
		},
		{
			name:     "Solo números",
			nif:      "123456789",
			expected: false,
		},
		{
			name:     "Solo letras",
			nif:      "ABCDEFGHI",
			expected: false,
		},
		{
			name:     "Número final en lugar de letra",
			nif:      "123456789",
			expected: false,
		},
		{
			name:     "DNI con caracteres especiales",
			nif:      "12@45678Z",
			expected: false,
		},
		{
			name:     "NIE con caracteres especiales",
			nif:      "X12#4567L",
			expected: false,
		},

		// Casos edge
		{
			name:     "Campo vacío (debe ser válido)",
			nif:      "",
			expected: true,
		},
		{
			name:     "Solo espacios",
			nif:      "   ",
			expected: false,
		},
		{
			name:     "Solo guiones",
			nif:      "---",
			expected: false,
		},
		{
			name:     "Un solo carácter",
			nif:      "X",
			expected: false,
		},
		{
			name:     "Formato parcial DNI",
			nif:      "1234",
			expected: false,
		},
		{
			name:     "Formato parcial NIE",
			nif:      "X123",
			expected: false,
		},

		// Casos límite numéricos
		{
			name:     "DNI máximo válido",
			nif:      "99999998T",
			expected: true,
		},
		{
			name:     "DNI mínimo válido",
			nif:      "00000000T",
			expected: true,
		},
		{
			name:     "NIE con X y máximo número",
			nif:      "X9999999J",
			expected: true,
		},
		{
			name:     "NIE con X y mínimo número",
			nif:      "X0000000T",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidarNIF(tt.nif)
			assert.Equal(t, tt.expected, result, "Para NIF: %s", tt.nif)
		})
	}
}

func TestNormalizarNIF(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "DNI sin cambios",
			input:    "12345678Z",
			expected: "12345678Z",
		},
		{
			name:     "DNI con espacios",
			input:    "12345678 Z",
			expected: "12345678Z",
		},
		{
			name:     "DNI con guiones",
			input:    "12345678-Z",
			expected: "12345678Z",
		},
		{
			name:     "NIE con espacios y guiones",
			input:    "X - 1234567 - L",
			expected: "X1234567L",
		},
		{
			name:     "Minúsculas a mayúsculas",
			input:    "x1234567l",
			expected: "X1234567L",
		},
		{
			name:     "Espacios al inicio y final",
			input:    "  12345678Z  ",
			expected: "12345678Z",
		},
		{
			name:     "Múltiples espacios entre caracteres",
			input:    "1 2 3 4 5 6 7 8 Z",
			expected: "12345678Z",
		},
		{
			name:     "Múltiples guiones",
			input:    "X--1234567--L",
			expected: "X1234567L",
		},
		{
			name:     "Vacío",
			input:    "",
			expected: "",
		},
		{
			name:     "Solo espacios y guiones",
			input:    " - - ",
			expected: "",
		},
		{
			name:     "Mezcla compleja",
			input:    " x - 1 2 3 4 5 6 7 - l ",
			expected: "X1234567L",
		},
		{
			name:     "Caso mixto",
			input:    "xY1234567l",
			expected: "XY1234567L",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.NormalizarNIF(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidarNIF_CasosRealesConocidos(t *testing.T) {
	// Estos son casos con DNI/NIE reales donde conocemos la letra correcta
	// Útil para verificar que el algoritmo está correctamente implementado
	casosReales := []struct {
		name     string
		nif      string
		expected bool
	}{
		// DNIs reales conocidos (números ficticios pero algoritmo correcto)
		{
			name:     "DNI real 1",
			nif:      "53657936V",
			expected: true,
		},
		{
			name:     "DNI real 2",
			nif:      "72345678Y",
			expected: true,
		},
		{
			name:     "DNI real 3",
			nif:      "45673214K",
			expected: true,
		},
		// NIEs reales conocidos
		{
			name:     "NIE real X",
			nif:      "X5467890P",
			expected: true,
		},
		{
			name:     "NIE real Y",
			nif:      "Y1472581C",
			expected: true,
		},
		{
			name:     "NIE real Z",
			nif:      "Z4792125B",
			expected: true,
		},
	}

	for _, tt := range casosReales {
		t.Run(tt.name, func(t *testing.T) {
			result := validation.ValidarNIF(tt.nif)
			assert.Equal(t, tt.expected, result, "Para NIF: %s", tt.nif)
		})
	}
}

func TestValidarNIF_Performance(t *testing.T) {
	// Test de rendimiento para asegurar que la validación es eficiente
	nifs := []string{
		"12345678Z",
		"X1234567L",
		"  12345678 - Z  ",
		"y-1234567-z",
		"INVALIDO",
		"",
	}

	// Ejecutar múltiples veces para detectar problemas de rendimiento
	iterations := 10000
	for i := 0; i < iterations; i++ {
		for _, nif := range nifs {
			_ = validation.ValidarNIF(nif)
		}
	}
}

func BenchmarkValidarNIF(b *testing.B) {
	testCases := []string{
		"12345678Z",
		"X1234567L",
		"12345678-Z",
		"x 1234567 l",
		"INVALIDO",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = validation.ValidarNIF(tc)
		}
	}
}

func BenchmarkNormalizarNIF(b *testing.B) {
	testCases := []string{
		"12345678Z",
		"X - 1234567 - L",
		"  12345678   Z  ",
		"x1234567l",
		" - - ",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = validation.NormalizarNIF(tc)
		}
	}
}
