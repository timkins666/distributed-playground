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

	db, err := cmn.InitDB(cmn.DefaultConfig)
	if err != nil {
		log.Panicln(err.Error())
	}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(cancelCtx).WithDB(db),
		writer:  writer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/transfer", handlePaymentRequest)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Payment service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetUserIDMiddlewareHandler(
			cmn.SetContextValuesMiddleware(
				map[cmn.ContextKey]any{cmn.EnvKey: &env})(mux))))
}

// handles initial transfer request from gateway
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	env, ok := r.Context().Value(cmn.EnvKey).(*appEnv)
	if !ok {
		log.Println("invalid env")
		w.WriteHeader(http.StatusInternalServerError)
		return
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

	msg, err := cmn.ToBytes(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	key, err := cmn.ToBytes(req.SourceAccountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create payment in system for tracking and analytics/reconciliation
	if err = createDBPayment(req, env); err != nil {
		env.Logger().Println(err)
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

func createDBPayment(req cmn.PaymentRequest, env *appEnv) error {
	log.Printf("saving payment to db: %+v", req)
	return env.DB().CreatePayment(&req)
}
