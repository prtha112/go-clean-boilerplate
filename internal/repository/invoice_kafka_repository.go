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
	_, err := r.db.Exec("INSERT INTO invoice_item_logs (invoice_id, product_id, description, quantity, unit_price, total_price) VALUES ($1, $2, $3, $4, $5, $6)",
		inv.InvoiceID, inv.ProductID, inv.Description, inv.Quantity,
		inv.UnitPrice, inv.TotalPrice)
	if err != nil {
		log.Printf("Error saving invoice with ID: %d, Error: %v", inv.InvoiceID, err)
		return err
	}
	log.Printf("Saving invoice with raw data: %+v", inv)
	return nil
}
