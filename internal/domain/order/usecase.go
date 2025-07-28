package order

type UseCase interface {
    GetOrder(id int) (*Order, error)
    CreateOrder(order *Order) error
    GetAllOrders() ([]*Order, error)
    DeleteOrder(id int) error
}
