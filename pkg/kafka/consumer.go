package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaConsumer struct {
	// Add fields for the consumer, such as reader, topic, etc.
	reader *kafka.Reader
}

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewKafkaConsumer(config *Config) *kafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		Topic:   config.Topic,
		GroupID: config.GroupID,
	})

	return &kafkaConsumer{
		reader: reader,
	}
}

func (c *kafkaConsumer) ReadMessage() (kafka.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return kafka.Message{}, err
	}

	return message, nil
}

func (c *kafkaConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
