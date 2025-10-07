package main

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func TestDBPostgresCreatePayment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	dbPg := &dbPostgres{db: db}

	req := &cmn.PaymentRequest{
		SystemID:        "test-id",
		AppID:           "app-id",
		SourceAccountID: 123,
		TargetAccountID: 456,
		Amount:          10050,
	}

	mock.ExpectExec("INSERT INTO payments.transfer").
		WithArgs(req.SystemID, req.AppID, req.SourceAccountID, req.TargetAccountID, req.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = dbPg.createPayment(req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDBPostgresCreatePaymentError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	dbPg := &dbPostgres{db: db}

	req := &cmn.PaymentRequest{
		SystemID:        "test-id",
		AppID:           "app-id", 
		SourceAccountID: 123,
		TargetAccountID: 456,
		Amount:          10050,
	}

	mock.ExpectExec("INSERT INTO payments.transfer").
		WithArgs(req.SystemID, req.AppID, req.SourceAccountID, req.TargetAccountID, req.Amount).
		WillReturnError(sql.ErrConnDone)

	err = dbPg.createPayment(req)
	if err == nil {
		t.Error("expected error but got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}