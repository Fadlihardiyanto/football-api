package main

import (
	"fmt"

	"github.com/Fadlihardiyanto/football-api/internal/config"
)

func main() {
	fmt.Println("📌 main() started")

	viperConfig := config.NewViper()
	fmt.Println("viperConfig loaded")

	log := config.NewLogger(viperConfig)
	fmt.Println("logger created")

	db := config.NewDatabase(viperConfig, log)
	fmt.Println("database connected")

	validate := config.NewValidator(viperConfig)
	fmt.Println("validator created")

	app := config.NewGin(viperConfig)
	fmt.Println("gin app created")

	producer := config.NewKafkaProducer(viperConfig, log)
	fmt.Println("kafka producer created")

	jwtConfig := config.NewJWTConfig(viperConfig)
	fmt.Println("JWT config created")

	redisClient := config.NewRedisClient(viperConfig, log)
	fmt.Println("Redis client created")

	config.Bootstrap(&config.BootstrapConfig{
		DB:          db,
		App:         app,
		Log:         log,
		Validate:    validate,
		Viper:       viperConfig,
		Producer:    producer,
		JWTConfig:   jwtConfig,
		RedisClient: redisClient,
	})

	appPort := viperConfig.GetInt("APP_PORT")
	err := app.Run(fmt.Sprintf(":%d", appPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	fmt.Printf("Server is running on port %d\n", appPort)
}
