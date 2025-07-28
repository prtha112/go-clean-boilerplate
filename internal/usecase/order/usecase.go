package order

import domain "go-clean-architecture/internal/domain/order"

type orderUseCase struct {
	repo domain.Repository
}

type UseCase interface {
	GetOrder(id int) (*domain.Order, error)
	CreateOrder(order *domain.Order) error
	GetAllOrders() ([]*domain.Order, error)
	DeleteOrder(id int) error
}

func NewOrderUseCase(r domain.Repository) UseCase {
	return &orderUseCase{repo: r}
}

func (uc *orderUseCase) GetOrder(id int) (*domain.Order, error) {
	return uc.repo.GetByID(id)
}

func (uc *orderUseCase) CreateOrder(o *domain.Order) error {
	return uc.repo.Create(o)
}

func (uc *orderUseCase) GetAllOrders() ([]*domain.Order, error) {
	return uc.repo.GetAll()
}

func (uc *orderUseCase) DeleteOrder(id int) error {
	return uc.repo.DeleteByID(id)
}
