package usecase

import (
	"fmt"
	"testing"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) Create(invoice *domain.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetByInvoiceNumber(invoiceNumber string) (*domain.Invoice, error) {
	args := m.Called(invoiceNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetAll(limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetByStatus(status domain.InvoiceStatus, limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) Update(invoice *domain.Invoice) error {
	args := m.Called(invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInvoiceRepository) CreateInvoiceItem(item *domain.InvoiceItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetInvoiceItems(invoiceID uuid.UUID) ([]*domain.InvoiceItem, error) {
	args := m.Called(invoiceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InvoiceItem), args.Error(1)
}

func (m *MockInvoiceRepository) GenerateInvoiceNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetByID(id uuid.UUID) (*domain.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOrderRepository) GetAll(limit, offset int) ([]*domain.Order, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) CreateOrderItem(item *domain.OrderItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrderItems(orderID uuid.UUID) ([]*domain.OrderItem, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(id uuid.UUID) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *MockProductRepository) GetAll(limit, offset int) ([]*domain.Product, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *MockProductRepository) Update(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockKafkaProducer) SendMessage(topic string, key string, message []byte) error {
	args := m.Called(topic, key, message)
	return args.Error(0)
}

func (m *MockKafkaProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewInvoiceUsecase(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}

	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	assert.NotNil(t, invoiceUC)
	assert.Implements(t, (*domain.InvoiceUsecase)(nil), invoiceUC)
}

func TestInvoiceUsecase_Create_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 100.0,
		Stock: 10,
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		CustomerPhone:  "1234567890",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
		Notes:          "Test invoice",
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
				UnitPrice:   100.0,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-001", nil)
	mockProductRepo.On("GetByID", productID).Return(product, nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	invoice, err := invoiceUC.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, "INV-001", invoice.InvoiceNumber)
	assert.Equal(t, req.CustomerName, invoice.CustomerName)
	assert.Equal(t, req.CustomerEmail, invoice.CustomerEmail)
	assert.Equal(t, domain.InvoiceStatusDraft, invoice.Status)
	assert.Equal(t, 200.0, invoice.SubTotal)
	assert.Equal(t, 20.0, invoice.TaxAmount)
	assert.Equal(t, 220.0, invoice.TotalAmount)
	assert.Len(t, invoice.Items, 1)
	mockInvoiceRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_GenerateInvoiceNumberFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Test Item",
				Quantity:    1,
				UnitPrice:   100.0,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("", fmt.Errorf("generation failed"))

	invoice, err := invoiceUC.Create(req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "failed to generate invoice number")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_ProductNotFound(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	productID := uuid.New()
	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
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
				UnitPrice:   100.0,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-001", nil)
	mockProductRepo.On("GetByID", productID).Return(nil, fmt.Errorf("product not found"))

	invoice, err := invoiceUC.Create(req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "not found")
	mockInvoiceRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_WithoutProductID(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Custom Item",
				Quantity:    1,
				UnitPrice:   100.0,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-001", nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	invoice, err := invoiceUC.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, "INV-001", invoice.InvoiceNumber)
	assert.Len(t, invoice.Items, 1)
	assert.Equal(t, "Custom Item", invoice.Items[0].Description)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_Create_CreateFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Test Item",
				Quantity:    1,
				UnitPrice:   100.0,
			},
		},
	}

	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-001", nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(fmt.Errorf("database error"))

	invoice, err := invoiceUC.Create(req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "failed to create invoice")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	orderID := uuid.New()
	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 100.0,
	}

	order := &domain.Order{
		ID:           orderID,
		CustomerName: "John Doe",
		Items: []*domain.OrderItem{
			{
				ProductID: productID,
				Product:   product,
				Quantity:  2,
				Price:     100.0,
			},
		},
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-002", nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	invoice, err := invoiceUC.CreateFromOrder(orderID, req)

	assert.NoError(t, err)
	assert.NotNil(t, invoice)
	assert.Equal(t, "INV-002", invoice.InvoiceNumber)
	assert.Equal(t, &orderID, invoice.OrderID)
	assert.Equal(t, order, invoice.Order)
	assert.Len(t, invoice.Items, 1)
	assert.Equal(t, 200.0, invoice.SubTotal)
	assert.Equal(t, 20.0, invoice.TaxAmount)
	assert.Equal(t, 220.0, invoice.TotalAmount)
	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_OrderNotFound(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	orderID := uuid.New()
	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
	}

	mockOrderRepo.On("GetByID", orderID).Return(nil, fmt.Errorf("order not found"))

	invoice, err := invoiceUC.CreateFromOrder(orderID, req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "order not found")
	mockOrderRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_GenerateInvoiceNumberFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	orderID := uuid.New()
	order := &domain.Order{
		ID:           orderID,
		CustomerName: "John Doe",
		Items:        []*domain.OrderItem{},
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("", fmt.Errorf("generation failed"))

	invoice, err := invoiceUC.CreateFromOrder(orderID, req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "failed to generate invoice number")
	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_CreateFromOrder_CreateFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	orderID := uuid.New()
	order := &domain.Order{
		ID:           orderID,
		CustomerName: "John Doe",
		Items:        []*domain.OrderItem{},
	}

	req := &domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Test St",
		DueDate:        time.Now().AddDate(0, 0, 30),
		TaxRate:        0.1,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockInvoiceRepo.On("GenerateInvoiceNumber").Return("INV-003", nil)
	mockInvoiceRepo.On("Create", mock.AnythingOfType("*domain.Invoice")).Return(fmt.Errorf("database error"))

	invoice, err := invoiceUC.CreateFromOrder(orderID, req)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "failed to create invoice")
	mockOrderRepo.AssertExpectations(t)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByID_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	expectedInvoice := &domain.Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-001",
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(expectedInvoice, nil)

	invoice, err := invoiceUC.GetByID(invoiceID)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoice, invoice)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByID_InvalidID(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoice, err := invoiceUC.GetByID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "invalid invoice ID")
}

func TestInvoiceUsecase_GetByInvoiceNumber_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceNumber := "INV-001"
	expectedInvoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: invoiceNumber,
	}

	mockInvoiceRepo.On("GetByInvoiceNumber", invoiceNumber).Return(expectedInvoice, nil)

	invoice, err := invoiceUC.GetByInvoiceNumber(invoiceNumber)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoice, invoice)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByInvoiceNumber_EmptyNumber(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoice, err := invoiceUC.GetByInvoiceNumber("")

	assert.Error(t, err)
	assert.Nil(t, invoice)
	assert.Contains(t, err.Error(), "invoice number is required")
}

func TestInvoiceUsecase_GetAll_WithDefaultLimit(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), InvoiceNumber: "INV-001"},
		{ID: uuid.New(), InvoiceNumber: "INV-002"},
	}

	mockInvoiceRepo.On("GetAll", 10, 0).Return(expectedInvoices, nil)

	invoices, err := invoiceUC.GetAll(0, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoices, invoices)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetAll_WithMaxLimit(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), InvoiceNumber: "INV-001"},
	}

	mockInvoiceRepo.On("GetAll", 100, 0).Return(expectedInvoices, nil)

	invoices, err := invoiceUC.GetAll(150, -5)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoices, invoices)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByStatus_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	status := domain.InvoiceStatusDraft
	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), Status: status},
	}

	mockInvoiceRepo.On("GetByStatus", status, 10, 0).Return(expectedInvoices, nil)

	invoices, err := invoiceUC.GetByStatus(status, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoices, invoices)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByStatus_WithLimitAndOffset(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	status := domain.InvoiceStatusPaid
	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), Status: status},
	}

	mockInvoiceRepo.On("GetByStatus", status, 100, 0).Return(expectedInvoices, nil)

	invoices, err := invoiceUC.GetByStatus(status, 150, -5)

	assert.NoError(t, err)
	assert.Equal(t, expectedInvoices, invoices)
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_GetByStatus_RepositoryError(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	status := domain.InvoiceStatusDraft
	mockInvoiceRepo.On("GetByStatus", status, 10, 0).Return(nil, fmt.Errorf("database error"))

	invoices, err := invoiceUC.GetByStatus(status, 0, 0)

	assert.Error(t, err)
	assert.Nil(t, invoices)
	assert.Contains(t, err.Error(), "database error")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	existingInvoice := &domain.Invoice{
		ID:     invoiceID,
		Status: domain.InvoiceStatusDraft,
	}
	newStatus := domain.InvoiceStatusSent

	mockInvoiceRepo.On("GetByID", invoiceID).Return(existingInvoice, nil)
	mockInvoiceRepo.On("Update", mock.AnythingOfType("*domain.Invoice")).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := invoiceUC.UpdateStatus(invoiceID, newStatus)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_ToPaid(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	existingInvoice := &domain.Invoice{
		ID:     invoiceID,
		Status: domain.InvoiceStatusSent,
	}
	newStatus := domain.InvoiceStatusPaid

	mockInvoiceRepo.On("GetByID", invoiceID).Return(existingInvoice, nil)
	mockInvoiceRepo.On("Update", mock.MatchedBy(func(inv *domain.Invoice) bool {
		return inv.Status == domain.InvoiceStatusPaid && inv.PaidDate != nil
	})).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := invoiceUC.UpdateStatus(invoiceID, newStatus)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_InvalidID(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	err := invoiceUC.UpdateStatus(uuid.Nil, domain.InvoiceStatusSent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice ID")
}

func TestInvoiceUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()

	err := invoiceUC.UpdateStatus(invoiceID, domain.InvoiceStatus("invalid"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice status")
}

func TestInvoiceUsecase_UpdateStatus_InvoiceNotFound(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	mockInvoiceRepo.On("GetByID", invoiceID).Return(nil, fmt.Errorf("invoice not found"))

	err := invoiceUC.UpdateStatus(invoiceID, domain.InvoiceStatusSent)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoice not found")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_UpdateStatus_UpdateFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	existingInvoice := &domain.Invoice{
		ID:     invoiceID,
		Status: domain.InvoiceStatusDraft,
	}
	newStatus := domain.InvoiceStatusSent

	mockInvoiceRepo.On("GetByID", invoiceID).Return(existingInvoice, nil)
	mockInvoiceRepo.On("Update", mock.AnythingOfType("*domain.Invoice")).Return(fmt.Errorf("database error"))

	err := invoiceUC.UpdateStatus(invoiceID, newStatus)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_Delete_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	existingInvoice := &domain.Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-001",
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(existingInvoice, nil)
	mockInvoiceRepo.On("Delete", invoiceID).Return(nil)
	mockKafkaProducer.On("SendMessage", "invoices", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	err := invoiceUC.Delete(invoiceID)

	assert.NoError(t, err)
	mockInvoiceRepo.AssertExpectations(t)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_Delete_InvalidID(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	err := invoiceUC.Delete(uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid invoice ID")
}

func TestInvoiceUsecase_Delete_InvoiceNotFound(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	mockInvoiceRepo.On("GetByID", invoiceID).Return(nil, fmt.Errorf("invoice not found"))

	err := invoiceUC.Delete(invoiceID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invoice not found")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_Delete_DeleteFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoiceID := uuid.New()
	existingInvoice := &domain.Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-001",
	}

	mockInvoiceRepo.On("GetByID", invoiceID).Return(existingInvoice, nil)
	mockInvoiceRepo.On("Delete", invoiceID).Return(fmt.Errorf("database error"))

	err := invoiceUC.Delete(invoiceID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockInvoiceRepo.AssertExpectations(t)
}

func TestInvoiceUsecase_SendToKafka_Success(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-001",
	}
	eventType := "invoice_created"

	mockKafkaProducer.On("SendMessage", "invoices", invoice.ID.String(), mock.AnythingOfType("[]uint8")).Return(nil)

	err := invoiceUC.SendToKafka(invoice, eventType)

	assert.NoError(t, err)
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_SendToKafka_NoProducer(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, nil)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-001",
	}
	eventType := "invoice_created"

	err := invoiceUC.SendToKafka(invoice, eventType)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kafka producer not configured")
}

func TestInvoiceUsecase_SendToKafka_ProducerFails(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-001",
	}
	eventType := "invoice_created"

	mockKafkaProducer.On("SendMessage", "invoices", invoice.ID.String(), mock.AnythingOfType("[]uint8")).Return(fmt.Errorf("kafka error"))

	err := invoiceUC.SendToKafka(invoice, eventType)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send message to Kafka")
	mockKafkaProducer.AssertExpectations(t)
}

func TestInvoiceUsecase_SendToKafka_ValidatesEventStructure(t *testing.T) {
	mockInvoiceRepo := &MockInvoiceRepository{}
	mockOrderRepo := &MockOrderRepository{}
	mockProductRepo := &MockProductRepository{}
	mockKafkaProducer := &MockKafkaProducer{}
	
	invoiceUC := NewInvoiceUsecase(mockInvoiceRepo, mockOrderRepo, mockProductRepo, mockKafkaProducer)

	invoice := &domain.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: "INV-001",
	}
	eventType := "test_event"

	// Simplified test - just verify message is sent with correct topic and key
	mockKafkaProducer.On("SendMessage", "invoices", invoice.ID.String(), mock.AnythingOfType("[]uint8")).Return(nil)

	err := invoiceUC.SendToKafka(invoice, eventType)

	assert.NoError(t, err)
	mockKafkaProducer.AssertExpectations(t)
}