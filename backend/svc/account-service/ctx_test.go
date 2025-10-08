package main

import (
	"context"
	"testing"

	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

func TestNewAppCtx(t *testing.T) {
	t.Setenv("KAFKA_BROKER", "foo")
	t.Setenv("DB_TYPE", "_TEST_")

	config, err := LoadConfig()
	if err != nil {
		t.Errorf("test setup error: %v", err)
	}

	ctx := newAppCtx(context.Background(), config)

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
	if ctx.payReqReader == nil {
		t.Error("payReqReader should not be nil")
	}
}

func TestPaymentCtxClose(t *testing.T) {
	mockWriter := &tu.MockKafkaWriter{}

	ctx := &accountsCtx{
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
