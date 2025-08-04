package usecase

import (
	"errors"
	"testing"

	"go-clean-v2/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock ProductRepository
type mockProductRepository struct {
	mock.Mock
}

func (m *mockProductRepository) Create(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductRepository) GetByID(id uuid.UUID) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductRepository) GetAll(limit, offset int) ([]*domain.Product, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *mockProductRepository) Update(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestNewProductUsecase(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	assert.NotNil(t, usecase)
	assert.Implements(t, (*domain.ProductUsecase)(nil), usecase)
}

func TestProductUsecase_Create_Success(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mockRepo.On("Create", product).Return(nil)

	err := usecase.Create(product)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_Create_EmptyName(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	err := usecase.Create(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product name is required")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestProductUsecase_Create_NegativePrice(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "A test product",
		Price:       -10.00,
		Stock:       10,
	}

	err := usecase.Create(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product price cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestProductUsecase_Create_NegativeStock(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       -5,
	}

	err := usecase.Create(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product stock cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestProductUsecase_GetByID_Success(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	expectedProduct := &domain.Product{
		ID:          productID,
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mockRepo.On("GetByID", productID).Return(expectedProduct, nil)

	product, err := usecase.GetByID(productID)

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, expectedProduct, product)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product, err := usecase.GetByID(uuid.Nil)

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "invalid product ID")
	mockRepo.AssertNotCalled(t, "GetByID")
}

func TestProductUsecase_GetByID_NotFound(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	mockRepo.On("GetByID", productID).Return(nil, errors.New("product not found"))

	product, err := usecase.GetByID(productID)

	assert.Error(t, err)
	assert.Nil(t, product)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_GetAll_Success(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	expectedProducts := []*domain.Product{
		{
			ID:    uuid.New(),
			Name:  "Product 1",
			Price: 50.00,
			Stock: 5,
		},
		{
			ID:    uuid.New(),
			Name:  "Product 2",
			Price: 75.00,
			Stock: 3,
		},
	}

	mockRepo.On("GetAll", 10, 0).Return(expectedProducts, nil)

	products, err := usecase.GetAll(10, 0)

	assert.NoError(t, err)
	assert.NotNil(t, products)
	assert.Len(t, products, 2)
	assert.Equal(t, expectedProducts, products)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_GetAll_DefaultLimit(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	expectedProducts := []*domain.Product{}
	mockRepo.On("GetAll", 10, 0).Return(expectedProducts, nil)

	products, err := usecase.GetAll(0, 0) // Should use default limit of 10

	assert.NoError(t, err)
	assert.NotNil(t, products)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_GetAll_MaxLimit(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	expectedProducts := []*domain.Product{}
	mockRepo.On("GetAll", 100, 0).Return(expectedProducts, nil)

	products, err := usecase.GetAll(200, 0) // Should be capped at 100

	assert.NoError(t, err)
	assert.NotNil(t, products)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_GetAll_NegativeOffset(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	expectedProducts := []*domain.Product{}
	mockRepo.On("GetAll", 10, 0).Return(expectedProducts, nil)

	products, err := usecase.GetAll(10, -5) // Should be set to 0

	assert.NoError(t, err)
	assert.NotNil(t, products)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_Update_Success(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:          productID,
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mockRepo.On("GetByID", productID).Return(product, nil)
	mockRepo.On("Update", product).Return(nil)

	err := usecase.Update(product)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_Update_InvalidID(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:    uuid.Nil,
		Name:  "Updated Product",
		Price: 149.99,
		Stock: 20,
	}

	err := usecase.Update(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid product ID")
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestProductUsecase_Update_EmptyName(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	product := &domain.Product{
		ID:    uuid.New(),
		Name:  "",
		Price: 149.99,
		Stock: 20,
	}

	err := usecase.Update(product)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product name is required")
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestProductUsecase_Update_ProductNotFound(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Updated Product",
		Price: 149.99,
		Stock: 20,
	}

	mockRepo.On("GetByID", productID).Return(nil, errors.New("product not found"))

	err := usecase.Update(product)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestProductUsecase_Delete_Success(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	product := &domain.Product{
		ID:    productID,
		Name:  "Product to Delete",
		Price: 99.99,
		Stock: 5,
	}

	mockRepo.On("GetByID", productID).Return(product, nil)
	mockRepo.On("Delete", productID).Return(nil)

	err := usecase.Delete(productID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductUsecase_Delete_InvalidID(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	err := usecase.Delete(uuid.Nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid product ID")
	mockRepo.AssertNotCalled(t, "GetByID")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestProductUsecase_Delete_ProductNotFound(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	productID := uuid.New()
	mockRepo.On("GetByID", productID).Return(nil, errors.New("product not found"))

	err := usecase.Delete(productID)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestProductUsecase_ValidationEdgeCases(t *testing.T) {
	mockRepo := new(mockProductRepository)
	usecase := NewProductUsecase(mockRepo)

	tests := []struct {
		name        string
		product     *domain.Product
		expectError bool
		errorMsg    string
	}{
		{
			name: "zero price allowed",
			product: &domain.Product{
				ID:    uuid.New(),
				Name:  "Free Product",
				Price: 0.00,
				Stock: 10,
			},
			expectError: false,
		},
		{
			name: "zero stock allowed",
			product: &domain.Product{
				ID:    uuid.New(),
				Name:  "Out of Stock Product",
				Price: 99.99,
				Stock: 0,
			},
			expectError: false,
		},
		{
			name: "very small price",
			product: &domain.Product{
				ID:    uuid.New(),
				Name:  "Penny Product",
				Price: 0.01,
				Stock: 1,
			},
			expectError: false,
		},
		{
			name: "large stock number",
			product: &domain.Product{
				ID:    uuid.New(),
				Name:  "High Stock Product",
				Price: 1.00,
				Stock: 999999,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectError {
				mockRepo.On("Create", tt.product).Return(nil).Once()
			}

			err := usecase.Create(tt.product)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}

	mockRepo.AssertExpectations(t)
}
