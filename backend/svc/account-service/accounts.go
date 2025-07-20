package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type Account struct {
	AccountId int    `json:"accountId"`
	Username  string `json:"username"`
	Balance   int64  `json:"balance"`
	BankId    int    `json:"bankId"`
	BankName  string `json:"bankName"`
}

type Bank struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

var (
	openAccounts []Account = []Account{}
	banks        []Bank    = []Bank{{Name: "Bonzo", Id: 1}}
)

func main() {
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		log.Fatalf("KAFKA_BROKER not found")
	}
	log.Println("kafka broker:", kafkaBroker)

	mux := http.NewServeMux()
	mux.HandleFunc("/banks", getAllBanksHandler)
	mux.HandleFunc("/myaccounts", getUserAccountsHandler)
	mux.HandleFunc("/new", createUserAccountHandler)

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	go paymentValidator(kafkaBroker, cancelCtx)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Accounts service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

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
		Balance:   rand.Int64N(10e5),
		BankId:    banks[0].Id,
		BankName:  banks[0].Name,
	}

	openAccounts = append(openAccounts, newAccount)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newAccount)
}
