package config

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewKafkaConsumer(v *viper.Viper, log *logrus.Logger) *kafka.Consumer {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": v.GetString("KAFKA_BOOTSTRAP_SERVERS"),
		"group.id":          v.GetString("KAFKA_GROUP_ID"),
		"auto.offset.reset": v.GetString("KAFKA_AUTO_OFFSET_RESET"),
	}

	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	return consumer
}

func NewKafkaProducer(v *viper.Viper, log *logrus.Logger) *kafka.Producer {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": v.GetString("KAFKA_BOOTSTRAP_SERVERS"),
	}

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	return producer
}
