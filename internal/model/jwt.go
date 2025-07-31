package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT struct
type JWTConfig struct {
	SecretKey     string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type UserClaims struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	TokenID string `json:"token_id"`
	jwt.RegisteredClaims
}

type RefreshTokenRequest struct {
	ID string `json:"id" validate:"required"`
}

type RefreshTokenResponse struct {
	RefreshTokenID string `json:"refresh_token_id,omitempty"`
	AccessToken    string `json:"access_token,omitempty"`
	AccessExpiry   int64  `json:"access_expiry,omitempty"`
}

type LogoutRequest struct {
	RefreshTokenID string `json:"refresh_token_id" validate:"required"`
}
