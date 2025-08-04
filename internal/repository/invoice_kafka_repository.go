package repository

import (
	"database/sql"
	"go-clean-v2/internal/domain"
	"log"
)

type InvoicePostgres struct {
	db *sql.DB
}

func NewInvoicePostgres(db *sql.DB) *InvoicePostgres {
	return &InvoicePostgres{db: db}
}

func (r *InvoicePostgres) Save(inv *domain.InvoiceKafka) error {
	_, err := r.db.Exec("INSERT INTO invoices_log (id, amount) VALUES ($1, $2)", inv.ID, inv.Amount)
	log.Printf("Saving invoice with ID: %d and Amount: %.2f", inv.ID, inv.Amount)
	return err
}
