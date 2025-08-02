package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-clean-architecture/internal/domain/invoice"
	invoiceUsecase "go-clean-architecture/internal/usecase/invoice"

	"go-clean-architecture/config"

	"github.com/gorilla/mux"
)

type InvoiceHandler struct {
	uc invoice.UseCase
}

func NewInvoiceHandler(router *mux.Router, repo invoice.Repository) {
	uc := invoiceUsecase.NewInvoiceUseCase(repo)
	handler := &InvoiceHandler{uc: uc}
	router.HandleFunc("/invoices", handler.invoiceHandler).Methods("POST")
}

func (h *InvoiceHandler) invoiceHandler(w http.ResponseWriter, r *http.Request) {
	var inv invoice.Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if inv.ID == "" {
		inv.ID = config.GenerateID()
	}
	if inv.CreatedAt == 0 {
		inv.CreatedAt = time.Now().Unix()
	}

	err := h.uc.ProduceInvoiceMessage(&inv)
	if err != nil {
		log.Printf("produce error: %v", err)
		http.Error(w, "produce error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}
