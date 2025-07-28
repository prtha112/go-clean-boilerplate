package order

type Order struct {
    ID     int    `json:"id"`
    Item   string `json:"item"`
    Amount int    `json:"amount"`
}
