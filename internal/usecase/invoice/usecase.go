package invoice

import (
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

var _ domain.UseCase = (*invoiceUseCase)(nil)
