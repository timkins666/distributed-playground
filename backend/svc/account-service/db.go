package main

import (
	"fmt"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type accountsDB interface {
	getUserAccounts(int32) ([]cmn.Account, error)
	createAccount(cmn.Account) (int32, error)
	getAccountByID(int32) (*cmn.Account, error)
	getUserByID(int32) (*cmn.User, error)
}

func initDB() (accountsDB, error) {
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
