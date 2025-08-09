package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"go-clean-boilerplate/config"
	consumerDelivery "go-clean-boilerplate/internal/delivery/consumer"
	httpDelivery "go-clean-boilerplate/internal/delivery/http"
	"go-clean-boilerplate/internal/domain"
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
	dbConfig := &domain.DatabaseConfig{
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
	defer database.Close(db)

	switch *service {
	case "api":
		kafkaProducer := initKafkaProducer(cfg)
		defer kafkaProducer.Close()
		runAPI(cfg, db, kafkaProducer)
	case "consumer":
		kafkaConsumer := initKafkaConsumer(cfg)
		defer kafkaConsumer.Close()
		runConsumer(cfg, db, kafkaConsumer)
	}
}

func runAPI(cfg *domain.Config, db *sql.DB, kafkaProducer kafka.KafkaProducer) {
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

func runConsumer(cfg *domain.Config, db *sql.DB, kafka kafka.KafkaConsumer) {
	ctx := context.Background()

	// Initialize repositories and use cases
	repo := repository.NewInvoicePostgres(db)
	uc := usecase.NewInvoiceKafkaUsecase(repo)

	// Initialize consumer router (mirrors HTTP router pattern) and register workers
	router := consumerDelivery.NewRouter()
	defer router.Close()

	// Register invoice topic worker
	router.Register(kafka, consumerDelivery.NewInvoiceHandler(uc), "invoices")

	log.Print("Brokers:", cfg.Kafka.Brokers)
	log.Println("Topic:", cfg.Kafka.Topic)
	log.Println("Group ID:", cfg.Kafka.GroupID)
	// Start consuming messages
	router.Start(ctx)
}

func initKafkaProducer(cfg *domain.Config) kafka.KafkaProducer {
	kafkaProducer := kafka.NewKafkaProducer(&domain.KafkaConfig{
		Brokers: cfg.Kafka.Brokers,
		Topic:   cfg.Kafka.Topic,
	})
	return kafkaProducer
}

func initKafkaConsumer(cfg *domain.Config) kafka.KafkaConsumer {
	kafkaConsumer := kafka.NewKafkaConsumer(&domain.KafkaConfig{
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
	return kafkaConsumer
}
