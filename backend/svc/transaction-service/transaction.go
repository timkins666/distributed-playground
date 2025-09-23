package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

func init() {
	gob.Register(cmn.Transaction{})
}

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
			processMessage(msg, &env)
		}
	}
}

var (
	errorParsingTransaction    = errors.New("error parsing transaction")
	errorInvalidTransaction    = errors.New("parsed transaction but bad data")
	errorCommittingTransaction = errors.New("error committing transaction, this is probably bad")
)

func processMessage(msg kafka.Message, env *appEnv) error {
	tx, err := cmn.FromBytes[cmn.Transaction](msg.Value)
	if err != nil {
		env.Logger().Println(err)
		env.Logger().Printf("received bytes:\n%s", msg.Value)
		return errorParsingTransaction
	}

	if !tx.Valid() {
		env.Logger().Printf("%s:%+v", errorInvalidTransaction, tx)
		return errorInvalidTransaction
	}

	tx.TxID = uuid.NewString()
	tx.KafkaID = fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)

	err = env.DB().CommitTransaction(tx)
	if err != nil {
		env.Logger().Println(err)
		return errorCommittingTransaction
	}

	// TODO: complete kafka message
	env.Logger().Printf("Completed transaction %+v", tx)
	return nil
}
