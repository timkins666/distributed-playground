package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentRequest struct {
	FromAccount string  `json:"from_account"`
	ToAccount   string  `json:"to_account"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
}

type KafkaPaymentEvent struct {
	PaymentID   string         `json:"payment_id"`
	Request     PaymentRequest `json:"request"`
	RequestedAt time.Time      `json:"requested_at"`
}

var kafkaWriter *kafka.Writer

var proxyHosts map[string]string = map[string]string{
	"auth":    os.Getenv("AUTH_SERVICE_HOST"),
	"account": os.Getenv("ACCOUNT_SERVICE_HOST"),
}

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		log.Panicf("KAFKA_BROKER not set")
	}

	kafkaWriter = &kafka.Writer{
		Addr:  kafka.TCP(kafkaBroker),
		Topic: "payment-requested",
	}
	defer kafkaWriter.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)

	for srv, _ := range proxyHosts {
		mux.HandleFunc("/"+srv+"/", withAuth(proxyToService))
	}

	port := ":" + os.Getenv("SERVE_PORT")
	log.Println("API Gateway running on %s", port)
	log.Fatal(http.ListenAndServe(port, corsMiddleware(mux)))
}
