package usecase

import (
	"errors"
	"testing"

	"go-clean-boilerplate/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OrderRepository
type mockOrderRepository struct {
	mock.Mock
}

func (m *mockOrderRepository) Create(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderRepository) GetByID(id uuid.UUID) (*domain.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderRepository) GetAll(limit, offset int) ([]*domain.Order, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *mockOrderRepository) Update(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockOrderRepository) CreateOrderItem(item *domain.OrderItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *mockOrderRepository) GetOrderItems(orderID uuid.UUID) ([]*domain.OrderItem, error) {
	args := m.Called(orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.OrderItem), args.Error(1)
}

func TestNewOrderUsecase(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	assert.NotNil(t, usecase)
	assert.Implements(t, (*domain.OrderUsecase)(nil), usecase)
}

func TestOrderUsecase_Create_Success(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 50.00,
		Stock: 10,
	}

	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	// Mock product repository calls
	mockProductRepo.On("GetByID", productID).Return(product, nil)
	updatedProduct := *product
	updatedProduct.Stock = 8 // Stock reduced by 2
	mockProductRepo.On("Update", &updatedProduct).Return(nil)

	// Mock order repository call
	mockOrderRepo.On("Create", mock.AnythingOfType("*domain.Order")).Return(nil)

	order, err := usecase.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, req.CustomerName, order.CustomerName)
	assert.Equal(t, req.CustomerEmail, order.CustomerEmail)
	assert.Equal(t, domain.OrderStatusPending, order.Status)
	assert.Equal(t, 100.00, order.TotalAmount) // 2 * 50.00
	assert.Len(t, order.Items, 1)
	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_Create_EmptyCustomerName(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	req := &domain.CreateOrderRequest{
		CustomerName:  "",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
	}

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "customer name is required")
	mockOrderRepo.AssertNotCalled(t, "Create")
	mockProductRepo.AssertNotCalled(t, "GetByID")
}

func TestOrderUsecase_Create_EmptyCustomerEmail(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
	}

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "customer email is required")
	mockOrderRepo.AssertNotCalled(t, "Create")
	mockProductRepo.AssertNotCalled(t, "GetByID")
}

func TestOrderUsecase_Create_NoItems(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{},
	}

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "order must have at least one item")
	mockOrderRepo.AssertNotCalled(t, "Create")
	mockProductRepo.AssertNotCalled(t, "GetByID")
}

func TestOrderUsecase_Create_ProductNotFound(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	productID := uuid.New()
	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  1,
			},
		},
	}

	mockProductRepo.On("GetByID", productID).Return(nil, errors.New("product not found"))

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "not found")
	mockOrderRepo.AssertNotCalled(t, "Create")
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_Create_InsufficientStock(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 50.00,
		Stock: 2, // Only 2 in stock
	}

	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  5, // Requesting 5, but only 2 available
			},
		},
	}

	mockProductRepo.On("GetByID", productID).Return(product, nil)

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "insufficient stock")
	mockOrderRepo.AssertNotCalled(t, "Create")
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_Create_OrderCreationFails_StockRollback(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Test Product",
		Price: 50.00,
		Stock: 10,
	}

	req := &domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	// Mock product repository calls
	mockProductRepo.On("GetByID", productID).Return(product, nil)
	updatedProduct := *product
	updatedProduct.Stock = 8 // Stock reduced by 2
	mockProductRepo.On("Update", &updatedProduct).Return(nil)

	// Mock order creation failure
	mockOrderRepo.On("Create", mock.AnythingOfType("*domain.Order")).Return(errors.New("database error"))

	// Mock rollback calls
	rollbackProduct := *product
	rollbackProduct.Stock = 10 // Stock restored
	mockProductRepo.On("GetByID", productID).Return(&updatedProduct, nil)
	mockProductRepo.On("Update", &rollbackProduct).Return(nil)

	order, err := usecase.Create(req)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "failed to create order")
	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_GetByID_Success(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()
	expectedOrder := &domain.Order{
		ID:            orderID,
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Status:        domain.OrderStatusPending,
		TotalAmount:   100.00,
	}

	mockOrderRepo.On("GetByID", orderID).Return(expectedOrder, nil)

	order, err := usecase.GetByID(orderID)

	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, expectedOrder, order)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUsecase_Update(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	validID := uuid.New()
	validOrder := &domain.Order{
		ID:            validID,
		CustomerName:  "John",
		CustomerEmail: "john@example.com",
	}

	t.Run("invalid order ID", func(t *testing.T) {
		order := &domain.Order{ID: uuid.Nil, CustomerName: "A", CustomerEmail: "B"}
		err := usecase.Update(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid order ID")
	})

	t.Run("empty customer name", func(t *testing.T) {
		order := &domain.Order{ID: validID, CustomerName: "", CustomerEmail: "B"}
		err := usecase.Update(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer name is required")
	})

	t.Run("empty customer email", func(t *testing.T) {
		order := &domain.Order{ID: validID, CustomerName: "A", CustomerEmail: ""}
		err := usecase.Update(order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "customer email is required")
	})

	t.Run("order not found", func(t *testing.T) {
		mockOrderRepo.On("GetByID", validID).Return(nil, errors.New("not found")).Once()
		err := usecase.Update(validOrder)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockOrderRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		mockOrderRepo.On("GetByID", validID).Return(validOrder, nil).Once()
		mockOrderRepo.On("Update", validOrder).Return(nil).Once()
		err := usecase.Update(validOrder)
		assert.NoError(t, err)
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUsecase_GetByID_InvalidID(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	order, err := usecase.GetByID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, order)
	assert.Contains(t, err.Error(), "invalid order ID")
	mockOrderRepo.AssertNotCalled(t, "GetByID")
}

func TestOrderUsecase_UpdateStatus_Success(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()
	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPending,
		Items:  []*domain.OrderItem{},
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Update", mock.AnythingOfType("*domain.Order")).Return(nil)

	err := usecase.UpdateStatus(orderID, domain.OrderStatusConfirmed)

	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUsecase_UpdateStatus_InvalidStatus(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()

	err := usecase.UpdateStatus(orderID, "invalid_status")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid order status")
	mockOrderRepo.AssertNotCalled(t, "GetByID")
}

func TestOrderUsecase_UpdateStatus_CancelOrder_RestoreStock(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()
	productID := uuid.New()

	orderItem := &domain.OrderItem{
		ProductID: productID,
		Quantity:  3,
	}

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPending,
		Items:  []*domain.OrderItem{orderItem},
	}

	product := &domain.Product{
		ID:    productID,
		Stock: 5,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockProductRepo.On("GetByID", productID).Return(product, nil)

	updatedProduct := *product
	updatedProduct.Stock = 8 // Stock restored by 3
	mockProductRepo.On("Update", &updatedProduct).Return(nil)

	mockOrderRepo.On("Update", mock.AnythingOfType("*domain.Order")).Return(nil)

	err := usecase.UpdateStatus(orderID, domain.OrderStatusCancelled)

	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_Delete_Success_RestoreStock(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()
	productID := uuid.New()

	orderItem := &domain.OrderItem{
		ProductID: productID,
		Quantity:  2,
	}

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusPending, // Should restore stock for pending orders
		Items:  []*domain.OrderItem{orderItem},
	}

	product := &domain.Product{
		ID:    productID,
		Stock: 3,
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockProductRepo.On("GetByID", productID).Return(product, nil)

	updatedProduct := *product
	updatedProduct.Stock = 5 // Stock restored by 2
	mockProductRepo.On("Update", &updatedProduct).Return(nil)

	mockOrderRepo.On("Delete", orderID).Return(nil)

	err := usecase.Delete(orderID)

	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_Delete_DeliveredOrder_NoStockRestore(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	orderID := uuid.New()
	productID := uuid.New()

	orderItem := &domain.OrderItem{
		ProductID: productID,
		Quantity:  2,
	}

	order := &domain.Order{
		ID:     orderID,
		Status: domain.OrderStatusDelivered, // Should NOT restore stock for delivered orders
		Items:  []*domain.OrderItem{orderItem},
	}

	mockOrderRepo.On("GetByID", orderID).Return(order, nil)
	mockOrderRepo.On("Delete", orderID).Return(nil)
	// No product repo calls expected

	err := usecase.Delete(orderID)

	assert.NoError(t, err)
	mockOrderRepo.AssertExpectations(t)
	mockProductRepo.AssertExpectations(t)
}

func TestOrderUsecase_GetAll_WithPagination(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	expectedOrders := []*domain.Order{
		{ID: uuid.New(), CustomerName: "Customer 1"},
		{ID: uuid.New(), CustomerName: "Customer 2"},
	}

	mockOrderRepo.On("GetAll", 20, 10).Return(expectedOrders, nil)

	orders, err := usecase.GetAll(20, 10)

	assert.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
	mockOrderRepo.AssertExpectations(t)
}

func TestOrderUsecase_GetAll_DefaultAndMaxLimits(t *testing.T) {
	mockOrderRepo := new(mockOrderRepository)
	mockProductRepo := new(mockProductRepository)
	usecase := NewOrderUsecase(mockOrderRepo, mockProductRepo)

	tests := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		expectedLimit  int
		expectedOffset int
	}{
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
			mockOrderRepo.On("GetAll", tt.expectedLimit, tt.expectedOffset).Return([]*domain.Order{}, nil).Once()

			_, err := usecase.GetAll(tt.inputLimit, tt.inputOffset)

			assert.NoError(t, err)
		})
	}

	mockOrderRepo.AssertExpectations(t)
}
