package main

import (
	"context"
	"testing"
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
}
