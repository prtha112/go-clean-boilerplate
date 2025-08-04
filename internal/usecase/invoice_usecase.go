package usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/google/uuid"
)

type invoiceUsecase struct {
	invoiceRepo   domain.InvoiceRepository
	orderRepo     domain.OrderRepository
	productRepo   domain.ProductRepository
	kafkaProducer domain.KafkaProducer
}

func NewInvoiceUsecase(invoiceRepo domain.InvoiceRepository, orderRepo domain.OrderRepository, productRepo domain.ProductRepository, kafkaProducer domain.KafkaProducer) domain.InvoiceUsecase {
	return &invoiceUsecase{
		invoiceRepo:   invoiceRepo,
		orderRepo:     orderRepo,
		productRepo:   productRepo,
		kafkaProducer: kafkaProducer,
	}
}

func (u *invoiceUsecase) Create(req *domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	// Generate invoice number
	invoiceNumber, err := u.invoiceRepo.GenerateInvoiceNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice
	invoice := &domain.Invoice{
		InvoiceNumber:  invoiceNumber,
		CustomerName:   req.CustomerName,
		CustomerEmail:  req.CustomerEmail,
		CustomerPhone:  req.CustomerPhone,
		BillingAddress: req.BillingAddress,
		Status:         domain.InvoiceStatusDraft,
		DueDate:        req.DueDate,
		IssuedDate:     time.Now(),
		Notes:          req.Notes,
		Items:          make([]*domain.InvoiceItem, 0),
	}

	var subTotal float64

	// Process each item
	for _, reqItem := range req.Items {
		item := &domain.InvoiceItem{
			Description: reqItem.Description,
			Quantity:    reqItem.Quantity,
			UnitPrice:   reqItem.UnitPrice,
			TotalPrice:  reqItem.UnitPrice * float64(reqItem.Quantity),
		}

		// If product ID is provided, get product details
		if reqItem.ProductID != nil {
			product, err := u.productRepo.GetByID(*reqItem.ProductID)
			if err != nil {
				return nil, fmt.Errorf("product %s not found: %w", reqItem.ProductID, err)
			}
			item.ProductID = *reqItem.ProductID
			item.Product = product
		}

		invoice.Items = append(invoice.Items, item)
		subTotal += item.TotalPrice
	}

	// Calculate totals
	invoice.SubTotal = subTotal
	invoice.TaxAmount = subTotal * req.TaxRate
	invoice.TotalAmount = invoice.SubTotal + invoice.TaxAmount

	// Save to database
	if err := u.invoiceRepo.Create(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Send to Kafka
	if err := u.SendToKafka(invoice, "invoice_created"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send invoice to Kafka: %v\n", err)
	}

	return invoice, nil
}

func (u *invoiceUsecase) CreateFromOrder(orderID uuid.UUID, req *domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	// Get order details
	order, err := u.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	// Generate invoice number
	invoiceNumber, err := u.invoiceRepo.GenerateInvoiceNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Create invoice from order
	invoice := &domain.Invoice{
		InvoiceNumber:  invoiceNumber,
		OrderID:        &orderID,
		Order:          order,
		CustomerName:   req.CustomerName,
		CustomerEmail:  req.CustomerEmail,
		CustomerPhone:  req.CustomerPhone,
		BillingAddress: req.BillingAddress,
		Status:         domain.InvoiceStatusDraft,
		DueDate:        req.DueDate,
		IssuedDate:     time.Now(),
		Notes:          req.Notes,
		Items:          make([]*domain.InvoiceItem, 0),
	}

	var subTotal float64

	// Convert order items to invoice items
	for _, orderItem := range order.Items {
		invoiceItem := &domain.InvoiceItem{
			ProductID:   orderItem.ProductID,
			Product:     orderItem.Product,
			Description: orderItem.Product.Name,
			Quantity:    orderItem.Quantity,
			UnitPrice:   orderItem.Price,
			TotalPrice:  orderItem.Price * float64(orderItem.Quantity),
		}

		invoice.Items = append(invoice.Items, invoiceItem)
		subTotal += invoiceItem.TotalPrice
	}

	// Calculate totals
	invoice.SubTotal = subTotal
	invoice.TaxAmount = subTotal * req.TaxRate
	invoice.TotalAmount = invoice.SubTotal + invoice.TaxAmount

	// Save to database
	if err := u.invoiceRepo.Create(invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Send to Kafka
	if err := u.SendToKafka(invoice, "invoice_created_from_order"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send invoice to Kafka: %v\n", err)
	}

	return invoice, nil
}

func (u *invoiceUsecase) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid invoice ID")
	}

	return u.invoiceRepo.GetByID(id)
}

func (u *invoiceUsecase) GetByInvoiceNumber(invoiceNumber string) (*domain.Invoice, error) {
	if invoiceNumber == "" {
		return nil, fmt.Errorf("invoice number is required")
	}

	return u.invoiceRepo.GetByInvoiceNumber(invoiceNumber)
}

func (u *invoiceUsecase) GetAll(limit, offset int) ([]*domain.Invoice, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	if limit > 100 {
		limit = 100 // max limit
	}

	if offset < 0 {
		offset = 0
	}

	return u.invoiceRepo.GetAll(limit, offset)
}

func (u *invoiceUsecase) GetByStatus(status domain.InvoiceStatus, limit, offset int) ([]*domain.Invoice, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	if limit > 100 {
		limit = 100 // max limit
	}

	if offset < 0 {
		offset = 0
	}

	return u.invoiceRepo.GetByStatus(status, limit, offset)
}

func (u *invoiceUsecase) UpdateStatus(id uuid.UUID, status domain.InvoiceStatus) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid invoice ID")
	}

	// Validate status
	validStatuses := map[domain.InvoiceStatus]bool{
		domain.InvoiceStatusDraft:     true,
		domain.InvoiceStatusSent:      true,
		domain.InvoiceStatusPaid:      true,
		domain.InvoiceStatusOverdue:   true,
		domain.InvoiceStatusCancelled: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid invoice status: %s", status)
	}

	invoice, err := u.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}

	oldStatus := invoice.Status
	invoice.Status = status

	// Set paid date if status is paid
	if status == domain.InvoiceStatusPaid && oldStatus != domain.InvoiceStatusPaid {
		now := time.Now()
		invoice.PaidDate = &now
	}

	if err := u.invoiceRepo.Update(invoice); err != nil {
		return err
	}

	// Send status update to Kafka
	eventType := fmt.Sprintf("invoice_status_updated_%s", status)
	if err := u.SendToKafka(invoice, eventType); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send invoice status update to Kafka: %v\n", err)
	}

	return nil
}

func (u *invoiceUsecase) Delete(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid invoice ID")
	}

	// Get invoice before deletion for Kafka message
	invoice, err := u.invoiceRepo.GetByID(id)
	if err != nil {
		return err
	}

	if err := u.invoiceRepo.Delete(id); err != nil {
		return err
	}

	// Send deletion event to Kafka
	if err := u.SendToKafka(invoice, "invoice_deleted"); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to send invoice deletion to Kafka: %v\n", err)
	}

	return nil
}

func (u *invoiceUsecase) SendToKafka(invoice *domain.Invoice, eventType string) error {
	if u.kafkaProducer == nil {
		return fmt.Errorf("kafka producer not configured")
	}

	// Create invoice event
	event := &domain.InvoiceEvent{
		EventType: eventType,
		InvoiceID: invoice.ID,
		Invoice:   invoice,
		Timestamp: time.Now(),
	}

	// Marshal to JSON
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal invoice event: %w", err)
	}

	// Send to Kafka topic
	topic := "invoices"
	key := invoice.ID.String()

	if err := u.kafkaProducer.SendMessage(topic, key, eventData); err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}
