package usecase

import (
	"fmt"

	"go-clean-boilerplate/internal/domain"

	"github.com/google/uuid"
)

type orderUsecase struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
}

func NewOrderUsecase(orderRepo domain.OrderRepository, productRepo domain.ProductRepository) domain.OrderUsecase {
	return &orderUsecase{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (u *orderUsecase) Create(req *domain.CreateOrderRequest) (*domain.Order, error) {
	if req.CustomerName == "" {
		return nil, fmt.Errorf("customer name is required")
	}

	if req.CustomerEmail == "" {
		return nil, fmt.Errorf("customer email is required")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	order := &domain.Order{
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		Status:        domain.OrderStatusPending,
		TotalAmount:   0,
		Items:         make([]*domain.OrderItem, 0),
	}

	var totalAmount float64

	// Process each item
	for _, reqItem := range req.Items {
		// Get product details
		product, err := u.productRepo.GetByID(reqItem.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %s not found: %w", reqItem.ProductID, err)
		}

		// Check stock availability
		if product.Stock < reqItem.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s. Available: %d, Requested: %d",
				product.Name, product.Stock, reqItem.Quantity)
		}

		// Create order item
		orderItem := &domain.OrderItem{
			ProductID: reqItem.ProductID,
			Quantity:  reqItem.Quantity,
			Price:     product.Price,
			Product:   product,
		}

		order.Items = append(order.Items, orderItem)
		totalAmount += product.Price * float64(reqItem.Quantity)

		// Update product stock
		product.Stock -= reqItem.Quantity
		if err := u.productRepo.Update(product); err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	order.TotalAmount = totalAmount

	// Create the order
	if err := u.orderRepo.Create(order); err != nil {
		// Rollback stock changes if order creation fails
		for _, item := range order.Items {
			product, _ := u.productRepo.GetByID(item.ProductID)
			if product != nil {
				product.Stock += item.Quantity
				u.productRepo.Update(product)
			}
		}
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

func (u *orderUsecase) GetByID(id uuid.UUID) (*domain.Order, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("invalid order ID")
	}

	return u.orderRepo.GetByID(id)
}

func (u *orderUsecase) GetAll(limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 10 // default limit
	}

	if limit > 100 {
		limit = 100 // max limit
	}

	if offset < 0 {
		offset = 0
	}

	return u.orderRepo.GetAll(limit, offset)
}

func (u *orderUsecase) Update(order *domain.Order) error {
	if order.ID == uuid.Nil {
		return fmt.Errorf("invalid order ID")
	}

	if order.CustomerName == "" {
		return fmt.Errorf("customer name is required")
	}

	if order.CustomerEmail == "" {
		return fmt.Errorf("customer email is required")
	}

	// Check if order exists
	_, err := u.orderRepo.GetByID(order.ID)
	if err != nil {
		return err
	}

	return u.orderRepo.Update(order)
}

func (u *orderUsecase) Delete(id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid order ID")
	}

	// Get order details to restore stock
	order, err := u.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Only restore stock if order is not delivered or cancelled
	if order.Status == domain.OrderStatusPending || order.Status == domain.OrderStatusConfirmed {
		for _, item := range order.Items {
			product, err := u.productRepo.GetByID(item.ProductID)
			if err == nil {
				product.Stock += item.Quantity
				u.productRepo.Update(product)
			}
		}
	}

	return u.orderRepo.Delete(id)
}

func (u *orderUsecase) UpdateStatus(id uuid.UUID, status domain.OrderStatus) error {
	if id == uuid.Nil {
		return fmt.Errorf("invalid order ID")
	}

	// Validate status
	validStatuses := map[domain.OrderStatus]bool{
		domain.OrderStatusPending:   true,
		domain.OrderStatusConfirmed: true,
		domain.OrderStatusShipped:   true,
		domain.OrderStatusDelivered: true,
		domain.OrderStatusCancelled: true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid order status: %s", status)
	}

	order, err := u.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	// If cancelling order, restore stock
	if status == domain.OrderStatusCancelled && order.Status != domain.OrderStatusCancelled {
		for _, item := range order.Items {
			product, err := u.productRepo.GetByID(item.ProductID)
			if err == nil {
				product.Stock += item.Quantity
				u.productRepo.Update(product)
			}
		}
	}

	order.Status = status
	return u.orderRepo.Update(order)
}
