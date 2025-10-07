package main

import (
	"context"
	"testing"

	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

func TestPaymentCtxClose(t *testing.T) {
	mockWriter := &tu.MockKafkaWriter{}

	ctx := &paymentCtx{
		cancelCtx: context.Background(),
		writer:    mockWriter,
	}

	err := ctx.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !mockWriter.Closed {
		t.Error("writer should be closed")
	}
}
