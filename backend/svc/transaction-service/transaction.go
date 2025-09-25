package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

var (
	errTxProcessed     = errors.New("transaction already processed")
	errAccountNotExist = errors.New("account doesn't exist")
)

func main() {
	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	appCtx := newAppCtx(cancelCtx)
	defer appCtx.close()

	for {
		select {
		case <-cancelCtx.Done():
			return
		default:
			msg, err := appCtx.txReqReader.ReadMessage(appCtx.cancelCtx)
			if err != nil {
				appCtx.logger.Println(err)
				continue
			}
			processMessage(msg, &appCtx)
		}
	}
}

var (
	errorParsingTransaction    = errors.New("error parsing transaction")
	errorInvalidTransaction    = errors.New("parsed transaction but bad data")
	errorCommittingTransaction = errors.New("error committing transaction, this is probably bad")
)

func processMessage(msg kafka.Message, appCtx *transactionCtx) error {
	tx, err := cmn.FromBytes[cmn.Transaction](msg.Value)
	if err != nil {
		appCtx.logger.Println(err)
		appCtx.logger.Printf("received bytes:\n%s", msg.Value)
		return errorParsingTransaction
	}

	if !tx.Valid() {
		appCtx.logger.Printf("%s:%+v", errorInvalidTransaction, tx)
		return errorInvalidTransaction
	}

	tx.TxID = uuid.NewString()
	tx.KafkaID = fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)

	err = appCtx.db.commitTransaction(tx)
	if err != nil {
		appCtx.logger.Println(err)
		return errorCommittingTransaction
	}

	// TODO: complete kafka message
	appCtx.logger.Printf("Completed transaction %+v", tx)
	return nil
}
