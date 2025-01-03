package db

import (
	"context"
	"errors"
	"github.com/javicabdev/asam-backend/internal/ports/output"
	"gorm.io/gorm"
	"log"
	"time"
)

// RefreshToken modelo para almacenar tokens de refresh
type RefreshToken struct {
	UUID      string `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;index"`
	ExpiresAt int64  `gorm:"not null"`
	CreatedAt time.Time
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) output.TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) SaveRefreshToken(ctx context.Context, uuid string, userId uint, expires int64) error {
	token := RefreshToken{
		UUID:      uuid,
		UserID:    userId,
		ExpiresAt: expires,
	}
	result := r.db.WithContext(ctx).Create(&token)
	return result.Error
}

func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).Exec("SELECT cleanup_expired_tokens()")
	return result.Error
}

func (r *tokenRepository) ValidateRefreshToken(ctx context.Context, uuid string, userId uint) error {
	// Primero limpiamos tokens expirados
	if err := r.CleanupExpiredTokens(ctx); err != nil {
		// Log el error pero continuamos - no es crítico
		log.Printf("Error limpiando tokens expirados: %v", err)
	}

	// Luego validamos el token específico
	var token RefreshToken
	result := r.db.WithContext(ctx).
		Where("uuid = ? AND user_id = ?", uuid, userId).
		First(&token)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("token no encontrado")
		}
		return result.Error
	}

	// Verificar expiración
	if time.Now().Unix() > token.ExpiresAt {
		return errors.New("token expirado")
	}

	return nil
}

func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, uuid string) error {
	result := r.db.WithContext(ctx).
		Where("uuid = ?", uuid).
		Delete(&RefreshToken{})

	return result.Error
}
