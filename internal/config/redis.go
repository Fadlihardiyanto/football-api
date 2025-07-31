package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRedisClient(v *viper.Viper, log *logrus.Logger) *redis.Client {
	addr := fmt.Sprintf("%s:%s", v.GetString("REDIS_HOST"), v.GetString("REDIS_PORT"))

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: v.GetString("REDIS_PASSWORD"),
		DB:       0,
	})

	// Check connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis at %s: %v", addr, err)
	}

	log.Infof("Connected to Redis at %s", addr)
	return client
}
