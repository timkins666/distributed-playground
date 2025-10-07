package main

import (
	"context"
	"testing"

	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

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
