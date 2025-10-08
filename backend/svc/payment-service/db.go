package main

import (
	"fmt"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type transactionDB interface {
	createPayment(*cmn.PaymentRequest) error
}

func initDB() (transactionDB, error) {
	dbType, found := os.LookupEnv("DB_TYPE")

	if dbType == "_TEST_" {
		return &dbPostgres{}, nil
	}

	if dbType == "POSTGRES" || !found {
		db, err := cmn.InitPostgres(cmn.DefaultConfig) // TODO: pass config
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		return &dbPostgres{db}, nil
	}

	panic("cassandra not set up yet")
}
