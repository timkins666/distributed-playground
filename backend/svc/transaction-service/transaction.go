package main

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type appEnv struct {
	cmn.BaseEnv
	txReader cmn.KafkaReader
	writer   cmn.KafkaWriter
}

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cmn.KafkaBroker()},
		GroupID: "process-transaction",
		Topic:   cmn.Topics.TransactionRequested().S(),
	})
	defer reader.Close()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cmn.KafkaBroker()),
		RequiredAcks: 1,
	}
	defer writer.Close()

	db, err := cmn.InitDB(cmn.DefaultConfig)
	if err != nil {
		log.Panicln(err.Error())
	}

	env := appEnv{
		BaseEnv:  cmn.BaseEnv{}.WithCancelCtx(cancelCtx).WithDB(db),
		txReader: reader,
		writer:   writer,
	}

	for {
		select {
		case <-cancelCtx.Done():
			return
		default:
			msg, err := env.txReader.ReadMessage(env.CancelCtx())
			if err != nil {
				env.Logger().Println(err)
				continue
			}
			processMessage(msg, env)
		}
	}
}

var (
	errorParsingTransaction    = errors.New("error parsing transaction")
	errorInvalidTransaction    = errors.New("parsed transaction but bad data")
	errorCommittingTransaction = errors.New("error committing transaction, this is probably bad")
)

func processMessage(msg kafka.Message, env appEnv) error {
	tx, err := cmn.FromBytes[cmn.Transaction](msg.Value)

	if err != nil {
		env.Logger().Println(err)
		env.Logger().Printf("received bytes:\n%s", msg.Value)
		return errorParsingTransaction
	}

	if !tx.Valid() {
		env.Logger().Println(errorInvalidTransaction)
		return errorInvalidTransaction
	}

	tx.TxID = uuid.NewString()

	err = env.DB().CommitTransaction(tx)
	if err != nil {
		env.Logger().Println(err)
		return errorCommittingTransaction
	}

	// TODO: complete kafka message
	env.Logger().Printf("Completed transaction %+v", tx)
	return nil
}

// func commitToDB(transaction *cmn.Transaction, env appEnv) error {
// 	tx, err := env.DB().Expose().Begin()
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	// TODO: redis
// 	var exists bool
// 	err = tx.QueryRow(`
//         SELECT EXISTS (
//             SELECT 1 FROM transactions.transaction WHERE id = $1
//         )`, transaction.TxID).Scan(&exists)
// 	if err != nil {
// 		return err
// 	}
// 	if exists {
// 		env.Logger().Println("Transaction already processed:", transaction.TxID)
// 		return nil
// 	}

// 	var balance int64
// 	// FOR UPDATE = pessimistic lock
// 	err = tx.QueryRow(`
//         SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
//     `, transaction.AccountID).Scan(&balance)
// 	if err != nil {
// 		return fmt.Errorf("account not found: %w", err)
// 	}

// 	newBalance := balance + transaction.Amount
// 	_, err = tx.Exec(`
//         UPDATE accounts SET balance = $1 WHERE id = $2
//     `, newBalance, transaction.AccountID)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = tx.Exec(`
//         INSERT INTO transactions (id, account_id, amount) VALUES ($1, $2, $3)
//     `, transaction.TxID, transaction.AccountID, transaction.Amount)
// 	if err != nil {
// 		return err
// 	}

// 	return tx.Commit()
// }
