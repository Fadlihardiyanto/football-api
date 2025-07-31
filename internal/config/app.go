package config

import (
	"github.com/Fadlihardiyanto/football-api/internal/delivery/http"
	"github.com/Fadlihardiyanto/football-api/internal/delivery/http/middleware"
	"github.com/Fadlihardiyanto/football-api/internal/delivery/http/route"
	messaging "github.com/Fadlihardiyanto/football-api/internal/gateway"
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/Fadlihardiyanto/football-api/internal/repository"
	"github.com/Fadlihardiyanto/football-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/sirupsen/logrus"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB          *gorm.DB
	App         *gin.Engine
	Log         *logrus.Logger
	Validate    *validator.Validate
	Viper       *viper.Viper
	Producer    *kafka.Producer
	RedisClient *redis.Client
	JWTConfig   *model.JWTConfig
	RateLimiter gin.HandlerFunc
}

func Bootstrap(config *BootstrapConfig) {
	// Initialize repositories
	userRepo := repository.NewUserRepo(config.DB, config.Log)
	teamRepo := repository.NewTeamsRepo(config.DB, config.Log)
	playersRepo := repository.NewPlayersRepo(config.DB, config.Log)
	matchesRepo := repository.NewMatchesRepo(config.DB, config.Log)
	goalsRepo := repository.NewGoalsRepo(config.DB, config.Log)

	// Initialize producer
	logProducer := messaging.NewLogProducer(config.Producer, config.Log)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo, logProducer, config.DB, config.Log, config.JWTConfig, config.RedisClient)
	teamsUseCase := usecase.NewTeamsUseCase(teamRepo, logProducer, config.DB, config.Log)
	playersUseCase := usecase.NewPlayersUseCase(playersRepo, teamRepo, logProducer, config.DB, config.Log)
	matchesUseCase := usecase.NewMatchesUseCase(matchesRepo, logProducer, config.DB, config.Log)
	goalsUseCase := usecase.NewGoalsUseCase(goalsRepo, matchesRepo, playersRepo, logProducer, config.DB, config.Log)

	// Initialize controllers
	authController := http.NewAuthController(authUseCase, config.Log)
	teamsController := http.NewTeamsController(teamsUseCase, config.Log)
	playersController := http.NewPlayersController(playersUseCase, config.Log)
	matchesController := http.NewMatchesController(matchesUseCase, config.Log)
	goalsController := http.NewGoalsController(goalsUseCase, config.Log)

	// Set up middlewares
	authMiddleware := middleware.AuthMiddleware(authUseCase)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(config.Viper)

	routeConfig := route.RouteConfig{
		App:                   config.App,
		AuthController:        authController,
		TeamsController:       teamsController,
		PlayerController:      playersController,
		MatchesController:     matchesController,
		GoalsController:       goalsController,
		AuthMiddleware:        authMiddleware,
		RateLimiterMiddleware: rateLimiterMiddleware,
	}
	routeConfig.Setup()

	config.Log.Info("Bootstrap completed successfully")
}
