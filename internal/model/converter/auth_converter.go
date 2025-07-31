package converter

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/entity"
	"github.com/Fadlihardiyanto/football-api/internal/model"
)

func ToLoginResponse(user *entity.User, accessToken, refreshTokenID string, accessExpiry time.Time) *model.LoginResponse {
	return &model.LoginResponse{
		User:           ToUserResponse(user),
		AccessToken:    accessToken,
		RefreshTokenID: refreshTokenID,
		AccessExpiry:   accessExpiry.Unix(),
	}
}

func ToUserResponse(user *entity.User) *model.UserResponse {
	if user == nil {
		return nil
	}

	return &model.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		IsActive: user.IsActive,
	}
}

func ToRegisterResponse(user *entity.User) *model.RegisterResponse {
	if user == nil {
		return nil
	}

	return &model.RegisterResponse{
		User:          ToUserResponse(user),
		AccessToken:   "",
		RefreshToken:  "",
		AccessExpiry:  time.Time{},
		RefreshExpiry: time.Time{},
	}
}

func ToRefreshTokenResponse(refreshTokenID, accessToken string, accessExpiry time.Time) *model.RefreshTokenResponse {
	return &model.RefreshTokenResponse{
		RefreshTokenID: refreshTokenID,
		AccessToken:    accessToken,
		AccessExpiry:   accessExpiry.Unix(),
	}
}
