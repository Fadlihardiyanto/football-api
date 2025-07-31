package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(v *viper.Viper, log *logrus.Logger) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		v.GetString("DB_HOST"),
		v.GetString("DB_PORT"),
		v.GetString("DB_USER"),
		v.GetString("DB_PASSWORD"),
		v.GetString("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(&logrusWriter{Logger: log}, logger.Config{
			SlowThreshold:             time.Second * 5,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  logger.Info,
		}),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	connection, err := db.DB()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	connection.SetMaxIdleConns(v.GetInt("DB_POOL_IDLE"))
	connection.SetMaxOpenConns(v.GetInt("DB_POOL_MAX"))
	connection.SetConnMaxLifetime(time.Second * time.Duration(v.GetInt("DB_POOL_LIFETIME")))

	log.Println("database connected")
	return db
}

type logrusWriter struct {
	Logger *logrus.Logger
}

func (l *logrusWriter) Printf(message string, args ...interface{}) {
	l.Logger.Tracef(message, args...)
}
