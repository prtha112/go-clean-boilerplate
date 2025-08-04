package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/google/uuid"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) domain.OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(order *domain.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	order.ID = uuid.New()
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Insert order
	orderQuery := `
		INSERT INTO orders (id, customer_name, customer_email, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(orderQuery, order.ID, order.CustomerName, order.CustomerEmail, order.Status, order.TotalAmount, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	// Insert order items
	for _, item := range order.Items {
		item.ID = uuid.New()
		item.OrderID = order.ID
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		if err := r.createOrderItemTx(tx, item); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepository) GetByID(id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, customer_name, customer_email, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order := &domain.Order{}
	err := r.db.QueryRow(query, id).Scan(
		&order.ID,
		&order.CustomerName,
		&order.CustomerEmail,
		&order.Status,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	items, err := r.GetOrderItems(order.ID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (r *orderRepository) GetAll(limit, offset int) ([]*domain.Order, error) {
	query := `
		SELECT id, customer_name, customer_email, status, total_amount, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.CustomerName,
			&order.CustomerEmail,
			&order.Status,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Get order items for each order
		items, err := r.GetOrderItems(order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) Update(order *domain.Order) error {
	order.UpdatedAt = time.Now()

	query := `
		UPDATE orders
		SET customer_name = $2, customer_email = $3, status = $4, total_amount = $5, updated_at = $6
		WHERE id = $1
	`

	result, err := r.db.Exec(query, order.ID, order.CustomerName, order.CustomerEmail, order.Status, order.TotalAmount, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM orders WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

func (r *orderRepository) CreateOrderItem(item *domain.OrderItem) error {
	return r.createOrderItemTx(r.db, item)
}

func (r *orderRepository) createOrderItemTx(tx interface{}, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	var err error
	switch t := tx.(type) {
	case *sql.Tx:
		_, err = t.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price, item.CreatedAt, item.UpdatedAt)
	case *sql.DB:
		_, err = t.Exec(query, item.ID, item.OrderID, item.ProductID, item.Quantity, item.Price, item.CreatedAt, item.UpdatedAt)
	default:
		return fmt.Errorf("invalid transaction type")
	}

	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}

	return nil
}

func (r *orderRepository) GetOrderItems(orderID uuid.UUID) ([]*domain.OrderItem, error) {
	query := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.created_at, oi.updated_at,
		       p.name, p.description, p.price as product_price, p.stock
		FROM order_items oi
		LEFT JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
		ORDER BY oi.created_at
	`

	rows, err := r.db.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		item := &domain.OrderItem{}
		product := &domain.Product{}

		var productName, productDescription sql.NullString
		var productPrice sql.NullFloat64
		var productStock sql.NullInt32

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price,
			&item.CreatedAt,
			&item.UpdatedAt,
			&productName,
			&productDescription,
			&productPrice,
			&productStock,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}

		// Set product details if available
		if productName.Valid {
			product.ID = item.ProductID
			product.Name = productName.String
			product.Description = productDescription.String
			product.Price = productPrice.Float64
			product.Stock = int(productStock.Int32)
			item.Product = product
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate order items: %w", err)
	}

	return items, nil
}
