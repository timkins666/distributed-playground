package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func GetCancelContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}
