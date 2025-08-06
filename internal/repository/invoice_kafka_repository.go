package repository

import (
	"database/sql"
	"go-clean-boilerplate/internal/domain"
	"log"
)

type invoicePostgres struct {
	db *sql.DB
}

func NewInvoicePostgres(db *sql.DB) *invoicePostgres {
	return &invoicePostgres{db: db}
}

func (r *invoicePostgres) Save(inv *domain.InvoiceKafka) error {
	_, err := r.db.Exec("INSERT INTO invoices_log (id, amount) VALUES ($1, $2)", inv.ID, inv.Amount)
	log.Printf("Saving invoice with ID: %d and Amount: %.2f", inv.ID, inv.Amount)
	return err
}
