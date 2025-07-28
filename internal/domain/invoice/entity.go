package invoice

type Invoice struct {
	ID        string
	OrderID   string
	Amount    float64
	CreatedAt int64
}
