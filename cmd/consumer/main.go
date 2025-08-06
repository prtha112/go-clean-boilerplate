package main

import (
	"context"
	"log"

	_ "github.com/lib/pq"

	"go-clean-v2/config"
	internalKafka "go-clean-v2/internal/delivery/consumer"
	"go-clean-v2/internal/repository"
	"go-clean-v2/internal/usecase"
	"go-clean-v2/pkg/database"
	pkgKafka "go-clean-v2/pkg/kafka"
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

	// Create a new PostgreSQL connection using pkg/database
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.CloseConnection(db)
	// Create Kafka consumer using pkg/kafka
	kafka := pkgKafka.NewKafkaConsumer(&pkgKafka.ConfigConsumer{
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
	defer kafka.Close()

	// Initialize repositories and use cases
	repo := repository.NewInvoicePostgres(db)
	uc := usecase.NewInvoiceKafkaUsecase(repo)

	// Initialize Kafka consumer
	consumer := internalKafka.NewInvoiceConsumer(kafka, uc)
	defer consumer.Close()

	// Start consuming messages
	consumer.Start(ctx)
}
