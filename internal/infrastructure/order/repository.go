package order

import (
	"database/sql"
	"errors"
	domain "go-clean-architecture/internal/domain/order"

	_ "github.com/lib/pq"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) domain.Repository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) DeleteByID(id int) error {
	res, err := r.db.Exec("DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *OrderRepository) Create(order *domain.Order) error {
	err := r.db.QueryRow(
		"INSERT INTO orders (item, amount) VALUES ($1, $2) RETURNING id",
		order.Item, order.Amount,
	).Scan(&order.ID)
	return err
}

func (r *OrderRepository) GetAll() ([]*domain.Order, error) {
	rows, err := r.db.Query("SELECT id, item, amount FROM orders")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.Item, &o.Amount); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) GetByID(id int) (*domain.Order, error) {
	var order domain.Order
	err := r.db.QueryRow("SELECT id, item, amount FROM orders WHERE id = $1", id).Scan(&order.ID, &order.Item, &order.Amount)
	if err == sql.ErrNoRows {
		return nil, errors.New("order not found")
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}
