package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-clean-boilerplate/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock ProductUsecase
type mockProductUsecase struct {
	mock.Mock
}

func (m *mockProductUsecase) Create(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductUsecase) GetByID(id uuid.UUID) (*domain.Product, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Product), args.Error(1)
}

func (m *mockProductUsecase) GetAll(limit, offset int) ([]*domain.Product, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Product), args.Error(1)
}

func (m *mockProductUsecase) Update(product *domain.Product) error {
	args := m.Called(product)
	return args.Error(0)
}

func (m *mockProductUsecase) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func setupProductTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewProductHandler(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.productUsecase)
}

func TestProductHandler_CreateProduct_Success(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.POST("/products", handler.CreateProduct)

	req := CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.Product")).Return(nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Product created successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_CreateProduct_InvalidJSON(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.POST("/products", handler.CreateProduct)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")

	mockUsecase.AssertNotCalled(t, "Create")
}

func TestProductHandler_CreateProduct_ValidationError(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.POST("/products", handler.CreateProduct)

	req := CreateProductRequest{
		Name:        "", // Empty name should fail validation
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	mockUsecase.AssertNotCalled(t, "Create")
}

func TestProductHandler_CreateProduct_UsecaseError(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.POST("/products", handler.CreateProduct)

	req := CreateProductRequest{
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.Product")).Return(errors.New("database error"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database error", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_GetProduct_Success(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products/:id", handler.GetProduct)

	productID := uuid.New()
	expectedProduct := &domain.Product{
		ID:          productID,
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       10,
	}

	mockUsecase.On("GetByID", productID).Return(expectedProduct, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/products/%s", productID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Product retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_GetProduct_InvalidID(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products/:id", handler.GetProduct)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/products/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid product ID", response["error"])

	mockUsecase.AssertNotCalled(t, "GetByID")
}

func TestProductHandler_GetProduct_NotFound(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products/:id", handler.GetProduct)

	productID := uuid.New()
	mockUsecase.On("GetByID", productID).Return(nil, errors.New("product not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/products/%s", productID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "product not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_GetProducts_Success(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products", handler.GetProducts)

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

	mockUsecase.On("GetAll", 10, 0).Return(expectedProducts, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/products", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Products retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(0), meta["offset"])
	assert.Equal(t, float64(2), meta["count"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_GetProducts_WithPagination(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products", handler.GetProducts)

	expectedProducts := []*domain.Product{
		{ID: uuid.New(), Name: "Product 1"},
	}

	mockUsecase.On("GetAll", 5, 10).Return(expectedProducts, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/products?limit=5&offset=10", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(5), meta["limit"])
	assert.Equal(t, float64(10), meta["offset"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_GetProducts_InvalidLimit(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products", handler.GetProducts)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/products?limit=invalid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid limit parameter", response["error"])

	mockUsecase.AssertNotCalled(t, "GetAll")
}

func TestProductHandler_GetProducts_InvalidOffset(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.GET("/products", handler.GetProducts)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/products?offset=invalid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid offset parameter", response["error"])

	mockUsecase.AssertNotCalled(t, "GetAll")
}

func TestProductHandler_UpdateProduct_Success(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.PUT("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	req := UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mockUsecase.On("Update", mock.AnythingOfType("*domain.Product")).Return(nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/products/%s", productID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Product updated successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_InvalidID(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.PUT("/products/:id", handler.UpdateProduct)

	req := UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/products/invalid-uuid", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid product ID", response["error"])

	mockUsecase.AssertNotCalled(t, "Update")
}

func TestProductHandler_UpdateProduct_NotFound(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.PUT("/products/:id", handler.UpdateProduct)

	productID := uuid.New()
	req := UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       149.99,
		Stock:       20,
	}

	mockUsecase.On("Update", mock.AnythingOfType("*domain.Product")).Return(errors.New("product not found"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/products/%s", productID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "product not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_Success(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.DELETE("/products/:id", handler.DeleteProduct)

	productID := uuid.New()
	mockUsecase.On("Delete", productID).Return(nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/%s", productID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Product deleted successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_InvalidID(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.DELETE("/products/:id", handler.DeleteProduct)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/products/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid product ID", response["error"])

	mockUsecase.AssertNotCalled(t, "Delete")
}

func TestProductHandler_DeleteProduct_NotFound(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.DELETE("/products/:id", handler.DeleteProduct)

	productID := uuid.New()
	mockUsecase.On("Delete", productID).Return(errors.New("product not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/%s", productID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "product not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestProductHandler_DeleteProduct_InternalError(t *testing.T) {
	mockUsecase := new(mockProductUsecase)
	handler := NewProductHandler(mockUsecase)

	router := setupProductTestRouter()
	router.DELETE("/products/:id", handler.DeleteProduct)

	productID := uuid.New()
	mockUsecase.On("Delete", productID).Return(errors.New("database error"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/products/%s", productID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database error", response["error"])

	mockUsecase.AssertExpectations(t)
}
