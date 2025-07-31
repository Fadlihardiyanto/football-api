package config

import (
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/model"

	"github.com/spf13/viper"
)

func NewJWTConfig(v *viper.Viper) *model.JWTConfig {
	return &model.JWTConfig{
		SecretKey:     v.GetString("JWT_SECRET"),
		AccessExpiry:  time.Duration(v.GetInt("JWT_ACCESS_TOKEN_EXPIRY")) * time.Minute,
		RefreshExpiry: time.Duration(v.GetInt("JWT_REFRESH_TOKEN_EXPIRY")) * time.Minute,
		Issuer:        v.GetString("JWT_ISSUER"),
	}
}
