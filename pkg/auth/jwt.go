// Package auth provides authentication and authorization utilities for the ASAM backend.
// It includes JWT token generation, validation, and claim extraction functionality.
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
	tokenType := "access"
	if isRefreshToken {
		secret = j.refreshSecret
		tokenType = "refresh"
	}

	// Log token validation attempt (only in development)
	if tokenString != "" && len(tokenString) > 20 {
		fmt.Printf("[JWT-DEBUG] Validating %s token: %s...%s\n", tokenType, tokenString[:10], tokenString[len(tokenString)-4:])
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		// Enhanced error logging
		if ve, ok := err.(*jwt.ValidationError); ok {
			switch {
			case ve.Errors&jwt.ValidationErrorMalformed != 0:
				fmt.Printf("[JWT-DEBUG] Token is malformed\n")
			case ve.Errors&jwt.ValidationErrorExpired != 0:
				// Try to extract expiration time
				if token != nil {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						if exp, ok := claims["exp"].(float64); ok {
							expTime := time.Unix(int64(exp), 0)
							fmt.Printf("[JWT-DEBUG] Token expired at: %v (current time: %v)\n", expTime, time.Now())
						}
					}
				}
			case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
				fmt.Printf("[JWT-DEBUG] Token is not valid yet\n")
			default:
				fmt.Printf("[JWT-DEBUG] Token validation error: %v\n", ve.Inner)
			}
		}
		return nil, err
	}

	// Log successful validation
	if token != nil && token.Valid {
		fmt.Printf("[JWT-DEBUG] Token validated successfully\n")
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
