package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type accountsCtx struct {
	cancelCtx    context.Context
	db           accountsDB
	logger       *log.Logger
	payReqReader cmn.KafkaReader
	writer       cmn.KafkaWriter
	redisClient  *redis.Client
}

// Close releases all resources
func (a *accountsCtx) Close() error {
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

func newAppCtx(cancelCtx context.Context, config *Config) *accountsCtx {
	logger := cmn.AppLogger()

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

	redisClient, err := cmn.NewRedisClient() //TODO: env
	if err != nil {
		logger.Println("failed to initialise redis client, continuing without:", err.Error())
		redisClient = nil // make sure
	}

	db, err := initDB(redisClient)
	if err != nil {
		panic(err)
	}

	return &accountsCtx{
		cancelCtx:    cancelCtx,
		logger:       logger,
		payReqReader: reader,
		writer:       writer,
		db:           db,
		redisClient:  redisClient,
	}
}
