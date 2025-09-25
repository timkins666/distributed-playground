package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type transactionCtx struct {
	cancelCtx   context.Context
	db          transactionDB
	logger      *log.Logger
	writer      cmn.KafkaWriter
	txReqReader cmn.KafkaReader
}

// close releases all resources
func (a *transactionCtx) close() error {
	var errs []error

	if a.txReqReader != nil {
		if err := a.txReqReader.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if a.writer != nil {
		if err := a.writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func newAppCtx(cancelCtx context.Context) transactionCtx {
	txReqReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cmn.KafkaBroker()},
		GroupID: "process-transaction",
		Topic:   cmn.Topics.TransactionRequested().S(),
	})

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cmn.KafkaBroker()),
		RequiredAcks: 1,
	}

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	return transactionCtx{
		cancelCtx:   cancelCtx,
		db:          db,
		logger:      cmn.AppLogger(),
		writer:      writer,
		txReqReader: txReqReader,
	}
}
