package input

import (
	"context"
	"github.com/javicabdev/asam-backend/internal/domain/models"
)

// TokenDetails contiene la información de los tokens generados
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (*TokenDetails, error)
	Logout(ctx context.Context, accessToken string) error
	RefreshToken(ctx context.Context, refreshToken string) (*TokenDetails, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
}
