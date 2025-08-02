package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

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

// main is the application entrypoint. Selects mode by argument.
func main() {
	_ = godotenv.Load()
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	switch os.Args[1] {
	case "restapi":
		runRESTServer()
	case "consume-invoice":
		runInvoiceKafkaConsumer()
	default:
		printUsageAndExit()
	}
}

// runRESTServer starts the HTTP REST API server.
func runRESTServer() {
	cfg := mustLoadConfig()
	db := mustSetupDatabase(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	router := setupRouter(db)
	printRoutes(router)
	log.Printf("REST API listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// printRoutes logs all registered REST API paths and methods.
func printRoutes(router *mux.Router) {
	log.Println("Available REST API endpoints:")
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("  %s %s", methods, path)
		return nil
	})
	if err != nil {
		log.Printf("error walking routes: %v", err)
	}
}

// runInvoiceKafkaConsumer starts the Kafka consumer for invoice events.
func runInvoiceKafkaConsumer() {
	cfg := mustLoadConfig()
	db := mustSetupDatabase(cfg)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing db: %v", err)
		}
	}()
	brokers := []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	topic := getEnv("KAFKA_INVOICE_TOPIC", "invoice-topic")
	groupID := getEnv("KAFKA_INVOICE_GROUP", "invoice-group")
	repo := invoiceInfra.NewPostgresInvoiceRepository(db)
	uc := invoiceUsecase.NewInvoiceUseCase(repo)
	consumer := invoiceInfra.NewKafkaInvoiceConsumer(brokers, topic, groupID, uc.ConsumeInvoiceMessage)
	ctx := context.Background()
	log.Printf("Kafka invoice consumer started (topic: %s, group: %s)", topic, groupID)
	consumer.Start(ctx)
}

// printUsageAndExit prints usage and exits with error code.
func printUsageAndExit() {
	log.Println("Usage: ./app [restapi|consume-invoice]")
	os.Exit(1)
}

// config holds application configuration.
type config struct {
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDB       string
	PGSSL      string
	Port       string
}

// mustLoadConfig loads config and validates required fields.
func mustLoadConfig() *config {
	cfg := &config{
		PGHost:     getEnv("PG_HOST", "localhost"),
		PGPort:     getEnv("PG_PORT", "15432"),
		PGUser:     getEnv("PG_USER", "mock"),
		PGPassword: getEnv("PG_PASSWORD", "mock123"),
		PGDB:       getEnv("PG_DB", "mockdb"),
		PGSSL:      getEnv("PG_SSLMODE", "disable"),
		Port:       getEnv("PORT", "8085"),
	}
	// Add more validation as needed
	if cfg.PGHost == "" || cfg.PGUser == "" || cfg.PGPassword == "" || cfg.PGDB == "" {
		log.Fatal("database config is required (PG_HOST, PG_USER, PG_PASSWORD, PG_DB)")
	}
	return cfg
}

// mustSetupDatabase opens a postgres connection or exits on error.
func mustSetupDatabase(cfg *config) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDB, cfg.PGSSL,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	// Optionally ping to check connection
	if err := db.Ping(); err != nil {
		log.Fatalf("cannot ping postgres: %v", err)
	}
	return db
}

// setupRouter wires up all HTTP handlers and returns the router.
func setupRouter(db *sql.DB) *mux.Router {
	router := httpHandler.NewRouter()

	// Public endpoints
	httpHandler.NewAuthHandler(router, db) // /login (no JWT required)

	// Protected endpoints (JWT)
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(httpHandler.JWTMiddleware)

	uRepo := userRepo.NewUserRepository()
	uUC := userUsecase.NewUserUseCase(uRepo)
	httpHandler.NewUserHandler(protected, uUC)

	oRepo := orderRepo.NewOrderRepository(db)
	oUC := orderUsecase.NewOrderUseCase(oRepo)
	httpHandler.NewOrderHandler(protected, oUC)

	httpHandler.NewInvoiceHandler(protected)

	return router
}

// getEnv returns the value of the environment variable or fallback if not set.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
