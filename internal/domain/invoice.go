package domain

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

type InvoiceItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id"`
	ProductID   uuid.UUID `json:"product_id" db:"product_id"`
	Product     *Product  `json:"product,omitempty"`
	Description string    `json:"description" db:"description"`
	Quantity    int       `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TotalPrice  float64   `json:"total_price" db:"total_price"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Invoice struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	InvoiceNumber  string         `json:"invoice_number" db:"invoice_number"`
	OrderID        *uuid.UUID     `json:"order_id,omitempty" db:"order_id"`
	Order          *Order         `json:"order,omitempty"`
	CustomerName   string         `json:"customer_name" db:"customer_name"`
	CustomerEmail  string         `json:"customer_email" db:"customer_email"`
	CustomerPhone  string         `json:"customer_phone" db:"customer_phone"`
	BillingAddress string         `json:"billing_address" db:"billing_address"`
	Status         InvoiceStatus  `json:"status" db:"status"`
	SubTotal       float64        `json:"sub_total" db:"sub_total"`
	TaxAmount      float64        `json:"tax_amount" db:"tax_amount"`
	TotalAmount    float64        `json:"total_amount" db:"total_amount"`
	DueDate        time.Time      `json:"due_date" db:"due_date"`
	IssuedDate     time.Time      `json:"issued_date" db:"issued_date"`
	PaidDate       *time.Time     `json:"paid_date,omitempty" db:"paid_date"`
	Items          []*InvoiceItem `json:"items,omitempty"`
	Notes          string         `json:"notes" db:"notes"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type CreateInvoiceRequest struct {
	OrderID        *uuid.UUID `json:"order_id,omitempty"`
	CustomerName   string     `json:"customer_name" binding:"required"`
	CustomerEmail  string     `json:"customer_email" binding:"required,email"`
	CustomerPhone  string     `json:"customer_phone"`
	BillingAddress string     `json:"billing_address" binding:"required"`
	DueDate        time.Time  `json:"due_date" binding:"required"`
	TaxRate        float64    `json:"tax_rate" binding:"min=0,max=1"` // 0.0 to 1.0 (0% to 100%)
	Notes          string     `json:"notes"`
	Items          []struct {
		ProductID   *uuid.UUID `json:"product_id,omitempty"`
		Description string     `json:"description" binding:"required"`
		Quantity    int        `json:"quantity" binding:"required,min=1"`
		UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
	} `json:"items" binding:"required,min=1"`
}

type UpdateInvoiceStatusRequest struct {
	Status InvoiceStatus `json:"status" binding:"required"`
}

type InvoiceEvent struct {
	EventType string    `json:"event_type"`
	InvoiceID uuid.UUID `json:"invoice_id"`
	Invoice   *Invoice  `json:"invoice"`
	Timestamp time.Time `json:"timestamp"`
}

type InvoiceRepository interface {
	Create(invoice *Invoice) error
	GetByID(id uuid.UUID) (*Invoice, error)
	GetByInvoiceNumber(invoiceNumber string) (*Invoice, error)
	GetAll(limit, offset int) ([]*Invoice, error)
	GetByStatus(status InvoiceStatus, limit, offset int) ([]*Invoice, error)
	Update(invoice *Invoice) error
	Delete(id uuid.UUID) error
	CreateInvoiceItem(item *InvoiceItem) error
	GetInvoiceItems(invoiceID uuid.UUID) ([]*InvoiceItem, error)
	GenerateInvoiceNumber() (string, error)
}

type InvoiceUsecase interface {
	Create(req *CreateInvoiceRequest) (*Invoice, error)
	CreateFromOrder(orderID uuid.UUID, req *CreateInvoiceRequest) (*Invoice, error)
	GetByID(id uuid.UUID) (*Invoice, error)
	GetByInvoiceNumber(invoiceNumber string) (*Invoice, error)
	GetAll(limit, offset int) ([]*Invoice, error)
	GetByStatus(status InvoiceStatus, limit, offset int) ([]*Invoice, error)
	UpdateStatus(id uuid.UUID, status InvoiceStatus) error
	Delete(id uuid.UUID) error
	SendToKafka(invoice *Invoice, eventType string) error
}

type KafkaProducer interface {
	SendMessage(topic string, key string, message []byte) error
	Close() error
}
