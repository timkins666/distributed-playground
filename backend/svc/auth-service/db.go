package main

import (
	"fmt"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func initDB() (authDB, error) {
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

type authDB interface {
	getUserByName(string) (*cmn.User, error)
	createUser(*cmn.User) (int32, error)
}
