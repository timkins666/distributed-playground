package main

import (
	"database/sql"
	"log"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type dbPostgres struct {
	db *sql.DB
}

func (db *dbPostgres) commitTransaction(transaction *cmn.Transaction) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// TODO: redis
	var exists bool
	err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM transactions.transaction WHERE id = $1
        )`, transaction.TxID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Transaction already processed:", transaction.TxID)
		return errTxProcessed
	}

	var balance int64
	// FOR UPDATE = pessimistic lock
	err = tx.QueryRow(`
        SELECT balance FROM accounts.account WHERE id = $1 FOR UPDATE
    `, transaction.AccountID).Scan(&balance)
	if err != nil {
		log.Printf("account not found for transaction %+v\n%s", transaction, err)
		return errAccountNotExist
	}

	newBalance := balance + transaction.Amount
	_, err = tx.Exec(`
        UPDATE accounts.account SET balance = $1 WHERE id = $2
		`, newBalance, transaction.AccountID)
	if err != nil {
		log.Println("err update")
		return err
	}

	_, err = tx.Exec(`
        INSERT INTO transactions.transaction (id, account_id, kafka_id, amount) VALUES ($1, $2, $3, $4)
    `, transaction.TxID, transaction.AccountID, transaction.KafkaID, transaction.Amount)
	if err != nil {
		log.Println("err insert into")
		return err
	}

	return tx.Commit()
}

// get single account matching id. always uses db for source of truth.
func (db *dbPostgres) getAccountByID(accountID int32) (*cmn.Account, error) {
	// TODO: squirrel / sqlx

	acc := cmn.Account{}

	err := db.db.QueryRow(`
		SELECT id, user_id, balance from accounts.account WHERE id = $1
	`, accountID).Scan(&acc.AccountID, &acc.UserID, &acc.Balance)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}
