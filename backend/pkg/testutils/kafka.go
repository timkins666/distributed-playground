package testutils

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
)

// MockKafkaWriter is a mock implementation of the Kafka writer for testing
type MockKafkaWriter struct {
	Messages []kafka.Message
	WriteErr error
	Closed   bool
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if m.Closed {
		return errors.New("writer has been closed")
	}

	if m.WriteErr != nil {
		return m.WriteErr
	}

	m.Messages = append(m.Messages, msgs...)
	return nil
}

func (m *MockKafkaWriter) Close() error {
	m.Closed = true
	return nil
}

// MockKafkaReader is a mock implementation of the Kafka reader for testing
type MockKafkaReader struct {
	Messages []kafka.Message
	Closed   bool
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	if m.Closed {
		return kafka.Message{}, errors.New("reader has been closed")
	}

	return m.Messages[0], nil
}

func (m *MockKafkaReader) Close() error {
	m.Closed = true
	return nil
}
