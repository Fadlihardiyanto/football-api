package model

import "time"

type Auth struct {
	ID   string
	Role string
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginResponse struct {
	User           *UserResponse `json:"user"`
	AccessToken    string        `json:"access_token"`
	RefreshTokenID string        `json:"refresh_token_id"`
	AccessExpiry   int64         `json:"access_expiry"`
}

type LoginResponseToFrontend struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	AccessExpiry int64         `json:"access_expiry"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin"`
}

type RegisterResponse struct {
	User          *UserResponse `json:"user"`
	AccessToken   string        `json:"access_token"`
	RefreshToken  string        `json:"refresh_token"`
	AccessExpiry  time.Time     `json:"access_expiry"`
	RefreshExpiry time.Time     `json:"refresh_expiry"`
}
