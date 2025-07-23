package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	appdb "github.com/timkins666/distributed-playground/backend/pkg/appdb"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	// TODO import "github.com/Masterminds/squirrel"
)

type app struct {
	cancelCtx context.Context
	db        *appdb.DB
	txReader  *kafka.Reader
	writer    *kafka.Writer
	log       *log.Logger
}

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

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

	db, err := appdb.InitDB(appdb.DefaultConfig)
	if err != nil {
		log.Panicln(err.Error())
	}

	app := app{
		cancelCtx: cancelCtx,
		db:        db,
		txReader:  reader,
		writer:    writer,
		log:       cmn.AppLogger(),
	}

	go processTransaction(app)

	for {
		select {
		case <-cancelCtx.Done():
			return
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

func processTransaction(app app) {
	// commit transaction to db

	// TODO: failed messages

	for {
		msg, err := app.txReader.ReadMessage(app.cancelCtx)
		if err != nil {
			log.Println(err)
			continue
		}

		var tx *cmn.Transaction
		err = json.Unmarshal(msg.Value, tx)
		if err != nil {
			log.Println("Error reading transaction", err)
			return
		}

		err = commitToDB(tx, app)
		if err != nil {
			log.Println("Error committing transaction, this is probably bad", err)
			return
		}

		// TODO: complete message
		log.Printf("Completed transaction %+v", tx)
	}
}

func commitToDB(transaction *cmn.Transaction, app app) error {
	tx, err := app.db.Expose().Begin()
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

	var balance int64
	// FOR UPDATE = pessimistic lock
	err = tx.QueryRow(`
        SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
    `, transaction.AccountID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	newBalance := balance + transaction.Amount
	_, err = tx.Exec(`
        UPDATE accounts SET balance = $1 WHERE id = $2
    `, newBalance, transaction.AccountID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        INSERT INTO transactions (id, account_id, amount) VALUES ($1, $2, $3)
    `, transaction.TxID, transaction.AccountID, transaction.Amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
