package main

import (
	"fmt"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type transactionDB interface {
	commitTransaction(transaction *cmn.Transaction) error
}

func initDB() (transactionDB, error) {
	dbType, found := os.LookupEnv("DB_TYPE")
	if dbType == "POSTGRES" || !found {
		db, err := cmn.InitPostgres(cmn.DefaultConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		return &dbPostgres{db}, nil
	}

	panic("cassandra not set up yet")
}
