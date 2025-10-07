package common

import (
	"context"
	"encoding/json"
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

func FromBytes[T any](b []byte) (*T, error) {
	var t T
	err := json.Unmarshal(b, &t)
	return &t, err
}

func ToBytes[T any](t T) ([]byte, error) {
	return json.Marshal(t)
}
