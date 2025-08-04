package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OrderUsecase
type mockOrderUsecase struct {
	mock.Mock
}

func (m *mockOrderUsecase) Create(req *domain.CreateOrderRequest) (*domain.Order, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderUsecase) GetByID(id uuid.UUID) (*domain.Order, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *mockOrderUsecase) GetAll(limit, offset int) ([]*domain.Order, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Order), args.Error(1)
}

func (m *mockOrderUsecase) Update(order *domain.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *mockOrderUsecase) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockOrderUsecase) UpdateStatus(id uuid.UUID, status domain.OrderStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func setupOrderTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewOrderHandler(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.orderUsecase)
}

func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.POST("/orders", handler.CreateOrder)

	productID := uuid.New()

	req := domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	expectedOrder := &domain.Order{
		ID:            uuid.New(),
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Status:        domain.OrderStatusPending,
		TotalAmount:   100.00,
		CreatedAt:     time.Now(),
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.CreateOrderRequest")).Return(expectedOrder, nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order created successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_InvalidJSON(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.POST("/orders", handler.CreateOrder)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")

	mockUsecase.AssertNotCalled(t, "Create")
}

func TestOrderHandler_CreateOrder_InsufficientStock(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.POST("/orders", handler.CreateOrder)

	productID := uuid.New()

	req := domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  100, // Large quantity to trigger insufficient stock
			},
		},
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.CreateOrderRequest")).Return(nil, errors.New("insufficient stock"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "insufficient stock", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_UsecaseError(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.POST("/orders", handler.CreateOrder)

	productID := uuid.New()

	req := domain.CreateOrderRequest{
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Items: []struct {
			ProductID uuid.UUID `json:"product_id" binding:"required"`
			Quantity  int       `json:"quantity" binding:"required,min=1"`
		}{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.CreateOrderRequest")).Return(nil, errors.New("database error"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database error", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_Success(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders/:id", handler.GetOrder)

	orderID := uuid.New()
	productID := uuid.New()

	expectedOrder := &domain.Order{
		ID:            orderID,
		CustomerName:  "John Doe",
		CustomerEmail: "john@example.com",
		Status:        domain.OrderStatusPending,
		TotalAmount:   100.00,
		CreatedAt:     time.Now(),
		Items: []*domain.OrderItem{
			{
				ID:        uuid.New(),
				OrderID:   orderID,
				ProductID: productID,
				Quantity:  2,
				Price:     50.00,
			},
		},
	}

	mockUsecase.On("GetByID", orderID).Return(expectedOrder, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/orders/%s", orderID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_InvalidID(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders/:id", handler.GetOrder)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/orders/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid order ID", response["error"])

	mockUsecase.AssertNotCalled(t, "GetByID")
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders/:id", handler.GetOrder)

	orderID := uuid.New()
	mockUsecase.On("GetByID", orderID).Return(nil, errors.New("order not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/orders/%s", orderID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_GetOrders_Success(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders", handler.GetOrders)

	expectedOrders := []*domain.Order{
		{
			ID:            uuid.New(),
			CustomerName:  "John Doe",
			CustomerEmail: "john@example.com",
			Status:        domain.OrderStatusPending,
			TotalAmount:   100.00,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			CustomerName:  "Jane Smith",
			CustomerEmail: "jane@example.com",
			Status:        domain.OrderStatusConfirmed,
			TotalAmount:   150.00,
			CreatedAt:     time.Now(),
		},
	}

	mockUsecase.On("GetAll", 10, 0).Return(expectedOrders, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/orders", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Orders retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(0), meta["offset"])
	assert.Equal(t, float64(2), meta["count"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_GetOrders_WithPagination(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders", handler.GetOrders)

	expectedOrders := []*domain.Order{
		{ID: uuid.New(), CustomerName: "John Doe", Status: domain.OrderStatusPending},
	}

	mockUsecase.On("GetAll", 5, 10).Return(expectedOrders, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/orders?limit=5&offset=10", nil)

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

func TestOrderHandler_GetOrders_InvalidLimit(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.GET("/orders", handler.GetOrders)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/orders?limit=invalid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid limit parameter", response["error"])

	mockUsecase.AssertNotCalled(t, "GetAll")
}

func TestOrderHandler_UpdateOrderStatus_Success(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.PUT("/orders/:id/status", handler.UpdateOrderStatus)

	orderID := uuid.New()
	req := UpdateOrderStatusRequest{
		Status: "confirmed",
	}

	mockUsecase.On("UpdateStatus", orderID, domain.OrderStatusConfirmed).Return(nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/orders/%s/status", orderID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order status updated successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_UpdateOrderStatus_InvalidID(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.PUT("/orders/:id/status", handler.UpdateOrderStatus)

	req := UpdateOrderStatusRequest{
		Status: "confirmed",
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/orders/invalid-uuid/status", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid order ID", response["error"])

	mockUsecase.AssertNotCalled(t, "UpdateStatus")
}

func TestOrderHandler_UpdateOrderStatus_NotFound(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.PUT("/orders/:id/status", handler.UpdateOrderStatus)

	orderID := uuid.New()
	req := UpdateOrderStatusRequest{
		Status: "confirmed",
	}

	mockUsecase.On("UpdateStatus", orderID, domain.OrderStatusConfirmed).Return(errors.New("order not found"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/orders/%s/status", orderID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_DeleteOrder_Success(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.DELETE("/orders/:id", handler.DeleteOrder)

	orderID := uuid.New()
	mockUsecase.On("Delete", orderID).Return(nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/orders/%s", orderID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Order deleted successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestOrderHandler_DeleteOrder_InvalidID(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.DELETE("/orders/:id", handler.DeleteOrder)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/orders/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid order ID", response["error"])

	mockUsecase.AssertNotCalled(t, "Delete")
}

func TestOrderHandler_DeleteOrder_NotFound(t *testing.T) {
	mockUsecase := new(mockOrderUsecase)
	handler := NewOrderHandler(mockUsecase)

	router := setupOrderTestRouter()
	router.DELETE("/orders/:id", handler.DeleteOrder)

	orderID := uuid.New()
	mockUsecase.On("Delete", orderID).Return(errors.New("order not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/orders/%s", orderID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "order not found", response["error"])

	mockUsecase.AssertExpectations(t)
}
