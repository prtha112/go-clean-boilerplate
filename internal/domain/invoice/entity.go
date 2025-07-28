package invoice

type Invoice struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	CreatedAt int64   `json:"created_at"`
}
