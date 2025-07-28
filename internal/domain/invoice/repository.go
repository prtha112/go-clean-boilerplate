package invoice

type Repository interface {
	CreateInvoice(invoice *Invoice) error
}
