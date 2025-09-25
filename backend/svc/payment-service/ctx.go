package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type paymentCtx struct {
	cancelCtx context.Context
	db        transactionDB
	logger    *log.Logger
	writer    cmn.KafkaWriter
}

// Close releases all resources
func (a *paymentCtx) Close() error {
	var errs []error

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

func newAppCtx(cancelCtx context.Context) paymentCtx {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cmn.KafkaBroker()),
		RequiredAcks: 1,
	}
	defer writer.Close()

	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	return paymentCtx{
		cancelCtx: cancelCtx,
		db:        db,
		writer:    writer,
	}
}
