package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	v := viper.New()

	// Ambil dari ENV dulu agar yang dari Docker menang
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Baru ambil dari .env jika ada
	v.SetConfigFile(".env")
	err := v.ReadInConfig()
	if err != nil {
		fmt.Printf("⚠️ No .env file loaded: %v\n", err)
	}

	return v
}
