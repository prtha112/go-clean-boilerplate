package order

type Repository interface {
    GetByID(id int) (*Order, error)
    Create(order *Order) error
    GetAll() ([]*Order, error)
    DeleteByID(id int) error
}
