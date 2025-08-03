package invoice

import (
	"context"
	"encoding/json"
	domain "go-clean-architecture/internal/domain/invoice"
)

type invoiceUseCase struct {
	repo domain.Repository
}

func NewInvoiceUseCase(repo domain.Repository) *invoiceUseCase {
	return &invoiceUseCase{repo: repo}
}

func (u *invoiceUseCase) ConsumeInvoiceMessage(msg []byte) error {
	var inv domain.Invoice
	if err := json.Unmarshal(msg, &inv); err != nil {
		return err
	}
	return u.repo.CreateInvoice(&inv)
}

// ProduceInvoiceMessage implements invoice.UseCase.
func (u *invoiceUseCase) ProduceInvoiceMessage(ctx context.Context, invoice *domain.Invoice) error {
	// Assume span already started in handler, just use ctx
	data, err := json.Marshal(invoice)
	if err != nil {
		return err
	}
	return u.repo.PublishInvoiceMessage(ctx, data)
}

var _ domain.UseCase = (*invoiceUseCase)(nil)
