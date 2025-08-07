package consumer

import (
	"context"
	"encoding/json"
	"log"

	"go-clean-boilerplate/internal/domain"
	"go-clean-boilerplate/internal/usecase"
	"go-clean-boilerplate/pkg/kafka"
)

type InvoiceConsumer struct {
	consumer kafka.KafkaConsumer
	usecase  *usecase.InvoiceKafkaUsecase
}

func NewInvoiceConsumer(consumer kafka.KafkaConsumer, uc *usecase.InvoiceKafkaUsecase) *InvoiceConsumer {
	return &InvoiceConsumer{consumer: consumer, usecase: uc}
}

func (c *InvoiceConsumer) Start(ctx context.Context) {
	log.Println("Kafka consumer starting...")

	for {
		msg, err := c.consumer.ReadMessage()
		if err != nil {
			log.Println("Kafka read error:", err)
			continue
		}

		log.Printf("Kafka message payload: %s", string(msg.Value))

		var inv domain.InvoiceKafka
		if err := json.Unmarshal(msg.Value, &inv); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		if err := c.usecase.HandleInvoice(&inv); err != nil {
			log.Println("Usecase error:", err)
		} else {
			log.Printf("Invoice %d processed", inv.ID)
		}
	}
}

// Close closes the consumer
func (c *InvoiceConsumer) Close() error {
	log.Println("Closing Kafka consumer")
	return c.consumer.Close()
}
