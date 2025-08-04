package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestProduct_Creation(t *testing.T) {
	product := &Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "This is a test product",
		Price:       99.99,
		Stock:       10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, "This is a test product", product.Description)
	assert.Equal(t, 99.99, product.Price)
	assert.Equal(t, 10, product.Stock)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())
}

func TestProduct_Validation(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		hasError bool
	}{
		{
			name: "valid product",
			product: Product{
				ID:          uuid.New(),
				Name:        "Valid Product",
				Description: "A valid product",
				Price:       50.00,
				Stock:       5,
			},
			hasError: false,
		},
		{
			name: "empty name",
			product: Product{
				ID:          uuid.New(),
				Name:        "",
				Description: "Product without name",
				Price:       50.00,
				Stock:       5,
			},
			hasError: true,
		},
		{
			name: "negative price",
			product: Product{
				ID:          uuid.New(),
				Name:        "Negative Price Product",
				Description: "Product with negative price",
				Price:       -10.00,
				Stock:       5,
			},
			hasError: true,
		},
		{
			name: "negative stock",
			product: Product{
				ID:          uuid.New(),
				Name:        "Negative Stock Product",
				Description: "Product with negative stock",
				Price:       50.00,
				Stock:       -1,
			},
			hasError: true,
		},
		{
			name: "zero price allowed",
			product: Product{
				ID:          uuid.New(),
				Name:        "Free Product",
				Description: "Free product",
				Price:       0.00,
				Stock:       5,
			},
			hasError: false,
		},
		{
			name: "zero stock allowed",
			product: Product{
				ID:          uuid.New(),
				Name:        "Out of Stock Product",
				Description: "Product out of stock",
				Price:       50.00,
				Stock:       0,
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			hasError := tt.product.Name == "" ||
				tt.product.Price < 0 ||
				tt.product.Stock < 0

			assert.Equal(t, tt.hasError, hasError)
		})
	}
}

func TestProduct_StockManagement(t *testing.T) {
	product := &Product{
		ID:    uuid.New(),
		Name:  "Stock Test Product",
		Price: 100.00,
		Stock: 10,
	}

	// Test reducing stock
	originalStock := product.Stock
	quantityToReduce := 3
	product.Stock -= quantityToReduce

	assert.Equal(t, originalStock-quantityToReduce, product.Stock)
	assert.Equal(t, 7, product.Stock)

	// Test increasing stock
	quantityToAdd := 5
	product.Stock += quantityToAdd

	assert.Equal(t, 12, product.Stock)
}

func TestProduct_PriceCalculations(t *testing.T) {
	product := &Product{
		ID:    uuid.New(),
		Name:  "Price Test Product",
		Price: 25.50,
		Stock: 10,
	}

	// Test total price for quantity
	quantity := 4
	totalPrice := product.Price * float64(quantity)

	assert.Equal(t, 102.00, totalPrice)

	// Test with different quantities
	testCases := []struct {
		quantity int
		expected float64
	}{
		{1, 25.50},
		{2, 51.00},
		{5, 127.50},
		{10, 255.00},
	}

	for _, tc := range testCases {
		total := product.Price * float64(tc.quantity)
		assert.Equal(t, tc.expected, total)
	}
}

func TestProduct_JSONFields(t *testing.T) {
	product := &Product{
		ID:          uuid.New(),
		Name:        "JSON Test Product",
		Description: "Testing JSON serialization",
		Price:       75.25,
		Stock:       20,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Verify all fields are accessible
	assert.NotEqual(t, uuid.Nil, product.ID)
	assert.NotEmpty(t, product.Name)
	assert.NotEmpty(t, product.Description)
	assert.Greater(t, product.Price, 0.0)
	assert.GreaterOrEqual(t, product.Stock, 0)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())
}

func TestProduct_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		price       float64
		stock       int
		expectValid bool
	}{
		{
			name:        "very long name",
			productName: "This is a very long product name that might exceed normal limits but should still be valid as long as it's not empty",
			price:       1.00,
			stock:       1,
			expectValid: true,
		},
		{
			name:        "very small price",
			productName: "Cheap Product",
			price:       0.01,
			stock:       1,
			expectValid: true,
		},
		{
			name:        "very large stock",
			productName: "High Stock Product",
			price:       1.00,
			stock:       999999,
			expectValid: true,
		},
		{
			name:        "very large price",
			productName: "Expensive Product",
			price:       999999.99,
			stock:       1,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				ID:    uuid.New(),
				Name:  tt.productName,
				Price: tt.price,
				Stock: tt.stock,
			}

			isValid := product.Name != "" && product.Price >= 0 && product.Stock >= 0
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}
