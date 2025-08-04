package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatus_Constants(t *testing.T) {
	assert.Equal(t, OrderStatus("pending"), OrderStatusPending)
	assert.Equal(t, OrderStatus("confirmed"), OrderStatusConfirmed)
	assert.Equal(t, OrderStatus("shipped"), OrderStatusShipped)
	assert.Equal(t, OrderStatus("delivered"), OrderStatusDelivered)
	assert.Equal(t, OrderStatus("cancelled"), OrderStatusCancelled)
}

func TestOrderItem_Creation(t *testing.T) {
	productID := uuid.New()
	orderID := uuid.New()

	item := &OrderItem{
		ID:        uuid.New(),
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  2,
		Price:     99.99,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.Equal(t, orderID, item.OrderID)
	assert.Equal(t, productID, item.ProductID)
	assert.Equal(t, 2, item.Quantity)
	assert.Equal(t, 99.99, item.Price)
}

func TestOrder_Creation(t *testing.T) {
	order := &Order{
		ID:            uuid.New(),
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Status:        OrderStatusPending,
		TotalAmount:   199.98,
		Items:         make([]*OrderItem, 0),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, order.ID)
	assert.Equal(t, "John Doe", order.CustomerName)
	assert.Equal(t, "john@example.com", order.CustomerEmail)
	assert.Equal(t, OrderStatusPending, order.Status)
	assert.Equal(t, 199.98, order.TotalAmount)
	assert.NotNil(t, order.Items)
	assert.Len(t, order.Items, 0)
}

func TestCreateOrderRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateOrderRequest
		hasError bool
	}{
		{
			name: "valid order request",
			request: CreateOrderRequest{
				CustomerName:  "John Doe",
				CustomerEmail: "john@example.com",
				Items: []struct {
					ProductID uuid.UUID `json:"product_id" binding:"required"`
					Quantity  int       `json:"quantity" binding:"required,min=1"`
				}{
					{
						ProductID: uuid.New(),
						Quantity:  2,
					},
				},
			},
			hasError: false,
		},
		{
			name: "empty customer name",
			request: CreateOrderRequest{
				CustomerName:  "",
				CustomerEmail: "john@example.com",
				Items: []struct {
					ProductID uuid.UUID `json:"product_id" binding:"required"`
					Quantity  int       `json:"quantity" binding:"required,min=1"`
				}{
					{
						ProductID: uuid.New(),
						Quantity:  2,
					},
				},
			},
			hasError: true,
		},
		{
			name: "invalid email",
			request: CreateOrderRequest{
				CustomerName:  "John Doe",
				CustomerEmail: "invalid-email",
				Items: []struct {
					ProductID uuid.UUID `json:"product_id" binding:"required"`
					Quantity  int       `json:"quantity" binding:"required,min=1"`
				}{
					{
						ProductID: uuid.New(),
						Quantity:  2,
					},
				},
			},
			hasError: true,
		},
		{
			name: "no items",
			request: CreateOrderRequest{
				CustomerName:  "John Doe",
				CustomerEmail: "john@example.com",
				Items: []struct {
					ProductID uuid.UUID `json:"product_id" binding:"required"`
					Quantity  int       `json:"quantity" binding:"required,min=1"`
				}{},
			},
			hasError: true,
		},
		{
			name: "zero quantity",
			request: CreateOrderRequest{
				CustomerName:  "John Doe",
				CustomerEmail: "john@example.com",
				Items: []struct {
					ProductID uuid.UUID `json:"product_id" binding:"required"`
					Quantity  int       `json:"quantity" binding:"required,min=1"`
				}{
					{
						ProductID: uuid.New(),
						Quantity:  0,
					},
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation
			hasError := tt.request.CustomerName == "" ||
				!isValidEmail(tt.request.CustomerEmail) ||
				len(tt.request.Items) == 0 ||
				hasInvalidQuantity(tt.request.Items)

			assert.Equal(t, tt.hasError, hasError)
		})
	}
}

func TestOrder_WithItems(t *testing.T) {
	productID1 := uuid.New()
	productID2 := uuid.New()
	orderID := uuid.New()

	item1 := &OrderItem{
		ID:        uuid.New(),
		OrderID:   orderID,
		ProductID: productID1,
		Quantity:  2,
		Price:     50.00,
	}

	item2 := &OrderItem{
		ID:        uuid.New(),
		OrderID:   orderID,
		ProductID: productID2,
		Quantity:  1,
		Price:     99.99,
	}

	order := &Order{
		ID:            orderID,
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Status:        OrderStatusPending,
		TotalAmount:   199.99,
		Items:         []*OrderItem{item1, item2},
	}

	assert.Len(t, order.Items, 2)
	assert.Equal(t, item1, order.Items[0])
	assert.Equal(t, item2, order.Items[1])
	assert.Equal(t, 199.99, order.TotalAmount)
}

// Helper functions
func hasInvalidQuantity(items []struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}) bool {
	for _, item := range items {
		if item.Quantity < 1 {
			return true
		}
	}
	return false
}
