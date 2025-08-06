package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

type ConfigProducer struct {
	Brokers []string
	Topic   string
}

func NewKafkaProducer(config *ConfigProducer) domain.KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Compression:  kafka.Snappy,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
		ErrorLogger:  kafka.LoggerFunc(log.Printf),
	}

	return &kafkaProducer{
		writer: writer,
	}
}

func (p *kafkaProducer) SendMessage(topic string, key string, message []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: message,
		Time:  time.Now(),
	}

	err := p.writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Message sent to Kafka topic '%s' with key '%s'", topic, key)
	return nil
}

func (p *kafkaProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// Helper function to create a simple Kafka producer for testing
func NewSimpleProducer(brokers []string) domain.KafkaProducer {
	config := &ConfigProducer{
		Brokers: brokers,
		Topic:   "invoices", // default topic
	}
	return NewKafkaProducer(config)
}
