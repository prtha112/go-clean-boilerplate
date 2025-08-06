package usecase

import "go-clean-boilerplate/internal/domain"

type InvoiceKafkaUsecase struct {
	repo domain.InvoiceKafkaRepository
}

func NewInvoiceKafkaUsecase(repo domain.InvoiceKafkaRepository) *InvoiceKafkaUsecase {
	return &InvoiceKafkaUsecase{repo: repo}
}

func (uc *InvoiceKafkaUsecase) HandleInvoice(inv *domain.InvoiceKafka) error {
	// Logic to handle the invoice, e.g., validation, transformation, etc.
	return uc.repo.Save(inv)
}
