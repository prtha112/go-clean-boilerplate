package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInvoiceStatus_Constants(t *testing.T) {
	assert.Equal(t, InvoiceStatus("draft"), InvoiceStatusDraft)
	assert.Equal(t, InvoiceStatus("sent"), InvoiceStatusSent)
	assert.Equal(t, InvoiceStatus("paid"), InvoiceStatusPaid)
	assert.Equal(t, InvoiceStatus("overdue"), InvoiceStatusOverdue)
	assert.Equal(t, InvoiceStatus("cancelled"), InvoiceStatusCancelled)
}

func TestInvoiceItem_Creation(t *testing.T) {
	invoiceID := uuid.New()
	productID := uuid.New()

	item := &InvoiceItem{
		ID:          uuid.New(),
		InvoiceID:   invoiceID,
		ProductID:   productID,
		Description: "Test Product",
		Quantity:    2,
		UnitPrice:   50.00,
		TotalPrice:  100.00,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, item.ID)
	assert.Equal(t, invoiceID, item.InvoiceID)
	assert.Equal(t, productID, item.ProductID)
	assert.Equal(t, "Test Product", item.Description)
	assert.Equal(t, 2, item.Quantity)
	assert.Equal(t, 50.00, item.UnitPrice)
	assert.Equal(t, 100.00, item.TotalPrice)
}

func TestInvoiceItem_TotalPriceCalculation(t *testing.T) {
	tests := []struct {
		name      string
		quantity  int
		unitPrice float64
		expected  float64
	}{
		{
			name:      "simple calculation",
			quantity:  2,
			unitPrice: 50.00,
			expected:  100.00,
		},
		{
			name:      "decimal calculation",
			quantity:  3,
			unitPrice: 33.33,
			expected:  99.99,
		},
		{
			name:      "single item",
			quantity:  1,
			unitPrice: 75.50,
			expected:  75.50,
		},
		{
			name:      "large quantity",
			quantity:  100,
			unitPrice: 1.99,
			expected:  199.00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &InvoiceItem{
				Quantity:   tt.quantity,
				UnitPrice:  tt.unitPrice,
				TotalPrice: tt.unitPrice * float64(tt.quantity),
			}

			assert.Equal(t, tt.expected, item.TotalPrice)
		})
	}
}

func TestInvoice_Creation(t *testing.T) {
	orderID := uuid.New()
	dueDate := time.Now().AddDate(0, 0, 30) // 30 days from now
	issuedDate := time.Now()

	invoice := &Invoice{
		ID:             uuid.New(),
		InvoiceNumber:  "INV-2024-001",
		OrderID:        &orderID,
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		CustomerPhone:  "+1234567890",
		BillingAddress: "123 Main St, City, State 12345",
		Status:         InvoiceStatusDraft,
		SubTotal:       100.00,
		TaxAmount:      10.00,
		TotalAmount:    110.00,
		DueDate:        dueDate,
		IssuedDate:     issuedDate,
		Notes:          "Test invoice",
		Items:          make([]*InvoiceItem, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, invoice.ID)
	assert.Equal(t, "INV-2024-001", invoice.InvoiceNumber)
	assert.Equal(t, &orderID, invoice.OrderID)
	assert.Equal(t, "John Doe", invoice.CustomerName)
	assert.Equal(t, "john@example.com", invoice.CustomerEmail)
	assert.Equal(t, InvoiceStatusDraft, invoice.Status)
	assert.Equal(t, 100.00, invoice.SubTotal)
	assert.Equal(t, 10.00, invoice.TaxAmount)
	assert.Equal(t, 110.00, invoice.TotalAmount)
	assert.NotNil(t, invoice.Items)
	assert.Len(t, invoice.Items, 0)
}

func TestInvoice_TaxCalculation(t *testing.T) {
	tests := []struct {
		name     string
		subTotal float64
		taxRate  float64
		expected float64
	}{
		{
			name:     "10% tax",
			subTotal: 100.00,
			taxRate:  0.10,
			expected: 10.00,
		},
		{
			name:     "8.5% tax",
			subTotal: 200.00,
			taxRate:  0.085,
			expected: 17.00,
		},
		{
			name:     "no tax",
			subTotal: 150.00,
			taxRate:  0.00,
			expected: 0.00,
		},
		{
			name:     "high tax rate",
			subTotal: 50.00,
			taxRate:  0.25,
			expected: 12.50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taxAmount := tt.subTotal * tt.taxRate
			totalAmount := tt.subTotal + taxAmount

			assert.Equal(t, tt.expected, taxAmount)
			assert.Equal(t, tt.subTotal+tt.expected, totalAmount)
		})
	}
}

func TestCreateInvoiceRequest_Validation(t *testing.T) {
	dueDate := time.Now().AddDate(0, 0, 30)
	productID := uuid.New()

	tests := []struct {
		name     string
		request  CreateInvoiceRequest
		hasError bool
	}{
		{
			name: "valid request",
			request: CreateInvoiceRequest{
				CustomerName:   "John Doe",
				CustomerEmail:  "john@example.com",
				BillingAddress: "123 Main St",
				DueDate:        dueDate,
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
			},
			hasError: false,
		},
		{
			name: "empty customer name",
			request: CreateInvoiceRequest{
				CustomerName:   "",
				CustomerEmail:  "john@example.com",
				BillingAddress: "123 Main St",
				DueDate:        dueDate,
				TaxRate:        0.10,
				Items: []struct {
					ProductID   *uuid.UUID `json:"product_id,omitempty"`
					Description string     `json:"description" binding:"required"`
					Quantity    int        `json:"quantity" binding:"required,min=1"`
					UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
				}{
					{
						Description: "Test Product",
						Quantity:    2,
						UnitPrice:   50.00,
					},
				},
			},
			hasError: true,
		},
		{
			name: "invalid email",
			request: CreateInvoiceRequest{
				CustomerName:   "John Doe",
				CustomerEmail:  "invalid-email",
				BillingAddress: "123 Main St",
				DueDate:        dueDate,
				TaxRate:        0.10,
				Items: []struct {
					ProductID   *uuid.UUID `json:"product_id,omitempty"`
					Description string     `json:"description" binding:"required"`
					Quantity    int        `json:"quantity" binding:"required,min=1"`
					UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
				}{
					{
						Description: "Test Product",
						Quantity:    2,
						UnitPrice:   50.00,
					},
				},
			},
			hasError: true,
		},
		{
			name: "no items",
			request: CreateInvoiceRequest{
				CustomerName:   "John Doe",
				CustomerEmail:  "john@example.com",
				BillingAddress: "123 Main St",
				DueDate:        dueDate,
				TaxRate:        0.10,
				Items: []struct {
					ProductID   *uuid.UUID `json:"product_id,omitempty"`
					Description string     `json:"description" binding:"required"`
					Quantity    int        `json:"quantity" binding:"required,min=1"`
					UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
				}{},
			},
			hasError: true,
		},
		{
			name: "negative tax rate",
			request: CreateInvoiceRequest{
				CustomerName:   "John Doe",
				CustomerEmail:  "john@example.com",
				BillingAddress: "123 Main St",
				DueDate:        dueDate,
				TaxRate:        -0.10,
				Items: []struct {
					ProductID   *uuid.UUID `json:"product_id,omitempty"`
					Description string     `json:"description" binding:"required"`
					Quantity    int        `json:"quantity" binding:"required,min=1"`
					UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
				}{
					{
						Description: "Test Product",
						Quantity:    2,
						UnitPrice:   50.00,
					},
				},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.request.CustomerName == "" ||
				!isValidEmail(tt.request.CustomerEmail) ||
				tt.request.BillingAddress == "" ||
				len(tt.request.Items) == 0 ||
				tt.request.TaxRate < 0 ||
				tt.request.TaxRate > 1

			assert.Equal(t, tt.hasError, hasError)
		})
	}
}

func TestInvoiceEvent_Creation(t *testing.T) {
	invoiceID := uuid.New()
	invoice := &Invoice{
		ID:            invoiceID,
		InvoiceNumber: "INV-2024-001",
		Status:        InvoiceStatusDraft,
	}

	event := &InvoiceEvent{
		EventType: "invoice_created",
		InvoiceID: invoiceID,
		Invoice:   invoice,
		Timestamp: time.Now(),
	}

	assert.Equal(t, "invoice_created", event.EventType)
	assert.Equal(t, invoiceID, event.InvoiceID)
	assert.Equal(t, invoice, event.Invoice)
	assert.False(t, event.Timestamp.IsZero())
}

func TestUpdateInvoiceStatusRequest_Validation(t *testing.T) {
	tests := []struct {
		name     string
		request  UpdateInvoiceStatusRequest
		hasError bool
	}{
		{
			name: "valid status update",
			request: UpdateInvoiceStatusRequest{
				Status: InvoiceStatusSent,
			},
			hasError: false,
		},
		{
			name: "empty status",
			request: UpdateInvoiceStatusRequest{
				Status: "",
			},
			hasError: true,
		},
		{
			name: "invalid status",
			request: UpdateInvoiceStatusRequest{
				Status: "invalid_status",
			},
			hasError: true,
		},
	}

	validStatuses := map[InvoiceStatus]bool{
		InvoiceStatusDraft:     true,
		InvoiceStatusSent:      true,
		InvoiceStatusPaid:      true,
		InvoiceStatusOverdue:   true,
		InvoiceStatusCancelled: true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.request.Status == "" || !validStatuses[tt.request.Status]
			assert.Equal(t, tt.hasError, hasError)
		})
	}
}
