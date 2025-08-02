package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/mux"
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
