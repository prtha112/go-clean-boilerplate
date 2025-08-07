package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	SetConnMaxLifetime int // in seconds
	SetMaxOpenConns    int // max open connections
	SetMaxIdleConns    int // max idle connections
}

func NewPostgresConnection(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Connection pool settings (customize as needed)
	db.SetMaxOpenConns(cfg.SetMaxOpenConns)
	db.SetMaxIdleConns(cfg.SetMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.SetConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database (with connection pool)")
	return db, nil
}

func CloseConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
}
