package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type transactionCtx struct {
	cancelCtx   context.Context
	db          transactionDB
	logger      *log.Logger
	writer      cmn.KafkaWriter
	txReqReader cmn.KafkaReader
	redisClient *redis.Client
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

	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func newAppCtx(cancelCtx context.Context) transactionCtx {
	logger := cmn.AppLogger()

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
		logger.Fatal(err)
	}

	redisClient, err := cmn.NewRedisClient()
	if err != nil {
		logger.Println("Failed to start redis client, continuing without :(")
		redisClient = nil // make sure it is
	}

	return transactionCtx{
		cancelCtx:   cancelCtx,
		db:          db,
		writer:      writer,
		txReqReader: txReqReader,
		redisClient: redisClient,
		logger:      logger,
	}
}
