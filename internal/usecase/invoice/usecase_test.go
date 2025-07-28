package invoice

import (
	"encoding/json"
	domain "go-clean-architecture/internal/domain/invoice"
	"testing"
)

type mockRepo struct {
	called bool
}

func (m *mockRepo) CreateInvoice(invoice *domain.Invoice) error {
	m.called = true
	return nil
}

func TestConsumeInvoiceMessage(t *testing.T) {
	repo := &mockRepo{}
	uc := NewInvoiceUseCase(repo)
	inv := domain.Invoice{ID: "1", OrderID: "2", Amount: 100, CreatedAt: 123456}
	msg, _ := json.Marshal(inv)
	err := uc.ConsumeInvoiceMessage(msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.called {
		t.Fatal("repo.CreateInvoice not called")
	}
}
