package messaging

import (
	"github.com/Fadlihardiyanto/football-api/internal/model"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type LogProducer struct {
	Producer[*model.LogEvent]
}

func NewLogProducer(producer *kafka.Producer, log *logrus.Logger) *LogProducer {
	return &LogProducer{
		Producer: Producer[*model.LogEvent]{
			Producer: producer,
			Topic:    "log-event",
			Log:      log,
		},
	}
}
