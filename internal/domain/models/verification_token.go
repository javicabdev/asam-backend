package models

import (
	"time"
)

// TokenType represents the type of token
type TokenType string

const (
	// TokenTypeEmailVerification represents email verification token type
	TokenTypeEmailVerification TokenType = "email_verification" //nolint:gosec // This is a token type, not a credential
	// TokenTypePasswordReset represents password reset token type
	TokenTypePasswordReset TokenType = "password_reset"
)

// VerificationToken represents a verification or reset token
type VerificationToken struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`

	Token     string `gorm:"not null;uniqueIndex;size:64"`
	UserID    uint   `gorm:"not null;index"`
	Type      string `gorm:"not null;size:20"`
	Email     string `gorm:"not null;size:100"`
	Used      bool   `gorm:"not null;default:false"`
	UsedAt    *time.Time
	ExpiresAt time.Time `gorm:"not null"`

	// Associations
	User *User `gorm:"foreignKey:UserID"`
}

// IsExpired checks if the token has expired
func (t *VerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsed checks if the token has been used
func (t *VerificationToken) IsUsed() bool {
	return t.Used
}

// IsValid checks if the token is valid (not expired and not used)
func (t *VerificationToken) IsValid() bool {
	return !t.IsExpired() && !t.IsUsed()
}

// MarkAsUsed marks the token as used
func (t *VerificationToken) MarkAsUsed() {
	t.Used = true
	now := time.Now()
	t.UsedAt = &now
}

// Use is an alias for MarkAsUsed for compatibility
func (t *VerificationToken) Use() {
	t.MarkAsUsed()
}
