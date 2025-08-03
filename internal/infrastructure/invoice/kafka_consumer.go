package invoice

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaInvoiceConsumer struct {
	Reader  *kafka.Reader
	Handler func(msg []byte) error
}

func NewKafkaInvoiceConsumer(reader *kafka.Reader, handler func(msg []byte) error) *KafkaInvoiceConsumer {
	return &KafkaInvoiceConsumer{
		Reader:  reader,
		Handler: handler,
	}
}

func (c *KafkaInvoiceConsumer) Start(ctx context.Context) {
	log.Printf("Kafka consumer started for topic %s", c.Reader.Config().Topic)
	defer c.Reader.Close()
	for {
		m, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("kafka context error: %v", ctx.Err())
				return
			}
			log.Printf("kafka read error: %v", err)
			continue
		}
		log.Printf("Kafka message received: partition=%d offset=%d", m.Partition, m.Offset)
		if err := c.Handler(m.Value); err != nil {
			log.Printf("invoice handler error: %v", err)
		}
	}
}
