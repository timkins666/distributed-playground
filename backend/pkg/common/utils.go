package common

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func GetCancelContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func KafkaBroker() string {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		log.Fatalf("KAFKA_BROKER not found")
	}
	return kafkaBroker
}
