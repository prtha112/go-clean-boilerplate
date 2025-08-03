package invoice

import (
	"context"
	"database/sql"
	"encoding/json"
	domain "go-clean-architecture/internal/domain/invoice"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
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
func (r *PostgresInvoiceRepository) PublishInvoiceMessage(ctx context.Context, msg []byte) error {
	// Use inv.ID as key if possible, else empty
	tr := otel.Tracer("invoiceRepository")
	ctx, span := tr.Start(ctx, "PublishInvoiceMessage")
	defer span.End()
	var key []byte
	var inv domain.Invoice

	// ✅ Unmarshal invoice to get ID for key
	func() {
		_, span := tr.Start(ctx, "Unmarshal invoice")
		defer span.End()

		time.Sleep(5 * time.Second)
		if err := json.Unmarshal(msg, &inv); err == nil {
			key = []byte(inv.ID)
		}
	}()

	// ✅ Start Kafka span
	func() {
		_, span := tr.Start(ctx, "Write message to Kafka")
		defer span.End()

		err := r.Writer.WriteMessages(ctx, kafka.Message{
			Key:   key,
			Value: msg,
		})
		if err != nil {
			log.Printf("PublishInvoiceMessage error: %v", err)
		} else {
			log.Printf("PublishInvoiceMessage success: %s", string(msg))
		}
	}()
	return nil
}

var _ domain.Repository = (*PostgresInvoiceRepository)(nil)
