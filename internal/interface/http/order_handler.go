package http

import (
    "encoding/json"
    "net/http"
    "strconv"

    domain "go-clean-architecture/internal/domain/order"
    "github.com/gorilla/mux"
)

type OrderHandler struct {
    useCase domain.UseCase
}

func NewOrderHandler(r *mux.Router, uc domain.UseCase) {
    handler := &OrderHandler{useCase: uc}
    r.HandleFunc("/orders/{id}", handler.GetOrder).Methods("GET")
    r.HandleFunc("/orders/{id}", handler.DeleteOrder).Methods("DELETE")
    r.HandleFunc("/orders", handler.GetAllOrders).Methods("GET")
    r.HandleFunc("/orders", handler.CreateOrder).Methods("POST")
}
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }
    if err := h.useCase.DeleteOrder(id); err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
    orders, err := h.useCase.GetAllOrders()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    order, err := h.useCase.GetOrder(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var order domain.Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    if err := h.useCase.CreateOrder(&order); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(order)
}
