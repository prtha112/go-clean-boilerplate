package order_test

import (
	"errors"
	"testing"

	"go-clean-architecture/internal/domain/order"
	usecase "go-clean-architecture/internal/usecase/order"

	"github.com/stretchr/testify/assert"
)

// Mock Repository
type mockRepo struct {
	GetByIDFunc    func(int) (*order.Order, error)
	CreateFunc     func(*order.Order) error
	GetAllFunc     func() ([]*order.Order, error)
	DeleteByIDFunc func(int) error
}

func (m *mockRepo) GetByID(id int) (*order.Order, error) {
	return m.GetByIDFunc(id)
}
func (m *mockRepo) Create(o *order.Order) error {
	return m.CreateFunc(o)
}
func (m *mockRepo) GetAll() ([]*order.Order, error) {
	return m.GetAllFunc()
}
func (m *mockRepo) DeleteByID(id int) error {
	return m.DeleteByIDFunc(id)
}

func TestGetOrder_Success(t *testing.T) {
	mock := &mockRepo{
		GetByIDFunc: func(id int) (*order.Order, error) {
			return &order.Order{ID: id, Item: "Apple", Amount: 5}, nil
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	ord, err := uc.GetOrder(1)
	assert.NoError(t, err)
	assert.Equal(t, "Apple", ord.Item)
}

func TestGetOrder_NotFound(t *testing.T) {
	mock := &mockRepo{
		GetByIDFunc: func(id int) (*order.Order, error) {
			return nil, errors.New("order not found")
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	ord, err := uc.GetOrder(99)
	assert.Error(t, err)
	assert.Nil(t, ord)
}

func TestCreateOrder_Success(t *testing.T) {
	mock := &mockRepo{
		CreateFunc: func(o *order.Order) error {
			o.ID = 42
			return nil
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	ord := &order.Order{Item: "Banana", Amount: 10}
	err := uc.CreateOrder(ord)
	assert.NoError(t, err)
	assert.Equal(t, 42, ord.ID)
}

func TestGetAllOrders(t *testing.T) {
	mock := &mockRepo{
		GetAllFunc: func() ([]*order.Order, error) {
			return []*order.Order{
				{ID: 1, Item: "Apple", Amount: 5},
				{ID: 2, Item: "Banana", Amount: 10},
			}, nil
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	orders, err := uc.GetAllOrders()
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, "Banana", orders[1].Item)
}

func TestDeleteOrder_Success(t *testing.T) {
	mock := &mockRepo{
		DeleteByIDFunc: func(id int) error {
			return nil
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	err := uc.DeleteOrder(1)
	assert.NoError(t, err)
}

func TestDeleteOrder_NotFound(t *testing.T) {
	mock := &mockRepo{
		DeleteByIDFunc: func(id int) error {
			return errors.New("order not found")
		},
	}
	uc := usecase.NewOrderUseCase(mock)
	err := uc.DeleteOrder(99)
	assert.Error(t, err)
	assert.Equal(t, "order not found", err.Error())
}
