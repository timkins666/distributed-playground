package common

import (
	"context"
	"log"
	"net/http"
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

func SetContextValuesMiddleware(kv map[ContextKey]any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			for k, v := range kv {
				ctx = context.WithValue(ctx, k, v)
			}
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
