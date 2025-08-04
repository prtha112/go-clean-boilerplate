package repository

import (
	"database/sql"
	"testing"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewProductRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	assert.NotNil(t, repo)
	assert.Implements(t, (*domain.ProductRepository)(nil), repo)
}

func TestProductRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mock.ExpectExec(`INSERT INTO products`).
		WithArgs(sqlmock.AnyArg(), product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(product)

	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, product.ID)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Create_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mock.ExpectExec(`INSERT INTO products`).
		WithArgs(sqlmock.AnyArg(), product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.Create(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create product")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()
	expectedProduct := &domain.Product{
		ID:          productID,
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
		AddRow(expectedProduct.ID, expectedProduct.Name, expectedProduct.Description, expectedProduct.Price, expectedProduct.Stock, expectedProduct.CreatedAt, expectedProduct.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnRows(rows)

	product, err := repo.GetByID(productID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, expectedProduct.ID, product.ID)
	assert.Equal(t, expectedProduct.Name, product.Name)
	assert.Equal(t, expectedProduct.Description, product.Description)
	assert.Equal(t, expectedProduct.Price, product.Price)
	assert.Equal(t, expectedProduct.Stock, product.Stock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnError(sql.ErrNoRows)

	product, err := repo.GetByID(productID)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "product not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnError(sql.ErrConnDone)

	product, err := repo.GetByID(productID)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "failed to get product")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product1 := &domain.Product{
		ID:          uuid.New(),
		Name:        "Product 1",
		Description: "First product",
		Price:       50.00,
		Stock:       5,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	product2 := &domain.Product{
		ID:          uuid.New(),
		Name:        "Product 2",
		Description: "Second product",
		Price:       75.00,
		Stock:       3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
		AddRow(product1.ID, product1.Name, product1.Description, product1.Price, product1.Stock, product1.CreatedAt, product1.UpdatedAt).
		AddRow(product2.ID, product2.Name, product2.Description, product2.Price, product2.Stock, product2.CreatedAt, product2.UpdatedAt)

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	products, err := repo.GetAll(10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, products)
	assert.Len(t, products, 2)
	assert.Equal(t, product1.Name, products[0].Name)
	assert.Equal(t, product2.Name, products[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_EmptyResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"})

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	products, err := repo.GetAll(10, 0)

	assert.NoError(t, err)
	assert.Empty(t, products)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnError(sql.ErrConnDone)

	products, err := repo.GetAll(10, 0)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Contains(t, err.Error(), "failed to get products")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_GetAll_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	// Create rows with invalid data type for price (string instead of float)
	rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "created_at", "updated_at"}).
		AddRow(uuid.New(), "Product 1", "Description", "invalid_price", 5, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, name, description, price, stock, created_at, updated_at FROM products ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
		WithArgs(10, 0).
		WillReturnRows(rows)

	products, err := repo.GetAll(10, 0)

	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Contains(t, err.Error(), "failed to scan product")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mock.ExpectExec(`UPDATE products SET name = \$2, description = \$3, price = \$4, stock = \$5, updated_at = \$6 WHERE id = \$1`).
		WithArgs(product.ID, product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(product)

	assert.NoError(t, err)
	assert.False(t, product.UpdatedAt.IsZero())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mock.ExpectExec(`UPDATE products SET name = \$2, description = \$3, price = \$4, stock = \$5, updated_at = \$6 WHERE id = \$1`).
		WithArgs(product.ID, product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mock.ExpectExec(`UPDATE products SET name = \$2, description = \$3, price = \$4, stock = \$5, updated_at = \$6 WHERE id = \$1`).
		WithArgs(product.ID, product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	err = repo.Update(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update product")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(productID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnError(sql.ErrConnDone)

	err = repo.Delete(productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete product")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Update_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mock.ExpectExec(`UPDATE products SET name = \$2, description = \$3, price = \$4, stock = \$5, updated_at = \$6 WHERE id = \$1`).
		WithArgs(product.ID, product.Name, product.Description, product.Price, product.Stock, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err = repo.Update(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get rows affected")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestProductRepository_Delete_RowsAffectedError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewProductRepository(db)

	productID := uuid.New()

	mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
		WithArgs(productID).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err = repo.Delete(productID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get rows affected")
	assert.NoError(t, mock.ExpectationsWereMet())
}
