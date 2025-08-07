package domain

import (
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type JWTConfig struct {
	Secret string
}

type DatabaseConfig struct {
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	SetConnMaxLifetime int
	SetMaxOpenConns    int
	SetMaxIdleConns    int
}

type KafkaConfig struct {
	Brokers                []string
	Topic                  string
	GroupID                string
	ReadTimeout            time.Duration
	MinBytes               int
	MaxBytes               int
	MaxWait                time.Duration
	CommitInterval         time.Duration
	QueueCapacity          int
	ReadLagInterval        time.Duration
	WatchPartitionChanges  bool
	PartitionWatchInterval time.Duration
}
