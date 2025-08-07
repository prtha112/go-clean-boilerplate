package config

import (
	"os"
	"strconv"
	"time"

	"go-clean-boilerplate/internal/domain"

	"github.com/joho/godotenv"
)

func Load() (*domain.Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &domain.Config{
		Server: domain.ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: domain.DatabaseConfig{
			Host:               getEnv("DB_HOST", "localhost"),
			Port:               getEnv("DB_PORT", "7775"),
			User:               getEnv("DB_USER", "postgres"),
			Password:           getEnv("DB_PASSWORD", "password"),
			DBName:             getEnv("DB_NAME", "go_clean_db"),
			SSLMode:            getEnv("DB_SSLMODE", "disable"),
			SetConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600),
			SetMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 5),
			SetMaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 1),
		},
		JWT: domain.JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		},
		Kafka: domain.KafkaConfig{
			Brokers:                []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:                  getEnv("KAFKA_TOPIC", "invoices"),
			GroupID:                getEnv("KAFKA_GROUP_ID", "invoice-group"),
			ReadTimeout:            time.Duration(getEnvAsInt("KAFKA_READ_TIMEOUT", 0)) * time.Second,
			MinBytes:               getEnvAsInt("KAFKA_MIN_BYTES", 1),                                             // Default to 1 byte
			MaxBytes:               getEnvAsInt("KAFKA_MAX_BYTES", 1024*1024),                                     // Default to 1MB
			MaxWait:                time.Duration(getEnvAsInt("KAFKA_MAX_WAIT", 1)) * time.Millisecond,            // Default to 1ms
			CommitInterval:         time.Duration(getEnvAsInt("KAFKA_COMMIT_INTERVAL", 100)) * time.Millisecond,   // Default to 100ms
			QueueCapacity:          getEnvAsInt("KAFKA_QUEUE_CAPACITY", 1000),                                     // Default to 1000 messages
			ReadLagInterval:        time.Duration(getEnvAsInt("KAFKA_READ_LAG_INTERVAL", 0)) * time.Second,        // Default to 0 (disabled)
			WatchPartitionChanges:  getEnvAsBool("KAFKA_WATCH_PARTITION_CHANGES", false),                          // Default to false
			PartitionWatchInterval: time.Duration(getEnvAsInt("KAFKA_PARTITION_WATCH_INTERVAL", 0)) * time.Second, // Default to 0 (disabled)
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
