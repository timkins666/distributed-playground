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

var (
	payReqWriter kafka.Writer
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	log.Println("broker env", kafkaBroker)

	// reader := kafka.NewReader(kafka.ReaderConfig{
	// 	Brokers: []string{kafkaBroker},
	// 	Topic:   "review_requests",
	// 	GroupID: "verification-service",

	payReqWriter = kafka.Writer{
		Addr:  kafka.TCP("kafka:9092"),
		Topic: "payment-requested",
	}
	defer payReqWriter.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/pay", handlePaymentRequest)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Payment service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
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

	msg, _ := json.Marshal(req)
	err := payReqWriter.WriteMessages(context.Background(), kafka.Message{Value: msg})
	if err != nil {
		log.Println("WRITE ERR:", err)
	} else {
		log.Printf("Sent payment-requested message: %s", msg)
	}
}
