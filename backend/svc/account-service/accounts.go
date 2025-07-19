package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type Account struct {
	AccountId int     `json:"accountId"`
	Username  string  `json:"username"`
	Balance   float64 `json:"balance"`
	BankId    int     `json:"bankId"`
	BankName  string  `json:"bankName"`
}

type Bank struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

var (
	openAccounts []Account = []Account{}
	banks        []Bank    = []Bank{{Name: "Bonzo", Id: 1}}
)

func getAllBanksHandler(w http.ResponseWriter, r *http.Request) {
	user, err := cmn.GetUserFromClaims(r)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	log.Printf("getallbanks user %s", user.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(banks)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := cmn.GetUserFromClaims(r)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	// change to map when finished messing about
	userAccounts := []Account{}
	for _, acc := range openAccounts {
		if acc.Username == user.Username {
			userAccounts = append(userAccounts, acc)
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userAccounts)
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	user, err := cmn.GetUserFromClaims(r)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	newAccount := Account{
		AccountId: len(openAccounts) + 1,
		Username:  user.Username,
		Balance:   math.Round(rand.Float64()*10e6) / 100,
		BankId:    banks[0].Id,
		BankName:  banks[0].Name,
	}

	openAccounts = append(openAccounts, newAccount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newAccount)
}

func paymentValidator() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		log.Fatalf("KAFKA_BROKER not found")
	} else {
		log.Println("broker env", kafkaBroker)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaBroker},
		Topic:   "payment-requested",
		GroupID: "payment-validator",
	})

	max_errors := 10
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("READ MSG ERROR", err)
			max_errors -= 1
			if max_errors == 0 {
				break
			}
			continue
		}

		var req cmn.PaymentRequest
		err = json.Unmarshal(msg.Value, &req)
		if err != nil {
			log.Println("ERROR: failed to parse message")
			log.Println(err)
			// TODO: dead letter/retry queue
			continue
		}

		log.Println("Read message: ", req)
	}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/banks", getAllBanksHandler)
	mux.HandleFunc("/myaccounts", getUserAccountsHandler)
	mux.HandleFunc("/new", createUserAccountHandler)

	go paymentValidator()

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Accounts service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
