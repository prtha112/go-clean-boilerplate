package main

import (
	"context"
	"log"
	"os"

	_ "github.com/lib/pq"

	"go-clean-v2/config"
	"go-clean-v2/internal/delivery/kafka"
	"go-clean-v2/internal/repository"
	"go-clean-v2/internal/usecase"
	"go-clean-v2/pkg/database"
)

func main() {
	ctx := context.Background()
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

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.CloseConnection(db)

	repo := repository.NewInvoicePostgres(db)
	uc := usecase.NewInvoiceKafkaUsecase(repo)

	consumer := kafka.NewInvoiceConsumer(
		[]string{os.Getenv("KAFKA_BROKER")},
		"invoice-topic",
		"invoice-group",
		uc,
	)

	consumer.Start(ctx)
}
