package consumer

import (
	"context"
	"encoding/json"

	"go-clean-boilerplate/internal/domain"
)

// InvoiceHandler adapts Kafka invoice events to the domain use case.
type InvoiceHandler struct {
	usecase domain.InvoiceKafkaUsecase
}

// NewInvoiceHandler creates a new InvoiceHandler with the given usecase.
func NewInvoiceHandler(uc domain.InvoiceKafkaUsecase) *InvoiceHandler {
	return &InvoiceHandler{usecase: uc}
}

// Handle implements Handler: it parses the raw Kafka message and delegates to the use case.
func (h *InvoiceHandler) Handle(_ context.Context, _ []byte, value []byte) error {
	var inv domain.InvoiceKafka
	if err := json.Unmarshal(value, &inv); err != nil {
		return err
	}
	return h.usecase.HandleInvoice(&inv)
}
