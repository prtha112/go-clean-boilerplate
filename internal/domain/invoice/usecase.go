package invoice

type UseCase interface {
	ConsumeInvoiceMessage(msg []byte) error
}
