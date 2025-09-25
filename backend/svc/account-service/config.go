package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

// Config holds all configuration for the account service
type Config struct {
	Server ServerConfig
	Kafka  KafkaConfig
	DB     DatabaseConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Broker           string
	PaymentTopic     string
	TransactionTopic string
	GroupID          string
	RequiredAcks     kafka.RequiredAcks
	MaxAttempts      int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	// Add database specific config here if needed
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnvOrDefault("SERVE_PORT", "8080"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Kafka: KafkaConfig{
			Broker:           os.Getenv("KAFKA_BROKER"),
			PaymentTopic:     "payment-requested",
			TransactionTopic: "transaction-requested",
			GroupID:          "payment-validator",
			RequiredAcks:     1,
			MaxAttempts:      5,
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	if c.Kafka.Broker == "" {
		return fmt.Errorf("KAFKA_BROKER environment variable is required")
	}

	if c.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(c.Server.Port); err != nil {
		return fmt.Errorf("invalid server port: %s", c.Server.Port)
	}

	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
