package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	invoiceInfra "go-clean-architecture/internal/infrastructure/invoice"
	invoiceUsecase "go-clean-architecture/internal/usecase/invoice"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	orderRepo "go-clean-architecture/internal/infrastructure/order"
	userRepo "go-clean-architecture/internal/infrastructure/user"
	httpHandler "go-clean-architecture/internal/interface/http"
	orderUsecase "go-clean-architecture/internal/usecase/order"
	userUsecase "go-clean-architecture/internal/usecase/user"
)

func main() {
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	switch os.Args[1] {
	case "restapi":
		runREST()
	case "consume-invoice":
		runInvoiceConsumer()
	default:
		printUsageAndExit()
	}
}

func runREST() {
	cfg := loadConfig()
	db := setupDatabase(cfg)
	defer db.Close()
	router := setupRouter(db)
	log.Printf("Listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, router))
}

func runInvoiceConsumer() {
	cfg := loadConfig()
	db := setupDatabase(cfg)
	defer db.Close()
	brokers := []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	topic := getEnv("KAFKA_INVOICE_TOPIC", "invoice-topic")
	groupID := getEnv("KAFKA_INVOICE_GROUP", "invoice-group")
	repo := invoiceInfra.NewPostgresInvoiceRepository(db)
	uc := invoiceUsecase.NewInvoiceUseCase(repo)
	consumer := invoiceInfra.NewKafkaInvoiceConsumer(brokers, topic, groupID, uc.ConsumeInvoiceMessage)
	ctx := context.Background()
	log.Println("Starting Kafka invoice consumer...")
	consumer.Start(ctx)
}

func printUsageAndExit() {
	log.Println("Usage: ./app [rest|consume-invoice]")
	os.Exit(1)
}

type config struct {
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDB       string
	PGSSL      string
	Port       string
}

func loadConfig() *config {
	return &config{
		PGHost:     getEnv("PG_HOST", "localhost"),
		PGPort:     getEnv("PG_PORT", "15432"),
		PGUser:     getEnv("PG_USER", "mock"),
		PGPassword: getEnv("PG_PASSWORD", "mock123"),
		PGDB:       getEnv("PG_DB", "mockdb"),
		PGSSL:      getEnv("PG_SSLMODE", "disable"),
		Port:       getEnv("PORT", "8085"),
	}
}

func setupDatabase(cfg *config) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDB, cfg.PGSSL,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	return db
}

func setupRouter(db *sql.DB) *mux.Router {
	router := httpHandler.NewRouter()

	// User
	uRepo := userRepo.NewUserRepository()
	uUC := userUsecase.NewUserUseCase(uRepo)
	httpHandler.NewUserHandler(router, uUC)

	// Order
	oRepo := orderRepo.NewPostgresOrderRepository(db)
	oUC := orderUsecase.NewOrderUseCase(oRepo)
	httpHandler.NewOrderHandler(router, oUC)

	return router
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
