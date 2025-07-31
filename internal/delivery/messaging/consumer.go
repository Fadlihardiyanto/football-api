package messaging

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type ConsumerHandler func(message *kafka.Message) error

func ConsumeTopic(ctx context.Context, consumer *kafka.Consumer, topic string, log *logrus.Logger, handler ConsumerHandler) {
	err := consumer.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe to topic: %v", err)
	}

	run := true

	for run {
		select {
		case <-ctx.Done():
			run = false
		default:
			ev := consumer.Poll(100) // 100ms timeout
			switch e := ev.(type) {
			case *kafka.Message:
				if err := handler(e); err != nil {
					log.Errorf("Failed to process message: %v", err)
				} else {
					_, err = consumer.CommitMessage(e)
					if err != nil {
						log.Errorf("Failed to commit message: %v", err)
					}
				}
			case kafka.Error:
				if e.IsFatal() {
					log.Fatalf("Fatal consumer error: %v", e)
				} else {
					log.Warnf("Consumer warning: %v", e)
				}
			}
		}
	}

	log.Infof("Closing consumer for topic : %s", topic)
	err = consumer.Close()
	if err != nil {
		panic(err)
	}
}
