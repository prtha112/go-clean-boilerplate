package http

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

type Invoice struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	CreatedAt int64   `json:"created_at"`
}

func NewInvoiceHandler(router *mux.Router) {
	router.HandleFunc("/invoices", invoiceHandler).Methods("POST")
}

func invoiceHandler(w http.ResponseWriter, r *http.Request) {
	var inv Invoice
	if err := json.NewDecoder(r.Body).Decode(&inv); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if inv.ID == "" {
		inv.ID = generateID()
	}
	if inv.CreatedAt == 0 {
		inv.CreatedAt = time.Now().Unix()
	}

	msg, err := json.Marshal(inv)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	topic := os.Getenv("KAFKA_INVOICE_TOPIC")
	if topic == "" {
		topic = "invoice-topic"
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	err = writer.WriteMessages(r.Context(), kafka.Message{
		Key:   []byte(inv.ID),
		Value: msg,
	})
	if err != nil {
		log.Printf("produce error: %v", err)
		http.Error(w, "produce error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}

func generateID() string {
	return "inv-" + time.Now().Format("20060102150405")
}
