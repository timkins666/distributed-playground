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
	payReqWriter := &kafka.Writer{
		Addr:  kafka.TCP(cmn.KafkaBroker()),
		Topic: cmn.Topics.PaymentRequested(),
	}
	defer payReqWriter.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/pay", handlePaymentRequest(payReqWriter))

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Payment service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handlePaymentRequest(writer *kafka.Writer) func(w http.ResponseWriter, r *http.Request) {
	// handles initial request from gateway
	return func(w http.ResponseWriter, r *http.Request) {
		var req cmn.PaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SystemId != "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		req.Timestamp = time.Now().UTC()
		req.SystemId = uuid.NewString()

		if !req.Valid() {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		msg, err := json.Marshal(req)
		if err != nil {
			log.Println("WRITE ERR:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = writer.WriteMessages(context.Background(), kafka.Message{Value: msg})
		if err != nil {
			log.Println("WRITE ERR:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("Sent payment-requested message: %s", msg)
		w.WriteHeader(http.StatusAccepted)
	}
}
