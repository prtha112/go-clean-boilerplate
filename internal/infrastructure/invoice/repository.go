package invoice

import (
	"context"
	"database/sql"
	domain "go-clean-architecture/internal/domain/invoice"
)

type PostgresInvoiceRepository struct {
	DB *sql.DB
}

func NewPostgresInvoiceRepository(db *sql.DB) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{DB: db}
}

func (r *PostgresInvoiceRepository) CreateInvoice(invoice *domain.Invoice) error {
	_, err := r.DB.ExecContext(context.Background(),
		`INSERT INTO invoices (id, order_id, amount, created_at) VALUES ($1, $2, $3, $4)`,
		invoice.ID, invoice.OrderID, invoice.Amount, invoice.CreatedAt,
	)
	return err
}

var _ domain.Repository = (*PostgresInvoiceRepository)(nil)
