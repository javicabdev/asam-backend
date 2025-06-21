package utils_test

import (
	"testing"

	"github.com/javicabdev/asam-backend/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid emails
		{"valid email", "user@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"valid email with plus", "user+tag@example.com", true},
		{"valid email with dots", "user.name@example.com", true},
		{"valid email with underscore", "user_name@example.com", true},
		{"valid email with hyphen", "user-name@example.com", true},
		{"valid email with numbers", "user123@example.com", true},
		{"valid email with TLD", "user@example.co.uk", true},

		// Invalid emails
		{"no @ symbol", "userexample.com", false},
		{"no domain", "user@", false},
		{"no local part", "@example.com", false},
		{"no TLD", "user@example", false},
		{"multiple @", "user@@example.com", false},
		{"spaces", "user @example.com", false},
		{"empty string", "", false},
		{"just @", "@", false},

		// Edge cases
		{"with whitespace", " user@example.com ", true}, // Should be trimmed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.IsEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "User@Example.COM", "user@example.com"},
		{"trim spaces", "  user@example.com  ", "user@example.com"},
		{"mixed case with spaces", "  User.Name@Example.COM  ", "user.name@example.com"},
		{"already normalized", "user@example.com", "user@example.com"},
		{"not an email", "username", "username"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.NormalizeEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractUsernameFromEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple email", "user@example.com", "user"},
		{"email with dots", "user.name@example.com", "user.name"},
		{"email with plus", "user+tag@example.com", "user+tag"},
		{"not an email", "username", "username"},
		{"empty string", "", ""},
		{"email with subdomain", "user@mail.example.com", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExtractUsernameFromEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestObfuscateEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"standard email", "john.doe@example.com", "j***e@example.com"},
		{"short local part", "ab@example.com", "ab***@example.com"},
		{"single char local part", "a@example.com", "a***@example.com"},
		{"not an email", "username", "username"},
		{"email with subdomain", "user@mail.example.com", "u***r@mail.example.com"},
		{"long local part", "verylongusername@example.com", "v***e@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ObfuscateEmail(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
