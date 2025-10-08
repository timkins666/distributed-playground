package main

import (
	"context"
	"testing"

	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

func TestNewAppCtx(t *testing.T) {
	t.Setenv("KAFKA_BROKER", "foo")
	t.Setenv("DB_TYPE", "_TEST_")

	ctx := newAppCtx(context.Background())
	if ctx.cancelCtx == nil {
		t.Error("cancelCtx should not be nil")
	}
	if ctx.db == nil {
		t.Error("db should not be nil")
	}
	if ctx.logger == nil {
		t.Error("logger should not be nil")
	}
	if ctx.writer == nil {
		t.Error("writer should not be nil")
	}
	if ctx.txReqReader == nil {
		t.Error("txReqReader should not be nil")
	}
}

func TestTransactionCtxClose(t *testing.T) {
	mockReader := &tu.MockKafkaReader{}
	mockWriter := &tu.MockKafkaWriter{}

	ctx := &transactionCtx{
		cancelCtx:   context.Background(),
		txReqReader: mockReader,
		writer:      mockWriter,
		redisClient: nil,
	}

	err := ctx.close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mockReader.Closed {
		t.Error("reader should be closed")
	}
	if !mockWriter.Closed {
		t.Error("writer should be closed")
	}
}
