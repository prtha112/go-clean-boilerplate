package usecase

import (
	"errors"
	"testing"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock InvoiceRepository
type mockInvoiceRepository struct {
	mock.Mock
}

func (m *mockInvoiceRepository) Create(invoice *domain.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepository) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetByInvoiceNumber(invoiceNumber string) (*domain.Invoice, error) {
	args := m.Called(invoiceNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetAll(limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) GetByStatus(status domain.InvoiceStatus, limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceRepository) Update(invoice *domain.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *mockInvoiceRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockInvoiceRepository) CreateInvoiceItem(item *domain.InvoiceItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *mockInvoiceRepository) GetInvoiceItems(invoiceID uuid.UUID) ([]*domain.InvoiceItem, error) {
	args := m.Called(invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InvoiceItem), args.Error(1)
}

func (m *mockInvoiceRepository) GenerateInvoiceNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Mock KafkaProducer
type mockKafkaProducer struct {
	mock.Mock
}

func (m *mockKafkaProducer) SendMessage(topic string, key string, message []byte) error {
	args := m.Called(topic, key, message)
	return args.Error(0)
}

func (m *mockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewInvoiceUsecase(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	assert.NotNil(t, usecase)
	assert.Implements(t, (*domain.InvoiceUsecase)(nil), usecase)
}

func TestInvoiceUsecase_Create_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 50.00,
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.10,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				ProductID:   &productID,
				Description: "Test Product",
				Quantity:    2,
				UnitPrice:   50.00,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-2024-001", nil)
	mockProductRepo.On("GetByID", productID).Return(product, nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafka.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	invoice, err := usecase.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, "INV-2024-001", invoice.InvoiceNumber)
	assert.Equal(t, req.CustomerName, invoice.CustomerName)
	assert.Equal(t, req.CustomerEmail, invoice.CustomerEmail)
	assert.Equal(t, domain.InvoiceStatusDraft, invoice.Status)
	assert.Equal(t, 100.00, invoice.SubTotal)    // 2 * 50.00
	assert.Equal(t, 10.00, invoice.TaxAmount)    // 100.00 * 0.10
	assert.Equal(t, 110.00, invoice.TotalAmount) // 100.00 + 10.00
	assert.Len(t, invoice.Items, 1)

	mockInvoiceRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_GenerateInvoiceNumberFails(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.10,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Test Product",
				Quantity:    1,
				UnitPrice:   50.00,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("", errors.New("failed to generate"))

	invoice, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "failed to generate invoice number")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_ProductNotFound(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	productID := uuid.New()
	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.10,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				ProductID:   &productID,
				Description: "Test Product",
				Quantity:    1,
				UnitPrice:   50.00,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-2024-001", nil)
	mockProductRepo.On("GetByID", productID).Return(nil, errors.New("product not found"))

	invoice, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "not found")
	mockInvoiceRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	orderID := uuid.New()
	productID := uuid.New()

	product := &domain.Product{
		ID:   productID,
		Name: "Test Product",
	}

	orderItem := &domain.OrderItem{
		ProductID: productID,
		Product:   product,
		Quantity:  2,
		Price:     50.00,
	}

	order := &domain.Order{
		ID:            orderID,
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items:         []*domain.OrderItem{orderItem},
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.08,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-2024-002", nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafka.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	invoice, err := usecase.CreateFromOrder(orderID, req)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, &orderID, invoice.OrderID)
	assert.Equal(t, order, invoice.Order)
	assert.Equal(t, 100.00, invoice.SubTotal)    // 2 * 50.00
	assert.Equal(t, 8.00, invoice.TaxAmount)     // 100.00 * 0.08
	assert.Equal(t, 108.00, invoice.TotalAmount) // 100.00 + 8.00
	assert.Len(t, invoice.Items, 1)

	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_OrderNotFound(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	orderID := uuid.New()
	req := &domain.CreateInvoiceRequest{}

	mockOrderRepo.On("GetByID", orderID).Return(nil, errors.New("order not found"))

	invoice, err := usecase.CreateFromOrder(orderID, req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "order not found")
	mockOrderRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByID_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceID := uuid.New()
	expectedInvoice := &domain.Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-2024-001",
		Status:        domain.InvoiceStatusDraft,
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(expectedInvoice, nil)

	invoice, err := usecase.GetByID(invoiceID)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoice, invoice)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByID_InvalidID(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoice, err := usecase.GetByID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "invalid invoice ID")
	mockInvoiceRepo.AssertNotCalled(t, "GetByID")
}

func TestInvoiceUsecase_UpdateStatus_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceID := uuid.New()
	invoice := &domain.Invoice{
		ID:     invoiceID,
		Status: domain.InvoiceStatusDraft,
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(invoice, nil)
	mockInvoiceRepo.On("Update", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafka.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := usecase.UpdateStatus(invoiceID, domain.InvoiceStatusSent)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_ToPaid_SetsPaidDate(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceID := uuid.New()
	invoice := &domain.Invoice{
		ID:       invoiceID,
		Status:   domain.InvoiceStatusSent,
		PaidDate: nil,
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(invoice, nil)
	mockInvoiceRepo.On("Update", mock.MatchedBy(func(inv *domain.Invoice) bool {
		return inv.Status == domain.InvoiceStatusPaid && inv.PaidDate != nil
	})).Return(nil)
	mockKafka.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := usecase.UpdateStatus(invoiceID, domain.InvoiceStatusPaid)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceID := uuid.New()

	err := usecase.UpdateStatus(invoiceID, "invalid_status")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice status")
	mockInvoiceRepo.AssertNotCalled(t, "GetByID")
}

func TestInvoiceUsecase_Delete_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceID := uuid.New()
	invoice := &domain.Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-2024-001",
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(invoice, nil)
	mockInvoiceRepo.On("Delete", invoiceID).Return(nil)
	mockKafka.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := usecase.Delete(invoiceID)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_SendToKafka_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-2024-001",
		Status:        domain.InvoiceStatusDraft,
	}

	mockKafka.On("SendMessage", "invoices", invoice.ID.String(), mock.AnythingOfType("[]uint8")).Return(nil)

	err := usecase.SendToKafka(invoice, "invoice_created")

	assert.NoError(t, err)
	mockKafka.AssertExpectations(t)
}

func TestInvoiceUsecase_SendToKafka_NoProducer(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, nil)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-2024-001",
	}

	err := usecase.SendToKafka(invoice, "invoice_created")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka producer not configured")
}

func TestInvoiceUsecase_GetByStatus_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), Status: domain.InvoiceStatusPaid},
		{ID: uuid.New(), Status: domain.InvoiceStatusPaid},
	}

	mockInvoiceRepo.On("GetByStatus", domain.InvoiceStatusPaid, 10, 0).Return(expectedInvoices, nil)

	invoices, err := usecase.GetByStatus(domain.InvoiceStatusPaid, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoices, invoices)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByInvoiceNumber_Success(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoiceNumber := "INV-2024-001"
	expectedInvoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: invoiceNumber,
	}

	mockInvoiceRepo.On("GetByInvoiceNumber", invoiceNumber).Return(expectedInvoice, nil)

	invoice, err := usecase.GetByInvoiceNumber(invoiceNumber)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoice, invoice)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByInvoiceNumber_EmptyNumber(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	invoice, err := usecase.GetByInvoiceNumber("")

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "invoice number is required")
	mockInvoiceRepo.AssertNotCalled(t, "GetByInvoiceNumber")
}

func TestInvoiceUsecase_GetAll_WithPagination(t *testing.T) {
	mockInvoiceRepo := new(mockInvoiceRepository)
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	mockKafka := new(mockKafkaProducer)

	usecase := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafka)

	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	tests := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "normal pagination",
			inputLimit:     20,
			inputOffset:    10,
			expectedLimit:  20,
			expectedOffset: 10,
		},
		{
			name:           "default limit",
			inputLimit:     0,
			inputOffset:    0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "max limit cap",
			inputLimit:     200,
			inputOffset:    0,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "negative offset",
			inputLimit:     10,
			inputOffset:    -5,
			expectedLimit:  10,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvoiceRepo.On("GetAll", tt.expectedLimit, tt.expectedOffset).Return(expectedInvoices, nil).Once()

			invoices, err := usecase.GetAll(tt.inputLimit, tt.inputOffset)

			assert.NoError(t, err)
			assert.Equal(t, expectedInvoices, invoices)
		})
	}

	mockInvoiceRepo.AssertExpectations(t)
}
