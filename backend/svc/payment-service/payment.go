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

func main() {
	writer := &kafka.Writer{
		Addr: kafka.TCP(cmn.KafkaBroker()),
	}
	defer writer.Close()

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	appCtx := newAppCtx(cancelCtx)

	mux := http.NewServeMux()
	mux.HandleFunc("/transfer", handlePaymentRequest)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Payment service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetUserIDMiddlewareHandler(
			cmn.SetContextValuesMiddleware(
				map[cmn.ContextKey]any{cmn.AppCtx: &appCtx})(mux))))
}

// handles initial transfer request from gateway
func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	appCtx, ok := r.Context().Value(cmn.AppCtx).(*paymentCtx)
	if !ok {
		log.Println("invalid appCtx")
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
	if err = createDBPayment(req, appCtx); err != nil {
		appCtx.logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = appCtx.writer.WriteMessages(context.Background(), kafka.Message{
		Topic: cmn.Topics.PaymentRequested().S(),
		Key:   key,
		Value: msg,
	})
	if err != nil {
		appCtx.logger.Println("WRITE ERR:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	appCtx.logger.Printf("Sent payment-requested message: %s", msg)
	w.WriteHeader(http.StatusAccepted)
}

func createDBPayment(req cmn.PaymentRequest, appCtx *paymentCtx) error {
	log.Printf("saving payment to db: %+v", req)
	return appCtx.db.createPayment(&req)
}

// TODO: on tx complete. + use types for status
// func (db *DB) UpdatePaymentStatus(sysId, status string) error {
// 	// TODO: check affected row count == 1
// 	_, err := db.db.Exec(`
// 		UPDATE payments.transfer
// 		SET status = $1
// 		WHERE system_id = $2
// 	`, status, sysId)

// 	return err
// }
