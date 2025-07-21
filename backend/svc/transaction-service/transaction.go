package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	// TODO import "github.com/Masterminds/squirrel"
)

func main() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID: "process-transaction",
		Topic:   cmn.Topics.TransactionRequested(),
	})
	defer reader.Close()
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cmn.KafkaBroker()),
		RequiredAcks: 1,
	}
	defer writer.Close()

	db, err := cmn.InitDB()
	if err != nil {
		log.Panicf(err.Error())
	}

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	go processTransaction(
		db,
		reader,
		writer,
		cancelCtx,
	)

	for {
		select {
		case <-cancelCtx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
		}

	}
}

func processTransaction(
	db *sql.DB,
	reader *kafka.Reader,
	writer *kafka.Writer,
	cancelCtx context.Context,
) {
	// commit transaction to db

	// TODO: failed messages

	for {
		msg, err := reader.ReadMessage(cancelCtx)
		if err != nil {
			log.Println(err)
			continue
		}

		var tx *cmn.Transaction
		err = json.Unmarshal(msg.Value, tx)
		if err != nil {
			log.Println("Error reading transaction", err)
		}

		err = commitToDB(db, tx)
		if err != nil {
			log.Println("Error committing transaction, this is probably bad", err)
		}
	}
}

func commitToDB(db *sql.DB, transaction *cmn.Transaction) error {
	tx, err := db.Begin()
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
		return nil
	}

	// FOR UPDATE = pessimistic lock
	var balance int64
	err = tx.QueryRow(`
        SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
    `, transaction.AccountID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// 3. Update balance
	newBalance := balance + transaction.Amount
	_, err = tx.Exec(`
        UPDATE accounts SET balance = $1 WHERE id = $2
    `, newBalance, transaction.AccountID)
	if err != nil {
		return err
	}

	// 4. Insert transaction log (idempotency record)
	_, err = tx.Exec(`
        INSERT INTO transactions (id, account_id, amount) VALUES ($1, $2, $3)
    `, transaction.TxID, transaction.AccountID, transaction.Amount)
	if err != nil {
		return err
	}

	// 5. Commit transaction
	return tx.Commit()
}
