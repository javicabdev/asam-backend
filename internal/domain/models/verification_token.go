package models

import (
	"gorm.io/gorm"
	"time"
)

// TokenType define los tipos de tokens disponibles
type TokenType string

const (
	// TokenTypeEmailVerification para verificación de email
	TokenTypeEmailVerification TokenType = "email_verification" //nolint:gosec // This is a token type identifier, not a credential
	// TokenTypePasswordReset para recuperación de contraseña
	TokenTypePasswordReset TokenType = "password_reset"
)

// VerificationToken representa tokens para verificación de email y recuperación de contraseña
type VerificationToken struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null;size:64"`
	UserID    uint      `gorm:"not null;index"`
	User      User      `gorm:"foreignKey:UserID"`
	Type      TokenType `gorm:"not null;size:20"`
	Email     string    `gorm:"not null;size:100"` // Email al que se envió
	Used      bool      `gorm:"not null;default:false"`
	UsedAt    *time.Time
	ExpiresAt time.Time `gorm:"not null"`
}

// IsExpired verifica si el token ha expirado
func (t *VerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid verifica si el token es válido (no usado y no expirado)
func (t *VerificationToken) IsValid() bool {
	return !t.Used && !t.IsExpired()
}

// MarkAsUsed marca el token como usado
func (t *VerificationToken) MarkAsUsed() {
	t.Used = true
	now := time.Now()
	t.UsedAt = &now
}
