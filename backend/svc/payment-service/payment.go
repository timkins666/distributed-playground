package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type appEnv struct {
	cmn.BaseEnv
	writer cmn.KafkaWriter
}

func main() {
	writer := &kafka.Writer{
		Addr: kafka.TCP(cmn.KafkaBroker()),
	}
	defer writer.Close()

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()
	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(cancelCtx),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pay", handlePaymentRequest)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Payment service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetUserIDMiddlewareHandler(
			cmn.SetContextValuesMiddleware(
				map[cmn.ContextKey]any{cmn.EnvKey: env})(mux))))
}

// handles initial request from gateway
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	env, ok := r.Context().Value(cmn.EnvKey).(appEnv)
	if !ok {
		log.Println("invalid env")
		w.WriteHeader(http.StatusInternalServerError)
	}

	var req cmn.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	req.Timestamp = time.Now().UTC()
	req.SystemID = uuid.NewString()

	if !req.Valid() {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	msg, err := req.MsgValue()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key, err := req.MsgKey()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = env.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: cmn.Topics.PaymentRequested().S(),
		Key:   key,
		Value: msg,
	})
	if err != nil {
		env.Logger().Println("WRITE ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	env.Logger().Printf("Sent payment-requested message: %s", msg)
	w.WriteHeader(http.StatusAccepted)
}
