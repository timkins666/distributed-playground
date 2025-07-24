package main

import (
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type appEnv struct {
	cmn.BaseEnv
	payReqReader cmn.KafkaReader
	writer       cmn.KafkaWriter
}

func (a *appEnv) PayReqReader() cmn.KafkaReader {
	return a.payReqReader
}

func (a *appEnv) Writer() cmn.KafkaWriter {
	return a.writer
}
