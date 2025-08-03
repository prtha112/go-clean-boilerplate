package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"

	config "go-clean-architecture/config"
	invoiceRepo "go-clean-architecture/internal/infrastructure/invoice"
	invoiceUsecase "go-clean-architecture/internal/usecase/invoice"

	"github.com/gorilla/mux"

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
	ctx := context.Background()

	cfg := config.MustLoadConfig()
	cfgKafka := config.MustLoadConfigKafkaInvoice()
	cfgOtel := config.MustLoadOtelConfig()
	db := config.MustSetupDatabase(cfg)
	kafka := config.MustSetupKafkaProducer(cfgKafka)

	otelTracer := config.MustSetupOtelTracer(cfgOtel, ctx)
	defer otelTracer(ctx)

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()

	router := setupRouter(db, kafka)
	config.PrintRoutes(router)
	log.Printf("REST API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// runInvoiceKafkaConsumer starts the Kafka consumer for invoice events.
func runInvoiceKafkaConsumer() {
	cfg := config.MustLoadConfig()
	cfgKafka := config.MustLoadConfigKafkaInvoice()
	db := config.MustSetupDatabase(cfg)
	kafkaProducer := config.MustSetupKafkaProducer(cfgKafka)
	kafkaConsumer := config.MustSetupKafkaConsumer(cfgKafka)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	invoiceRepoInstance := invoiceRepo.NewPostgresInvoiceRepository(db, kafkaProducer)
	invoiceUsecaseInstance := invoiceUsecase.NewInvoiceUseCase(invoiceRepoInstance)
	consumer := invoiceRepo.NewKafkaInvoiceConsumer(kafkaConsumer, invoiceUsecaseInstance.ConsumeInvoiceMessage)
	ctx := context.Background()
	log.Printf("Kafka invoice consumer started")
	consumer.Start(ctx)
}

// setupRouter wires up all HTTP handlers and returns the router.
func setupRouter(db *sql.DB, kafka *kafka.Writer) *mux.Router {
	router := httpHandler.NewRouter()

	// Public endpoints
	httpHandler.NewAuthHandler(router, db) // /login (no JWT required)

	// Protected endpoints (JWT)
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(httpHandler.JWTMiddleware)

	orderRepo := orderRepo.NewOrderRepository(db)
	orderUsecase := orderUsecase.NewOrderUseCase(orderRepo)
	httpHandler.NewOrderHandler(protected, orderUsecase)

	invoiceRepo := invoiceRepo.NewPostgresInvoiceRepository(db, kafka)
	httpHandler.NewInvoiceHandler(protected, invoiceRepo)

	return router
}
