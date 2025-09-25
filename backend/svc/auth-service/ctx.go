package main

import (
	"context"
	"log"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

// app context for the auth service
type authCtx struct {
	cancelCtx context.Context
	db        authDB
	logger    *log.Logger
}

func newAppCtx(cancelCtx context.Context) *authCtx {
	db, err := initDB()
	if err != nil {
		panic(err)
	}

	return &authCtx{
		cancelCtx: cancelCtx,
		logger:    cmn.AppLogger(),
		db:        db,
	}
}
