package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func TestDBPostgresCommitTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	dbPg := &dbPostgres{db: db}

	tx := &cmn.Transaction{
		TxID:      "test-tx-id",
		AccountID: 123,
		KafkaID:   "test-kafka-id",
		Amount:    1000,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(tx.TxID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT balance FROM accounts.account").
		WithArgs(tx.AccountID).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(5000))
	mock.ExpectExec("UPDATE accounts.account SET balance").
		WithArgs(6000, tx.AccountID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO transactions.transaction").
		WithArgs(tx.TxID, tx.AccountID, tx.KafkaID, tx.Amount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = dbPg.commitTransaction(tx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDBPostgresCommitTransactionAlreadyProcessed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	dbPg := &dbPostgres{db: db}

	tx := &cmn.Transaction{
		TxID:      "test-tx-id",
		AccountID: 123,
	}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(tx.TxID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectRollback()

	err = dbPg.commitTransaction(tx)
	if err != errTxProcessed {
		t.Errorf("expected errTxProcessed, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDBPostgresGetAccountByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	dbPg := &dbPostgres{db: db}

	mock.ExpectQuery("SELECT id, user_id, balance from accounts.account WHERE id").
		WithArgs(123).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "balance"}).
			AddRow(123, 1, 1000))

	account, err := dbPg.getAccountByID(123)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if account.AccountID != 123 {
		t.Errorf("expected account ID 123, got %d", account.AccountID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
