package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("KAFKA_BROKER", "localhost:9092")
	t.Setenv("SERVE_PORT", "8080")

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Kafka.Broker != "localhost:9092" {
		t.Errorf("expected broker localhost:9092, got %s", config.Kafka.Broker)
	}

	if config.Server.Port != "8080" {
		t.Errorf("expected port 8080, got %s", config.Server.Port)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{Port: "8080"},
				Kafka:  KafkaConfig{Broker: "localhost:9092"},
			},
			expectErr: false,
		},
		{
			name: "missing kafka broker",
			config: Config{
				Server: ServerConfig{Port: "8080"},
				Kafka:  KafkaConfig{},
			},
			expectErr: true,
		},
		{
			name: "invalid port",
			config: Config{
				Server: ServerConfig{Port: "invalid"},
				Kafka:  KafkaConfig{Broker: "localhost:9092"},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")

	result := getEnvOrDefault("TEST_VAR", "default")
	if result != "test_value" {
		t.Errorf("expected test_value, got %s", result)
	}

	result = getEnvOrDefault("NONEXISTENT", "default")
	if result != "default" {
		t.Errorf("expected default, got %s", result)
	}
}
