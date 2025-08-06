package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"go-clean-boilerplate/config"
	internalKafka "go-clean-boilerplate/internal/delivery/consumer"
	httpDelivery "go-clean-boilerplate/internal/delivery/http"
	"go-clean-boilerplate/internal/repository"
	"go-clean-boilerplate/internal/usecase"
	"go-clean-boilerplate/pkg/database"
	"go-clean-boilerplate/pkg/kafka"
)

func main() {
	// Parse command-line arguments
	service := flag.String("service", "api", "Service to run: 'api' for REST API, 'consumer' for Kafka consumer")
	flag.Parse()

	// Validate service argument
	if *service != "api" && *service != "consumer" {
		fmt.Println("Invalid service. Use 'api' or 'consumer'")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	// Create a new PostgreSQL connection using pkg/database
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.CloseConnection(db)

	switch *service {
	case "api":
		runAPI(cfg, db)
	case "consumer":
		runConsumer(cfg, db)
	}
}

func runAPI(cfg *config.Config, db *sql.DB) {
	// Initialize Kafka producer
	kafkaProducer := kafka.NewKafkaProducer(&kafka.ConfigProducer{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	defer kafkaProducer.Close()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	userRepo := repository.NewUserRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)

	// Initialize use cases
	productUsecase := usecase.NewProductUsecase(productRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, productRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWT.Secret)
	invoiceUsecase := usecase.NewInvoiceUsecase(invoiceRepo, orderRepo, productRepo, kafkaProducer)

	// Initialize HTTP router
	router := httpDelivery.NewRouter(productUsecase, orderUsecase, authUsecase, invoiceUsecase)
	httpRouter := router.SetupRoutes()

	// Start server
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on %s", serverAddr)

	if err := httpRouter.Run(serverAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func runConsumer(cfg *config.Config, db *sql.DB) {
	ctx := context.Background()

	// Create Kafka consumer using pkg/kafka
	kafkaConsumer := kafka.NewKafkaConsumer(&kafka.ConfigConsumer{
		Brokers:                cfg.Kafka.Brokers,
		Topic:                  cfg.Kafka.Topic,
		GroupID:                cfg.Kafka.GroupID,
		ReadTimeout:            cfg.Kafka.ReadTimeout,
		MinBytes:               cfg.Kafka.MinBytes,
		MaxBytes:               cfg.Kafka.MaxBytes,
		MaxWait:                cfg.Kafka.MaxWait,
		CommitInterval:         cfg.Kafka.CommitInterval,
		QueueCapacity:          cfg.Kafka.QueueCapacity,
		ReadLagInterval:        cfg.Kafka.ReadLagInterval,
		WatchPartitionChanges:  cfg.Kafka.WatchPartitionChanges,
		PartitionWatchInterval: cfg.Kafka.PartitionWatchInterval,
	})
	defer kafkaConsumer.Close()

	// Initialize repositories and use cases
	repo := repository.NewInvoicePostgres(db)
	uc := usecase.NewInvoiceKafkaUsecase(repo)

	// Initialize Kafka consumer
	consumer := internalKafka.NewInvoiceConsumer(kafkaConsumer, uc)
	defer consumer.Close()

	// Start consuming messages
	consumer.Start(ctx)
}
