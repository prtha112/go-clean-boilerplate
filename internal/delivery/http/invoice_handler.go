package http

import (
	"net/http"
	"strconv"
	"time"

	"go-clean-v2/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InvoiceHandler struct {
	invoiceUsecase domain.InvoiceUsecase
}

func NewInvoiceHandler(invoiceUsecase domain.InvoiceUsecase) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceUsecase: invoiceUsecase,
	}
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req domain.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.invoiceUsecase.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invoice created successfully",
		"data":    invoice,
	})
}

func (h *InvoiceHandler) CreateInvoiceFromOrder(c *gin.Context) {
	orderIDParam := c.Param("order_id")
	orderID, err := uuid.Parse(orderIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req domain.CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.invoiceUsecase.CreateFromOrder(orderID, &req)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Invoice created from order successfully",
		"data":    invoice,
	})
}

func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	invoice, err := h.invoiceUsecase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice retrieved successfully",
		"data":    invoice,
	})
}

func (h *InvoiceHandler) GetInvoiceByNumber(c *gin.Context) {
	invoiceNumber := c.Param("invoice_number")
	if invoiceNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice number is required"})
		return
	}

	invoice, err := h.invoiceUsecase.GetByInvoiceNumber(invoiceNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice retrieved successfully",
		"data":    invoice,
	})
}

func (h *InvoiceHandler) GetInvoices(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	var invoices []*domain.Invoice

	if status != "" {
		invoiceStatus := domain.InvoiceStatus(status)
		invoices, err = h.invoiceUsecase.GetByStatus(invoiceStatus, limit, offset)
	} else {
		invoices, err = h.invoiceUsecase.GetAll(limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoices retrieved successfully",
		"data":    invoices,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(invoices),
			"status": status,
		},
	})
}

func (h *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	var req domain.UpdateInvoiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.invoiceUsecase.UpdateStatus(id, req.Status); err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice status updated successfully",
	})
}

func (h *InvoiceHandler) DeleteInvoice(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invoice ID"})
		return
	}

	if err := h.invoiceUsecase.Delete(id); err != nil {
		if err.Error() == "invoice not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invoice deleted successfully",
	})
}

func (h *InvoiceHandler) GetOverdueInvoices(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	// Get all unpaid invoices and filter by due date
	invoices, err := h.invoiceUsecase.GetByStatus(domain.InvoiceStatusSent, limit*2, 0) // Get more to filter
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter overdue invoices
	var overdueInvoices []*domain.Invoice
	now := time.Now()
	count := 0

	for _, invoice := range invoices {
		if invoice.DueDate.Before(now) && count < limit {
			if count >= offset {
				overdueInvoices = append(overdueInvoices, invoice)
			}
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Overdue invoices retrieved successfully",
		"data":    overdueInvoices,
		"meta": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(overdueInvoices),
		},
	})
}
