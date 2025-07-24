package testutils

import (
	"database/sql"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

// mock DB type that poorly implements all methods. this won't last long.
type BaseTestDB struct{}

func (BaseTestDB) Expose() *sql.DB                            { return nil }
func (BaseTestDB) CreateUser(*cmn.User) (int, error)          { return -1, nil }
func (BaseTestDB) LoadUserByName(string) (cmn.User, error)    { return cmn.User{}, nil }
func (BaseTestDB) LoadUserByID(int) (cmn.User, error)         { return cmn.User{}, nil }
func (BaseTestDB) GetUserAccounts(int) ([]cmn.Account, error) { return []cmn.Account{}, nil }
func (BaseTestDB) CreateAccount(cmn.Account) (int, error)     { return -1, nil }
func (BaseTestDB) GetAccountByID(int) (*cmn.Account, error)   { return nil, nil }
