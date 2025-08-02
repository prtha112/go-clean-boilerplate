package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	config "go-clean-architecture/internal/config"
	invoiceInfra "go-clean-architecture/internal/infrastructure/invoice"
	invoiceUsecase "go-clean-architecture/internal/usecase/invoice"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	orderRepo "go-clean-architecture/internal/infrastructure/order"
	httpHandler "go-clean-architecture/internal/interface/http"
	orderUsecase "go-clean-architecture/internal/usecase/order"
)

// main is the application entrypoint. Selects mode by argument.
func main() {
	_ = godotenv.Load()
	if len(os.Args) < 2 {
		config.PrintUsageAndExit()
	}

	switch os.Args[1] {
	case "restapi":
		runRESTServer()
	case "consume-invoice":
		runInvoiceKafkaConsumer()
	default:
		config.PrintUsageAndExit()
	}
}

// runRESTServer starts the HTTP REST API server.
func runRESTServer() {
	cfg := config.MustLoadConfig()
	db := config.MustSetupDatabase(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	router := setupRouter(db)
	config.PrintRoutes(router)
	log.Printf("REST API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runInvoiceKafkaConsumer starts the Kafka consumer for invoice events.
func runInvoiceKafkaConsumer() {
	cfg := config.MustLoadConfig()
	db := config.MustSetupDatabase(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	brokers := []string{config.GetEnv("KAFKA_BROKER", "localhost:9092")}
	topic := config.GetEnv("KAFKA_INVOICE_TOPIC", "invoice-topic")
	groupID := config.GetEnv("KAFKA_INVOICE_GROUP", "invoice-group")
	repo := invoiceInfra.NewPostgresInvoiceRepository(db)
	uc := invoiceUsecase.NewInvoiceUseCase(repo)
	consumer := invoiceInfra.NewKafkaInvoiceConsumer(brokers, topic, groupID, uc.ConsumeInvoiceMessage)
	ctx := context.Background()
	log.Printf("Kafka invoice consumer started (topic: %s, group: %s)", topic, groupID)
	consumer.Start(ctx)
}

// setupRouter wires up all HTTP handlers and returns the router.
func setupRouter(db *sql.DB) *mux.Router {
	router := httpHandler.NewRouter()

	// Public endpoints
	httpHandler.NewAuthHandler(router, db) // /login (no JWT required)

	// Protected endpoints (JWT)
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(httpHandler.JWTMiddleware)

	oRepo := orderRepo.NewOrderRepository(db)
	oUC := orderUsecase.NewOrderUseCase(oRepo)
	httpHandler.NewOrderHandler(protected, oUC)

	httpHandler.NewInvoiceHandler(protected)

	return router
}
