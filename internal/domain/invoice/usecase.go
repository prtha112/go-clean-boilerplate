package invoice

import "context"

type UseCase interface {
	ConsumeInvoiceMessage(msg []byte) error
	ProduceInvoiceMessage(ctx context.Context, invoice *Invoice) error
}
