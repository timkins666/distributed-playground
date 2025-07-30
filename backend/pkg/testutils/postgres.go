package testutils

import (
	"database/sql"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

// mock DB type that poorly implements all methods. this won't last long.
type BaseTestDB struct{}

func (BaseTestDB) Expose() *sql.DB {
	return nil
}
func (BaseTestDB) CreateUser(*cmn.User) (int32, error) {
	return -1, cmn.ErrNotImplemented
}
func (BaseTestDB) LoadUserByName(string) (cmn.User, error) {
	return cmn.User{}, cmn.ErrNotImplemented
}
func (BaseTestDB) LoadUserByID(int32) (cmn.User, error) {
	return cmn.User{}, cmn.ErrNotImplemented
}
func (BaseTestDB) GetUserAccounts(int32) ([]cmn.Account, error) {
	return []cmn.Account{}, cmn.ErrNotImplemented
}
func (BaseTestDB) CreateAccount(cmn.Account) (int32, error) {
	return -1, cmn.ErrNotImplemented
}
func (BaseTestDB) GetAccountByID(int32) (*cmn.Account, error) {
	return nil, cmn.ErrNotImplemented
}
func (BaseTestDB) CreatePayment(*cmn.PaymentRequest) error {
	return cmn.ErrNotImplemented
}
func (BaseTestDB) UpdatePaymentStatus(string, string) error {
	return cmn.ErrNotImplemented
}
func (BaseTestDB) CommitTransaction(*cmn.Transaction) error {
	return cmn.ErrNotImplemented
}
