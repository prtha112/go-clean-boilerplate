package invoice

import "context"

// InvoiceConsumer is the contract for invoice consumer (domain layer)
type InvoiceConsumer interface {
	Start(ctx context.Context)
}
