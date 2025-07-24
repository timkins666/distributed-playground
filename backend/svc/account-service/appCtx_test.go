package main

import (
	"context"
	"log"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
	tu "github.com/timkins666/distributed-playground/backend/pkg/testutils"
)

type testApp struct {
	cmn.BaseEnv
	db           *MockDB
	writer       *tu.MockKafkaWriter
	payReqReader *tu.MockKafkaReader
	log          *log.Logger
	cancelCtx    context.Context
}

func (a *testApp) PayReqReader() cmn.KafkaReader {
	return a.payReqReader
}

func (a *testApp) Writer() cmn.KafkaWriter {
	return a.writer
}
