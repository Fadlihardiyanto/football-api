package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/config"
	"github.com/Fadlihardiyanto/football-api/internal/delivery/messaging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	viperConfig := config.NewViper()
	logger := config.NewLogger(viperConfig)

	logger.Info("Starting worker service")

	ctx, cancel := context.WithCancel(context.Background())

	go RunLogFileConsumer(logger, viperConfig, ctx)

	terminateSignals := make(chan os.Signal, 1)
	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM)

	stop := false
	for !stop {
		s := <-terminateSignals
		logger.Info("Got one of stop signals, shutting down worker gracefully, SIGNAL NAME :", s)
		cancel()
		stop = true
	}

	logger.Info("Waiting for all consumers to finish processing")

	time.Sleep(5 * time.Second) // wait for all consumers to finish processing
}

func RunLogFileConsumer(logger *logrus.Logger, viperConfig *viper.Viper, ctx context.Context) {
	logDir := viperConfig.GetString("LOG_FILE_DIR")
	if logDir == "" {
		logDir = "/app/logs"
	}

	logConsumer := config.NewKafkaConsumer(viperConfig, logger)
	handler := messaging.NewLogFileConsumer(logDir, logger)
	messaging.ConsumeTopic(ctx, logConsumer, "log-event", logger, handler.Consume)
}
