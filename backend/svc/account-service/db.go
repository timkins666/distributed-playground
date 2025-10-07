package main

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type accountsDB interface {
	getUserAccounts(int32) ([]cmn.Account, error)
	createAccount(cmn.Account) (int32, error)
	getAccountByID(int32) (*cmn.Account, error)
	getUserByID(int32) (*cmn.User, error)
}

func initDB(redisClient *redis.Client) (accountsDB, error) {
	dbType, found := os.LookupEnv("DB_TYPE")
	if dbType == "POSTGRES" || !found {
		db, err := cmn.InitPostgres(cmn.DefaultConfig) // TODO: pass config
		if err != nil {
			return nil, fmt.Errorf("failed to initialise database: %w", err)
		}
		return &dbPostgres{db, redisClient}, nil
	}

	panic("cassandra not set up yet")
}
