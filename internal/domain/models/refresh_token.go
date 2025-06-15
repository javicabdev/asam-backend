package models

import (
	"time"
)

// RefreshToken represents a refresh token for JWT authentication
type RefreshToken struct {
	ID        uint      `gorm:"primaryKey"`
	UUID      string    `gorm:"uniqueIndex;not null;size:255"`
	UserID    uint      `gorm:"not null;index"`
	User      *User     `gorm:"foreignKey:UserID"`
	ExpiresAt int64     `gorm:"not null;index"`
	CreatedAt time.Time `gorm:"not null"`

	// Additional fields for better session management
	DeviceName string    `gorm:"size:255"`
	IPAddress  string    `gorm:"size:45"` // Supports IPv6
	UserAgent  string    `gorm:"type:text"`
	LastUsedAt time.Time `gorm:"index"`
}

// TableName specifies the table name for GORM
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().Unix() > rt.ExpiresAt
}
