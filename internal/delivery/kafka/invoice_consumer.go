package kafka

import (
	"context"
	"encoding/json"
	"log"

	"go-clean-v2/internal/domain"
	"go-clean-v2/internal/usecase"

	"github.com/segmentio/kafka-go"
)

type InvoiceConsumer struct {
	reader  *kafka.Reader
	usecase *usecase.InvoiceKafkaUsecase
}

func NewInvoiceConsumer(brokers []string, topic, groupID string, uc *usecase.InvoiceKafkaUsecase) *InvoiceConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})

	return &InvoiceConsumer{reader: r, usecase: uc}
}

func (c *InvoiceConsumer) Start(ctx context.Context) {
	defer c.reader.Close()
	log.Println("Kafka consumer started")

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			log.Println("Kafka read error:", err)
			continue
		}

		var inv domain.InvoiceKafka
		if err := json.Unmarshal(msg.Value, &inv); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		if err := c.usecase.HandleInvoice(&inv); err != nil {
			log.Println("Usecase error:", err)
		} else {
			log.Printf("Invoice %s processed", inv.ID)
		}
	}
}
