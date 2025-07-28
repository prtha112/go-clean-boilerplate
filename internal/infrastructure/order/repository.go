package order

import (
    "errors"
    domain "go-clean-architecture/internal/domain/order"
)

type orderRepository struct {
    orders map[int]*domain.Order
    nextID int
}

func NewOrderRepository() domain.Repository {
    return &orderRepository{
        orders: map[int]*domain.Order{},
        nextID: 1,
    }
}

func (r *orderRepository) GetByID(id int) (*domain.Order, error) {
    if order, ok := r.orders[id]; ok {
        return order, nil
    }
    return nil, errors.New("order not found")
}

func (r *orderRepository) Create(order *domain.Order) error {
    order.ID = r.nextID
    r.orders[order.ID] = order
    r.nextID++
    return nil
}

func (r *orderRepository) GetAll() ([]*domain.Order, error) {
    orders := make([]*domain.Order, 0, len(r.orders))
    for _, o := range r.orders {
        orders = append(orders, o)
    }
    return orders, nil
}

func (r *orderRepository) DeleteByID(id int) error {
    if _, ok := r.orders[id]; !ok {
        return errors.New("order not found")
    }
    delete(r.orders, id)
    return nil
}
