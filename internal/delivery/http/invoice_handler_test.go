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

// Mock InvoiceUsecase
type mockInvoiceUsecase struct {
	mock.Mock
}

func (m *mockInvoiceUsecase) Create(req *domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) CreateFromOrder(orderID uuid.UUID, req *domain.CreateInvoiceRequest) (*domain.Invoice, error) {
	args := m.Called(orderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) GetByID(id uuid.UUID) (*domain.Invoice, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) GetByInvoiceNumber(invoiceNumber string) (*domain.Invoice, error) {
	args := m.Called(invoiceNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) GetAll(limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) GetByStatus(status domain.InvoiceStatus, limit, offset int) ([]*domain.Invoice, error) {
	args := m.Called(status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *mockInvoiceUsecase) UpdateStatus(id uuid.UUID, status domain.InvoiceStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *mockInvoiceUsecase) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockInvoiceUsecase) SendToKafka(invoice *domain.Invoice, eventType string) error {
	args := m.Called(invoice, eventType)
	return args.Error(0)
}

func setupInvoiceTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewInvoiceHandler(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUsecase, handler.invoiceUsecase)
}

func TestInvoiceHandler_CreateInvoice_Success(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.POST("/invoices", handler.CreateInvoice)

	req := domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().Add(30 * 24 * time.Hour),
		TaxRate:        0.1,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Test Item",
				Quantity:    2,
				UnitPrice:   50.00,
			},
		},
	}

	expectedInvoice := &domain.Invoice{
		ID:             uuid.New(),
		InvoiceNumber:  "INV-001",
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		Status:         domain.InvoiceStatusDraft,
		SubTotal:       100.00,
		TaxAmount:      10.00,
		TotalAmount:    110.00,
		CreatedAt:      time.Now(),
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.CreateInvoiceRequest")).Return(expectedInvoice, nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/invoices", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invoice created successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_CreateInvoice_InvalidJSON(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.POST("/invoices", handler.CreateInvoice)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/invoices", bytes.NewBuffer([]byte("invalid json")))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")

	mockUsecase.AssertNotCalled(t, "Create")
}

func TestInvoiceHandler_CreateInvoice_UsecaseError(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.POST("/invoices", handler.CreateInvoice)

	req := domain.CreateInvoiceRequest{
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		DueDate:        time.Now().Add(30 * 24 * time.Hour),
		TaxRate:        0.1,
		Items: []struct {
			ProductID   *uuid.UUID `json:"product_id,omitempty"`
			Description string     `json:"description" binding:"required"`
			Quantity    int        `json:"quantity" binding:"required,min=1"`
			UnitPrice   float64    `json:"unit_price" binding:"required,min=0"`
		}{
			{
				Description: "Test Item",
				Quantity:    2,
				UnitPrice:   50.00,
			},
		},
	}

	mockUsecase.On("Create", mock.AnythingOfType("*domain.CreateInvoiceRequest")).Return(nil, errors.New("database error"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/invoices", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "database error", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_GetInvoice_Success(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices/:id", handler.GetInvoice)

	invoiceID := uuid.New()
	orderID := uuid.New()

	expectedInvoice := &domain.Invoice{
		ID:             invoiceID,
		InvoiceNumber:  "INV-001",
		OrderID:        &orderID,
		CustomerName:   "John Doe",
		CustomerEmail:  "john@example.com",
		BillingAddress: "123 Main St",
		Status:         domain.InvoiceStatusSent,
		SubTotal:       100.00,
		TaxAmount:      10.00,
		TotalAmount:    110.00,
		DueDate:        time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:      time.Now(),
	}

	mockUsecase.On("GetByID", invoiceID).Return(expectedInvoice, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/invoices/%s", invoiceID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invoice retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_GetInvoice_InvalidID(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices/:id", handler.GetInvoice)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/invoices/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid invoice ID", response["error"])

	mockUsecase.AssertNotCalled(t, "GetByID")
}

func TestInvoiceHandler_GetInvoice_NotFound(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices/:id", handler.GetInvoice)

	invoiceID := uuid.New()
	mockUsecase.On("GetByID", invoiceID).Return(nil, errors.New("invoice not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("/invoices/%s", invoiceID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invoice not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_GetInvoices_Success(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices", handler.GetInvoices)

	expectedInvoices := []*domain.Invoice{
		{
			ID:            uuid.New(),
			InvoiceNumber: "INV-001",
			CustomerName:  "John Doe",
			Status:        domain.InvoiceStatusSent,
			TotalAmount:   110.00,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			InvoiceNumber: "INV-002",
			CustomerName:  "Jane Smith",
			Status:        domain.InvoiceStatusPaid,
			TotalAmount:   220.00,
			CreatedAt:     time.Now(),
		},
	}

	mockUsecase.On("GetAll", 10, 0).Return(expectedInvoices, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/invoices", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invoices retrieved successfully", response["message"])
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(0), meta["offset"])
	assert.Equal(t, float64(2), meta["count"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_GetInvoices_WithStatus(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices", handler.GetInvoices)

	expectedInvoices := []*domain.Invoice{
		{ID: uuid.New(), Status: domain.InvoiceStatusPaid},
	}

	mockUsecase.On("GetByStatus", domain.InvoiceStatusPaid, 10, 0).Return(expectedInvoices, nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/invoices?status=paid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, "paid", meta["status"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_GetInvoices_InvalidLimit(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.GET("/invoices", handler.GetInvoices)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/invoices?limit=invalid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid limit parameter", response["error"])

	mockUsecase.AssertNotCalled(t, "GetAll")
}

func TestInvoiceHandler_UpdateInvoiceStatus_Success(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.PUT("/invoices/:id/status", handler.UpdateInvoiceStatus)

	invoiceID := uuid.New()
	req := domain.UpdateInvoiceStatusRequest{
		Status: domain.InvoiceStatusPaid,
	}

	mockUsecase.On("UpdateStatus", invoiceID, domain.InvoiceStatusPaid).Return(nil)

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/invoices/%s/status", invoiceID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invoice status updated successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_UpdateInvoiceStatus_InvalidID(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.PUT("/invoices/:id/status", handler.UpdateInvoiceStatus)

	req := domain.UpdateInvoiceStatusRequest{
		Status: domain.InvoiceStatusPaid,
	}

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", "/invoices/invalid-uuid/status", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid invoice ID", response["error"])

	mockUsecase.AssertNotCalled(t, "UpdateStatus")
}

func TestInvoiceHandler_UpdateInvoiceStatus_NotFound(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.PUT("/invoices/:id/status", handler.UpdateInvoiceStatus)

	invoiceID := uuid.New()
	req := domain.UpdateInvoiceStatusRequest{
		Status: domain.InvoiceStatusPaid,
	}

	mockUsecase.On("UpdateStatus", invoiceID, domain.InvoiceStatusPaid).Return(errors.New("invoice not found"))

	reqBody, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("PUT", fmt.Sprintf("/invoices/%s/status", invoiceID.String()), bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invoice not found", response["error"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_DeleteInvoice_Success(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.DELETE("/invoices/:id", handler.DeleteInvoice)

	invoiceID := uuid.New()
	mockUsecase.On("Delete", invoiceID).Return(nil)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/invoices/%s", invoiceID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invoice deleted successfully", response["message"])

	mockUsecase.AssertExpectations(t)
}

func TestInvoiceHandler_DeleteInvoice_InvalidID(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.DELETE("/invoices/:id", handler.DeleteInvoice)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", "/invoices/invalid-uuid", nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid invoice ID", response["error"])

	mockUsecase.AssertNotCalled(t, "Delete")
}

func TestInvoiceHandler_DeleteInvoice_NotFound(t *testing.T) {
	mockUsecase := new(mockInvoiceUsecase)
	handler := NewInvoiceHandler(mockUsecase)

	router := setupInvoiceTestRouter()
	router.DELETE("/invoices/:id", handler.DeleteInvoice)

	invoiceID := uuid.New()
	mockUsecase.On("Delete", invoiceID).Return(errors.New("invoice not found"))

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/invoices/%s", invoiceID.String()), nil)

	router.ServeHTTP(w, httpReq)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invoice not found", response["error"])

	mockUsecase.AssertExpectations(t)
}
