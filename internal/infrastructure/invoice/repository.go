package invoice

import (
	"context"
	"database/sql"
	domain "go-clean-architecture/internal/domain/invoice"
	"log"
)

type PostgresInvoiceRepository struct {
	DB *sql.DB
}

func NewPostgresInvoiceRepository(db *sql.DB) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{DB: db}
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

var _ domain.Repository = (*PostgresInvoiceRepository)(nil)
