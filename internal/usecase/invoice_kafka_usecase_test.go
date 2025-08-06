package usecase

import (
	"errors"
	"testing"

	"go-clean-boilerplate/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock InvoiceKafkaRepository
type mockInvoiceKafkaRepository struct {
	mock.Mock
}

func (m *mockInvoiceKafkaRepository) Save(inv *domain.InvoiceKafka) error {
	args := m.Called(inv)
	return args.Error(0)
}

func TestNewInvoiceKafkaUsecase(t *testing.T) {
	mockRepo := new(mockInvoiceKafkaRepository)
	usecase := NewInvoiceKafkaUsecase(mockRepo)

	assert.NotNil(t, usecase)
}

func TestInvoiceKafkaUsecase_HandleInvoice_Success(t *testing.T) {
	mockRepo := new(mockInvoiceKafkaRepository)
	usecase := NewInvoiceKafkaUsecase(mockRepo)

	invoice := &domain.InvoiceKafka{
		ID:     1,
		Amount: 100.50,
	}

	mockRepo.On("Save", invoice).Return(nil)

	err := usecase.HandleInvoice(invoice)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestInvoiceKafkaUsecase_HandleInvoice_Error(t *testing.T) {
	mockRepo := new(mockInvoiceKafkaRepository)
	usecase := NewInvoiceKafkaUsecase(mockRepo)

	invoice := &domain.InvoiceKafka{
		ID:     2,
		Amount: 200.75,
	}

	mockRepo.On("Save", invoice).Return(errors.New("database error"))

	err := usecase.HandleInvoice(invoice)

	assert.Error(t, err)
	assert.EqualError(t, err, "database error")
	mockRepo.AssertExpectations(t)
}
