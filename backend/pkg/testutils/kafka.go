package testutils

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// MockKafkaWriter is a mock implementation of the Kafka writer for testing
type MockKafkaWriter struct {
	Messages []kafka.Message
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.Messages = append(m.Messages, msgs...)
	return nil
}

func (m *MockKafkaWriter) Close() error {
	return nil
}

// MockKafkaReader is a mock implementation of the Kafka reader for testing
type MockKafkaReader struct {
	Messages []kafka.Message
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return m.Messages[0], nil
}

func (m *MockKafkaReader) Close() error {
	return nil
}
