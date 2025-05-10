package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
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

// JWTUtil proporciona funcionalidad para generar y validar tokens JWT
type JWTUtil struct {
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewJWTUtil crea una nueva instancia de JWTUtil
func NewJWTUtil(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTUtil {
	return &JWTUtil{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// GenerateTokenPair genera un par de tokens (acceso y refresco) para un usuario
func (j *JWTUtil) GenerateTokenPair(userID uint, role string) (*TokenDetails, error) {
	td := &TokenDetails{}
	now := time.Now()

	// Access Token
	td.AtExpires = now.Add(j.accessTTL).Unix()
	td.AccessUUID = uuid.New().String()

	atClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"uuid":    td.AccessUUID,
		"exp":     td.AtExpires,
		"iat":     now.Unix(),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString([]byte(j.accessSecret))
	if err != nil {
		return nil, err
	}
	td.AccessToken = accessToken

	// Refresh Token
	td.RtExpires = now.Add(j.refreshTTL).Unix()
	td.RefreshUUID = uuid.New().String()

	rtClaims := jwt.MapClaims{
		"user_id": userID,
		"uuid":    td.RefreshUUID,
		"exp":     td.RtExpires,
		"iat":     now.Unix(),
	}

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString([]byte(j.refreshSecret))
	if err != nil {
		return nil, err
	}
	td.RefreshToken = refreshToken

	return td, nil
}

// ValidateToken valida un token JWT
func (j *JWTUtil) ValidateToken(tokenString string, isRefreshToken bool) (*jwt.Token, error) {
	secret := j.accessSecret
	if isRefreshToken {
		secret = j.refreshSecret
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// ExtractClaims extracts and validates the claims from a JWT token.
// Returns the claims as a map and an error if the token is invalid or expired.
func (j *JWTUtil) ExtractClaims(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("claims inválidos o token expirado")
	}
	return claims, nil
}
