package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"golang.org/x/crypto/bcrypt"
)

// config holds application configuration.
type Config struct {
	PGHost     string
	PGPort     string
	PGUser     string
	PGPassword string
	PGDB       string
	PGSSL      string
	Port       string
}

type KafkaConfig struct {
	KAFKA_BROKER        string
	KAFKA_INVOICE_TOPIC string
	KAFKA_INVOICE_GROUP string
	KAFKA_USERNAME      string
	KAFKA_PASSWORD      string
}

type OtelConfig struct {
	OTEL_EXPORTER_OTLP_ENDPOINT string
	OTEL_EXPORTER_OTLP_INSECURE bool
	OTEL_SERVICE_NAME           string
}

// mustLoadConfig loads config and validates required fields.
func MustLoadConfig() *Config {
	cfg := &Config{
		PGHost:     GetEnv("PG_HOST", "localhost"),
		PGPort:     GetEnv("PG_PORT", "15432"),
		PGUser:     GetEnv("PG_USER", "mock"),
		PGPassword: GetEnv("PG_PASSWORD", "mock123"),
		PGDB:       GetEnv("PG_DB", "mockdb"),
		PGSSL:      GetEnv("PG_SSLMODE", "disable"),
		Port:       GetEnv("PORT", "8085"),
	}
	// Add more validation as needed
	if cfg.PGHost == "" || cfg.PGUser == "" || cfg.PGPassword == "" || cfg.PGDB == "" {
		log.Fatal("database config is required (PG_HOST, PG_USER, PG_PASSWORD, PG_DB)")
	}
	return cfg
}

func MustLoadConfigKafkaInvoice() *KafkaConfig {
	cfg := &KafkaConfig{
		KAFKA_BROKER:        GetEnv("KAFKA_BROKER", "localhost:9092"),
		KAFKA_INVOICE_TOPIC: GetEnv("KAFKA_INVOICE_TOPIC", "invoice-topic"),
		KAFKA_INVOICE_GROUP: GetEnv("KAFKA_INVOICE_GROUP", "invoice-group"),
		KAFKA_USERNAME:      GetEnv("KAFKA_USERNAME", ""),
		KAFKA_PASSWORD:      GetEnv("KAFKA_PASSWORD", ""),
	}
	if cfg.KAFKA_BROKER == "" || cfg.KAFKA_INVOICE_TOPIC == "" || cfg.KAFKA_INVOICE_GROUP == "" {
		log.Fatal("Kafka config is required (KAFKA_BROKER, KAFKA_INVOICE_TOPIC, KAFKA_INVOICE_GROUP)")
	}
	return cfg
}

func MustLoadOtelConfig() *OtelConfig {
	cfg := &OtelConfig{
		OTEL_EXPORTER_OTLP_ENDPOINT: GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		OTEL_EXPORTER_OTLP_INSECURE: GetEnv("OTEL_EXPORTER_OTLP_INSECURE", "true") == "true",
		OTEL_SERVICE_NAME:           GetEnv("OTEL_SERVICE_NAME", "go-clean-architecture"),
	}
	if cfg.OTEL_EXPORTER_OTLP_ENDPOINT == "" || cfg.OTEL_SERVICE_NAME == "" {
		log.Fatal("OpenTelemetry config is required (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME)")
	}
	return cfg
}

// mustSetupDatabase opens a postgres connection or exits on error.
func MustSetupDatabase(cfg *Config) *sql.DB {
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

func MustSetupKafkaProducer(cfg *KafkaConfig) *kafka.Writer {
	broker := cfg.KAFKA_BROKER
	topic := cfg.KAFKA_INVOICE_TOPIC
	username := cfg.KAFKA_USERNAME
	password := cfg.KAFKA_PASSWORD
	var dialer *kafka.Dialer
	if username != "" && password != "" {
		dialer = &kafka.Dialer{
			SASLMechanism: plain.Mechanism{
				Username: username,
				Password: password,
			},
		}
	}
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
		Dialer:   dialer,
	})
	return writer
}

func MustSetupKafkaConsumer(cfg *KafkaConfig) *kafka.Reader {
	broker := cfg.KAFKA_BROKER
	topic := cfg.KAFKA_INVOICE_TOPIC
	groupID := cfg.KAFKA_INVOICE_GROUP
	username := cfg.KAFKA_USERNAME
	password := cfg.KAFKA_PASSWORD

	var dialer *kafka.Dialer
	if username != "" && password != "" {
		dialer = &kafka.Dialer{
			SASLMechanism: plain.Mechanism{
				Username: username,
				Password: password,
			},
		}
	}

	readerConfig := kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID,
		Dialer:  dialer,
		MaxWait: 10 * time.Second, // Adjust as needed
	}
	return kafka.NewReader(readerConfig)
}

// getEnv returns the value of the environment variable or fallback if not set.
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// printRoutes logs all registered REST API paths and methods.
func PrintRoutes(router *mux.Router) {
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

// PrintUsageAndExit prints usage and exits with error code.
func PrintUsageAndExit() {
	log.Println("Usage: ./app [restapi|consume-invoice]")
	os.Exit(1)
}

func GenerateHashPassword(password string) string {
	// Placeholder for password hashing logic
	// In production, use a secure hashing algorithm like bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(hashedPassword string, password string) bool {
	// Placeholder for password verification logic
	// In production, use a secure hashing algorithm like bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func GenerateID() string {
	return "inv-" + time.Now().Format("20060102150405")
}

func MustSetupOtelTracer(serviceName string, ctx context.Context) func(context.Context) error {
	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint("localhost:4317"))
	if err != nil {
		log.Fatalf("failed to initialize exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown
}
