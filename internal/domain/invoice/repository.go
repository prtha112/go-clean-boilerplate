package invoice

import "context"

type Repository interface {
	CreateInvoice(invoice *Invoice) error
	PublishInvoiceMessage(ctx context.Context, msg []byte) error
}
