package messaging

import (
	"encoding/json"

	"github.com/Fadlihardiyanto/football-api/internal/model"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
)

type Producer[T model.Event] struct {
	Producer *kafka.Producer
	Topic    string
	Log      *logrus.Logger
}

func (p *Producer[T]) GetTopic() *string {
	return &p.Topic
}

func (p *Producer[T]) Send(event T) error {
	value, err := json.Marshal(event)
	if err != nil {
		p.Log.WithError(err).Error("failed to marshal event")
		return err
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     p.GetTopic(),
			Partition: int32(kafka.PartitionAny),
		},
		Value: value,
		Key:   []byte(event.GetKey()),
	}

	deliveryChan := make(chan kafka.Event)
	err = p.Producer.Produce(message, deliveryChan)
	if err != nil {
		p.Log.WithError(err).Error("failed to produce message")
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		p.Log.Errorf("Delivery failed: %v", m.TopicPartition.Error)
		return m.TopicPartition.Error
	}
	p.Log.Infof("Delivered message to topic %s [%d] at offset %v",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)

	return nil
}
