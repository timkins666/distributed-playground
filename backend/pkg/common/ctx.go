package common

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Writer interface {
	Writer() *kafka.Writer
}

type Env interface {
	CancelCtx() context.Context
	DB() *DB
	Logger() *log.Logger
}

type BaseEnv struct {
	cancelCtx context.Context
	db        DBAll
	logger    *log.Logger
}

func (b BaseEnv) WithCancelCtx(ctx context.Context) BaseEnv {
	b.cancelCtx = ctx
	return b
}

func (b BaseEnv) WithDB(db DBAll) BaseEnv {
	b.db = db
	return b
}

func (a *BaseEnv) CancelCtx() context.Context {
	if a.cancelCtx == nil {
		log.Panicf("cancel context accessed but not set")
	}
	return a.cancelCtx
}

func (a *BaseEnv) DB() DBAll {
	if a.db == nil {
		log.Panicf("db accessed but not set")
	}
	return a.db
}

func (a *BaseEnv) Logger() *log.Logger {
	if a.logger == nil {
		a.logger = AppLogger()
		a.logger.Println("Initialised default app logger")
	}
	return a.logger
}
