package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		log.Fatalf("KAFKA_BROKER not found")
	} else {
		log.Println("broker env", kafkaBroker)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		GroupID: "process-transaction",
		Topic:   cmn.Topics.TransactionRequested(),
	})
	defer reader.Close()
	completeWriter := &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: cmn.Topics.TransactionComplete(),
	}
	defer completeWriter.Close()
	failedWriter := &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: cmn.Topics.TransactionFailed(),
	}
	defer failedWriter.Close()

	db, err := cmn.InitDB()
	if err != nil {
		log.Panicf(err.Error())
	}

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	go processTransaction(
		db,
		reader,
		completeWriter,
		failedWriter,
		cancelCtx,
	)
}

func processTransaction(
	db *sql.DB,
	reader *kafka.Reader,
	completeWriter *kafka.Writer,
	failedWriter *kafka.Writer,
	cancelCtx context.Context,
) {
	for {
		msg := reader.ReadMessage(cancelCtx)

	}
}

func commitToDB(db *sql.DB, txID, accountID string, amount float64) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Check idempotency
	var exists bool
	err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM transactions.transaction WHERE id = $1
        )`, txID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		log.Println("Transaction already processed:", txID)
		return nil // idempotent — skip
	}

	// 2. Lock the account row FOR UPDATE (pessimistic lock)
	var balance float64
	err = tx.QueryRow(`
        SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
    `, accountID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	// 3. Update balance
	newBalance := balance + amount
	_, err = tx.Exec(`
        UPDATE accounts SET balance = $1 WHERE id = $2
    `, newBalance, accountID)
	if err != nil {
		return err
	}

	// 4. Insert transaction log (idempotency record)
	_, err = tx.Exec(`
        INSERT INTO transactions (id, account_id, amount) VALUES ($1, $2, $3)
    `, txID, accountID, amount)
	if err != nil {
		return err
	}

	// 5. Commit transaction
	return tx.Commit()
}
