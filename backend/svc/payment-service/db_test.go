package main

import (
	"errors"
	"testing"
	"time"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func TestCreateDBPayment(t *testing.T) {
	mockDB := &mockDB{}
	appCtx := &paymentCtx{db: mockDB}

	req := cmn.PaymentRequest{
		SourceAccountID: 123,
		TargetAccountID: 789,
		Amount:          10050,
		Timestamp:       time.Now(),
		SystemID:        "test-id",
	}

	err := createDBPayment(req, appCtx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mockDB.createPaymentErr = errors.New("db error")
	err = createDBPayment(req, appCtx)
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestInitDB(t *testing.T) {
	t.Setenv("DB_TYPE", "POSTGRES")
	t.Setenv("POSTGRES_HOST", "localhost")
	
	_, err := initDB()
	if err == nil {
		t.Error("expected connection error in test environment")
	}
}