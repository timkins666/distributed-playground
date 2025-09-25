package main

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type AccountsCtx struct {
	cancelCtx    context.Context
	db           accountsDB
	logger       *log.Logger
	payReqReader cmn.KafkaReader
	writer       cmn.KafkaWriter
}

// Close releases all resources
func (a *AccountsCtx) Close() error {
	var errs []error

	if a.payReqReader != nil {
		if err := a.payReqReader.Close(); err != nil {
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

func newAppCtx(cancelCtx context.Context, config *Config) *AccountsCtx {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{config.Kafka.Broker},
		Topic:   cmn.Topics.PaymentRequested().S(),
		GroupID: config.Kafka.GroupID,
	})

	writer := &kafka.Writer{
		Addr:         kafka.TCP(config.Kafka.Broker),
		RequiredAcks: config.Kafka.RequiredAcks,
		MaxAttempts:  config.Kafka.MaxAttempts,
	}

	db, err := initDB()
	if err != nil {
		panic(err)
	}

	return &AccountsCtx{
		cancelCtx:    cancelCtx,
		logger:       cmn.AppLogger(),
		payReqReader: reader,
		writer:       writer,
		db:           db,
	}
}
