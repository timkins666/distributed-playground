package main

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type dbPostgres struct {
	db *sql.DB
}

// get single account matching id.
func (db *dbPostgres) getAccountByID(accountID int32) (*cmn.Account, error) {

	// TODO: redis
	// TOFO: squirrel / sqlx

	acc := cmn.Account{}

	err := db.db.QueryRow(`
		SELECT id, user_id, balance from accounts.account WHERE id = $1
	`, accountID).Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// get all accounts for the user from db
func (db *dbPostgres) getUserAccounts(userID int32) ([]cmn.Account, error) {

	// TODO: redis
	// TOFO: squirrel / sqlx

	var accounts []cmn.Account

	rows, err := db.db.Query(`
		SELECT id, name, balance FROM accounts.account WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var acc cmn.Account
		err := rows.Scan(&acc.AccountID, &acc.Name, &acc.Balance)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("user %d accounts\n%+v", userID, accounts)

	return accounts, nil
}

func (db *dbPostgres) createAccount(a cmn.Account) (int32, error) {
	var newAccID int32
	err := db.db.QueryRow(`
		INSERT INTO accounts.account (user_id, name)
		VALUES ($1, $2)
		RETURNING id
		`, a.UserID, a.Name).Scan(&newAccID)
	return newAccID, err
}

func (db *dbPostgres) getUserByID(userID int32) (*cmn.User, error) {
	log.Printf("Try load user id %d from db...", userID)
	var user cmn.User
	err := db.db.QueryRow(`
		SELECT id, username, roles FROM accounts."user" WHERE id = $1
	`, userID).Scan(&user.ID, &user.Username, pq.Array(&user.Roles))
	return &user, err
}
