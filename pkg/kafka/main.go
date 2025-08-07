package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type kafkaProducer struct {
	writer *kafka.Writer
}

type kafkaConsumer struct {
	reader *kafka.Reader
	config *Config
}

type Config struct {
	Brokers                []string
	Topic                  string
	GroupID                string
	ReadTimeout            time.Duration
	MinBytes               int
	MaxBytes               int
	MaxWait                time.Duration
	CommitInterval         time.Duration
	QueueCapacity          int
	ReadLagInterval        time.Duration
	WatchPartitionChanges  bool
	PartitionWatchInterval time.Duration
}

type KafkaProducer interface {
	SendMessage(topic string, key string, message []byte) error
	Close() error
}

type KafkaConsumer interface {
	ReadMessage() (kafka.Message, error)
	Close() error
}

func NewKafkaProducer(config *Config) KafkaProducer {
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

func NewKafkaConsumer(config *Config) KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                config.Brokers,
		Topic:                  config.Topic,
		GroupID:                config.GroupID,
		MinBytes:               config.MinBytes,
		MaxBytes:               config.MaxBytes,
		MaxWait:                config.MaxWait,
		CommitInterval:         config.CommitInterval,
		QueueCapacity:          config.QueueCapacity,
		ReadLagInterval:        config.ReadLagInterval,
		WatchPartitionChanges:  config.WatchPartitionChanges,
		PartitionWatchInterval: config.PartitionWatchInterval,
	})

	return &kafkaConsumer{
		reader: reader,
		config: config,
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

func (c *kafkaConsumer) ReadMessage() (kafka.Message, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	if c.config.ReadTimeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), c.config.ReadTimeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
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
