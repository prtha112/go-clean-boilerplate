package main

import (
	"log"

	"go-clean-boilerplate/config"
	httpDelivery "go-clean-boilerplate/internal/delivery/http"
	"go-clean-boilerplate/internal/repository"
	"go-clean-boilerplate/internal/usecase"
	"go-clean-boilerplate/pkg/database"
	"go-clean-boilerplate/pkg/kafka"
)

func main() {
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
