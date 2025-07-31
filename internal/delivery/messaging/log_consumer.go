package messaging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type LogFileConsumer struct {
	LogDir string
	Log    *logrus.Logger
}

func NewLogFileConsumer(logDir string, logger *logrus.Logger) *LogFileConsumer {
	return &LogFileConsumer{
		LogDir: logDir,
		Log:    logger,
	}
}

func (c *LogFileConsumer) Consume(message *kafka.Message) error {
	var event model.LogEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		c.Log.WithError(err).Error("Failed to unmarshal LogEvent")
		return err
	}

	// Create directory per service
	dir := filepath.Join(c.LogDir, event.Service)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Append to daily log file
	filename := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02"))
	filepath := filepath.Join(dir, filename)
	logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer logFile.Close()

	logLine := fmt.Sprintf("[%s] [%s] %s\n", event.Time, event.Level, event.Message)
	_, err = logFile.WriteString(logLine)
	if err != nil {
		return err
	}

	c.Log.Infof("LogEvent saved to file: %s", filepath)
	return nil
}
