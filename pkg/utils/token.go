package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSecureToken genera un token aleatorio seguro
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateVerificationToken genera un token para verificación de email
func GenerateVerificationToken() (string, error) {
	// 32 bytes = 64 caracteres hexadecimales
	return GenerateSecureToken(32)
}

// GeneratePasswordResetToken genera un token para recuperación de contraseña
func GeneratePasswordResetToken() (string, error) {
	// 32 bytes = 64 caracteres hexadecimales
	return GenerateSecureToken(32)
}
