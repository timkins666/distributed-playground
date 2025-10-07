package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func TestDBPostgresGetUserAccounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{})
	dbPg := &dbPostgres{db: db, redisClient: redisClient}

	rows := sqlmock.NewRows([]string{"id", "name", "balance"}).
		AddRow(1, "Test Account", 1000).
		AddRow(2, "Another Account", 2000)

	mock.ExpectQuery("SELECT id, name, balance FROM accounts.account WHERE user_id").
		WithArgs(1).
		WillReturnRows(rows)

	accounts, err := dbPg.getUserAccounts(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDBPostgresCreateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	redisClient := redis.NewClient(&redis.Options{})
	dbPg := &dbPostgres{db: db, redisClient: redisClient}

	account := cmn.Account{
		UserID: 1,
		Name:   "Test Account",
	}

	mock.ExpectQuery("INSERT INTO accounts.account").
		WithArgs(account.UserID, account.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(123))

	accountID, err := dbPg.createAccount(account)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if accountID != 123 {
		t.Errorf("expected account ID 123, got %d", accountID)
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

	redisClient := redis.NewClient(&redis.Options{})
	dbPg := &dbPostgres{db: db, redisClient: redisClient}

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
