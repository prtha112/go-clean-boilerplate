package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OrderID   uuid.UUID `json:"order_id" db:"order_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity" db:"quantity"`
	Price     float64   `json:"price" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Order struct {
	ID            uuid.UUID    `json:"id" db:"id"`
	CustomerName  string       `json:"customer_name" db:"customer_name"`
	CustomerEmail string       `json:"customer_email" db:"customer_email"`
	Status        OrderStatus  `json:"status" db:"status"`
	TotalAmount   float64      `json:"total_amount" db:"total_amount"`
	Items         []*OrderItem `json:"items,omitempty"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
}

type CreateOrderRequest struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
	Items         []struct {
		ProductID uuid.UUID `json:"product_id" binding:"required"`
		Quantity  int       `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required,min=1"`
}

type OrderRepository interface {
	Create(order *Order) error
	GetByID(id uuid.UUID) (*Order, error)
	GetAll(limit, offset int) ([]*Order, error)
	Update(order *Order) error
	Delete(id uuid.UUID) error
	CreateOrderItem(item *OrderItem) error
	GetOrderItems(orderID uuid.UUID) ([]*OrderItem, error)
}

type OrderUsecase interface {
	Create(req *CreateOrderRequest) (*Order, error)
	GetByID(id uuid.UUID) (*Order, error)
	GetAll(limit, offset int) ([]*Order, error)
	Update(order *Order) error
	Delete(id uuid.UUID) error
	UpdateStatus(id uuid.UUID, status OrderStatus) error
}
