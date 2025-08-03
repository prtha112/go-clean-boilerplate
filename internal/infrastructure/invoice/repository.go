package invoice

import (
	"context"
	"database/sql"
	"encoding/json"
	domain "go-clean-architecture/internal/domain/invoice"
	"log"

	"github.com/segmentio/kafka-go"
)

type PostgresInvoiceRepository struct {
	DB     *sql.DB
	Writer *kafka.Writer
}

func NewPostgresInvoiceRepository(db *sql.DB, writer *kafka.Writer) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{DB: db, Writer: writer}
}

func (r *PostgresInvoiceRepository) CreateInvoice(invoice *domain.Invoice) error {
	_, err := r.DB.ExecContext(context.Background(),
		`INSERT INTO invoices (order_id, amount) VALUES ($1, $2)`,
		invoice.OrderID, invoice.Amount,
	)
	if err != nil {
		// log error and input for investigation
		log.Printf("CreateInvoice error: %v | order_id=%v amount=%v", err, invoice.OrderID, invoice.Amount)
	} else {
		log.Printf("CreateInvoice success | order_id=%v amount=%v", invoice.OrderID, invoice.Amount)
	}
	return err
}

// PublishInvoiceMessage implements invoice.Repository.
func (r *PostgresInvoiceRepository) PublishInvoiceMessage(msg []byte) error {
	// Use inv.ID as key if possible, else empty
	var key []byte
	var inv domain.Invoice
	if err := json.Unmarshal(msg, &inv); err == nil {
		key = []byte(inv.ID)
	}

	err := r.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   key,
		Value: msg,
	})
	if err != nil {
		log.Printf("PublishInvoiceMessage error: %v", err)
		return err
	}
	log.Printf("PublishInvoiceMessage success: %s", string(msg))
	return nil
}

var _ domain.Repository = (*PostgresInvoiceRepository)(nil)
