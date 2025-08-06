package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/google/uuid"
)

type invoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) domain.InvoiceRepository {
	return &invoiceRepository{
		db: db,
	}
}

func (r *invoiceRepository) Create(invoice *domain.Invoice) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	invoice.ID = uuid.New()
	invoice.CreatedAt = time.Now()
	invoice.UpdatedAt = time.Now()

	// Insert invoice
	invoiceQuery := `
		INSERT INTO invoices (id, invoice_number, order_id, customer_name, customer_email, customer_phone, 
		                     billing_address, status, sub_total, tax_amount, total_amount, due_date, 
		                     issued_date, paid_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err = tx.Exec(invoiceQuery, invoice.ID, invoice.InvoiceNumber, invoice.OrderID, invoice.CustomerName,
		invoice.CustomerEmail, invoice.CustomerPhone, invoice.BillingAddress, invoice.Status,
		invoice.SubTotal, invoice.TaxAmount, invoice.TotalAmount, invoice.DueDate,
		invoice.IssuedDate, invoice.PaidDate, invoice.Notes, invoice.CreatedAt, invoice.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	// Insert invoice items
	for _, item := range invoice.Items {
		item.ID = uuid.New()
		item.InvoiceID = invoice.ID
		item.CreatedAt = time.Now()
		item.UpdatedAt = time.Now()

		if err := r.createInvoiceItemTx(tx, item); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *invoiceRepository) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, invoice_number, order_id, customer_name, customer_email, customer_phone,
		       billing_address, status, sub_total, tax_amount, total_amount, due_date,
		       issued_date, paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE id = $1
	`

	invoice := &domain.Invoice{}
	var orderID sql.NullString
	var paidDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &orderID, &invoice.CustomerName,
		&invoice.CustomerEmail, &invoice.CustomerPhone, &invoice.BillingAddress,
		&invoice.Status, &invoice.SubTotal, &invoice.TaxAmount, &invoice.TotalAmount,
		&invoice.DueDate, &invoice.IssuedDate, &paidDate, &invoice.Notes,
		&invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Handle nullable fields
	if orderID.Valid {
		orderUUID, _ := uuid.Parse(orderID.String)
		invoice.OrderID = &orderUUID
	}
	if paidDate.Valid {
		invoice.PaidDate = &paidDate.Time
	}

	// Get invoice items
	items, err := r.GetInvoiceItems(invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return invoice, nil
}

func (r *invoiceRepository) GetByInvoiceNumber(invoiceNumber string) (*domain.Invoice, error) {
	query := `
		SELECT id, invoice_number, order_id, customer_name, customer_email, customer_phone,
		       billing_address, status, sub_total, tax_amount, total_amount, due_date,
		       issued_date, paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE invoice_number = $1
	`

	invoice := &domain.Invoice{}
	var orderID sql.NullString
	var paidDate sql.NullTime

	err := r.db.QueryRow(query, invoiceNumber).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &orderID, &invoice.CustomerName,
		&invoice.CustomerEmail, &invoice.CustomerPhone, &invoice.BillingAddress,
		&invoice.Status, &invoice.SubTotal, &invoice.TaxAmount, &invoice.TotalAmount,
		&invoice.DueDate, &invoice.IssuedDate, &paidDate, &invoice.Notes,
		&invoice.CreatedAt, &invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice not found")
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	// Handle nullable fields
	if orderID.Valid {
		orderUUID, _ := uuid.Parse(orderID.String)
		invoice.OrderID = &orderUUID
	}
	if paidDate.Valid {
		invoice.PaidDate = &paidDate.Time
	}

	// Get invoice items
	items, err := r.GetInvoiceItems(invoice.ID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items

	return invoice, nil
}

func (r *invoiceRepository) GetAll(limit, offset int) ([]*domain.Invoice, error) {
	query := `
		SELECT id, invoice_number, order_id, customer_name, customer_email, customer_phone,
		       billing_address, status, sub_total, tax_amount, total_amount, due_date,
		       issued_date, paid_date, notes, created_at, updated_at
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.queryInvoices(query, limit, offset)
}

func (r *invoiceRepository) GetByStatus(status domain.InvoiceStatus, limit, offset int) ([]*domain.Invoice, error) {
	query := `
		SELECT id, invoice_number, order_id, customer_name, customer_email, customer_phone,
		       billing_address, status, sub_total, tax_amount, total_amount, due_date,
		       issued_date, paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.queryInvoices(query, status, limit, offset)
}

func (r *invoiceRepository) Update(invoice *domain.Invoice) error {
	invoice.UpdatedAt = time.Now()

	query := `
		UPDATE invoices
		SET invoice_number = $2, order_id = $3, customer_name = $4, customer_email = $5,
		    customer_phone = $6, billing_address = $7, status = $8, sub_total = $9,
		    tax_amount = $10, total_amount = $11, due_date = $12, issued_date = $13,
		    paid_date = $14, notes = $15, updated_at = $16
		WHERE id = $1
	`

	result, err := r.db.Exec(query, invoice.ID, invoice.InvoiceNumber, invoice.OrderID,
		invoice.CustomerName, invoice.CustomerEmail, invoice.CustomerPhone,
		invoice.BillingAddress, invoice.Status, invoice.SubTotal, invoice.TaxAmount,
		invoice.TotalAmount, invoice.DueDate, invoice.IssuedDate, invoice.PaidDate,
		invoice.Notes, invoice.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

func (r *invoiceRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM invoices WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

func (r *invoiceRepository) CreateInvoiceItem(item *domain.InvoiceItem) error {
	return r.createInvoiceItemTx(r.db, item)
}

func (r *invoiceRepository) createInvoiceItemTx(tx interface{}, item *domain.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (id, invoice_id, product_id, description, quantity, unit_price, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Handle nullable product_id - if ProductID is zero UUID, set to nil
	var productID interface{}
	if item.ProductID == uuid.Nil {
		productID = nil
	} else {
		productID = item.ProductID
	}

	var err error
	switch t := tx.(type) {
	case *sql.Tx:
		_, err = t.Exec(query, item.ID, item.InvoiceID, productID, item.Description,
			item.Quantity, item.UnitPrice, item.TotalPrice, item.CreatedAt, item.UpdatedAt)
	case *sql.DB:
		_, err = t.Exec(query, item.ID, item.InvoiceID, productID, item.Description,
			item.Quantity, item.UnitPrice, item.TotalPrice, item.CreatedAt, item.UpdatedAt)
	default:
		return fmt.Errorf("invalid transaction type")
	}

	if err != nil {
		return fmt.Errorf("failed to create invoice item: %w", err)
	}

	return nil
}

func (r *invoiceRepository) GetInvoiceItems(invoiceID uuid.UUID) ([]*domain.InvoiceItem, error) {
	query := `
		SELECT ii.id, ii.invoice_id, ii.product_id, ii.description, ii.quantity, ii.unit_price, ii.total_price, ii.created_at, ii.updated_at,
		       p.name, p.description as product_description, p.price as product_price, p.stock
		FROM invoice_items ii
		LEFT JOIN products p ON ii.product_id = p.id
		WHERE ii.invoice_id = $1
		ORDER BY ii.created_at
	`

	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice items: %w", err)
	}
	defer rows.Close()

	var items []*domain.InvoiceItem
	for rows.Next() {
		item := &domain.InvoiceItem{}
		product := &domain.Product{}

		var productID sql.NullString
		var productName, productDescription sql.NullString
		var productPrice sql.NullFloat64
		var productStock sql.NullInt32

		err := rows.Scan(
			&item.ID, &item.InvoiceID, &productID, &item.Description,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice,
			&item.CreatedAt, &item.UpdatedAt,
			&productName, &productDescription, &productPrice, &productStock,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice item: %w", err)
		}

		// Set product details if available
		if productID.Valid {
			productUUID, _ := uuid.Parse(productID.String)
			item.ProductID = productUUID
			if productName.Valid {
				product.ID = productUUID
				product.Name = productName.String
				product.Description = productDescription.String
				product.Price = productPrice.Float64
				product.Stock = int(productStock.Int32)
				item.Product = product
			}
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoice items: %w", err)
	}

	return items, nil
}

func (r *invoiceRepository) GenerateInvoiceNumber() (string, error) {
	// Get current year and month (YYYYMM format)
	now := time.Now()
	yearMonth := now.Format("200601") // YYYYMM format

	// Get the next sequence number for this month
	query := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM 8) AS INTEGER)), 0) + 1
		FROM invoices
		WHERE invoice_number LIKE $1
	`

	pattern := "INV" + yearMonth + "%"
	var sequenceNum int
	err := r.db.QueryRow(query, pattern).Scan(&sequenceNum)
	if err != nil {
		return "", fmt.Errorf("failed to get sequence number: %w", err)
	}

	// Format: INV + YYYYMM + 4-digit sequence (e.g., INV2024010001)
	invoiceNumber := fmt.Sprintf("INV%s%04d", yearMonth, sequenceNum)

	return invoiceNumber, nil
}

func (r *invoiceRepository) queryInvoices(query string, args ...interface{}) ([]*domain.Invoice, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		invoice := &domain.Invoice{}
		var orderID sql.NullString
		var paidDate sql.NullTime

		err := rows.Scan(
			&invoice.ID, &invoice.InvoiceNumber, &orderID, &invoice.CustomerName,
			&invoice.CustomerEmail, &invoice.CustomerPhone, &invoice.BillingAddress,
			&invoice.Status, &invoice.SubTotal, &invoice.TaxAmount, &invoice.TotalAmount,
			&invoice.DueDate, &invoice.IssuedDate, &paidDate, &invoice.Notes,
			&invoice.CreatedAt, &invoice.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}

		// Handle nullable fields
		if orderID.Valid {
			orderUUID, _ := uuid.Parse(orderID.String)
			invoice.OrderID = &orderUUID
		}
		if paidDate.Valid {
			invoice.PaidDate = &paidDate.Time
		}

		// Get invoice items for each invoice
		items, err := r.GetInvoiceItems(invoice.ID)
		if err != nil {
			return nil, err
		}
		invoice.Items = items

		invoices = append(invoices, invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invoices: %w", err)
	}

	return invoices, nil
}
