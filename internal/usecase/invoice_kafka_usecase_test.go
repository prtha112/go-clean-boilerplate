package usecase

import (
	"errors"
	"testing"

	"go-clean-boilerplate/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInvoiceKafkaRepository struct {
	mock.Mock
}

func (m *MockInvoiceKafkaRepository) Save(inv *domain.InvoiceKafka) error {
	args := m.Called(inv)
	return args.Error(0)
}

func TestInvoiceKafkaUsecase_HandleInvoice(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockInvoiceKafkaRepository)
		uc := NewInvoiceKafkaUsecase(mockRepo)
		invoice := &domain.InvoiceKafka{}
		mockRepo.On("Save", invoice).Return(nil)
		err := uc.HandleInvoice(invoice)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo error", func(t *testing.T) {
		mockRepo := new(MockInvoiceKafkaRepository)
		uc := NewInvoiceKafkaUsecase(mockRepo)
		invoice := &domain.InvoiceKafka{}
		repoErr := errors.New("repo error")
		mockRepo.On("Save", invoice).Return(repoErr)
		err := uc.HandleInvoice(invoice)
		assert.EqualError(t, err, "repo error")
		mockRepo.AssertExpectations(t)
	})
}
