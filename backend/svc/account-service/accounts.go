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
	AccountId int32  `json:"accountId"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Balance   int64  `json:"balance"`
	BankId    int32  `json:"bankId"`
	BankName  string `json:"bankName"`
}

type Bank struct {
	Name string `json:"name"`
	Id   int32  `json:"id"`
}

var (
	openAccounts []*Account = []*Account{}
	banks        []*Bank    = []*Bank{{Name: "Bonzo", Id: 1}}
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/banks", getAllBanksHandler)
	mux.HandleFunc("/myaccounts", getUserAccountsHandler)
	mux.HandleFunc("/new", createUserAccountHandler)

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	go paymentValidator(cancelCtx)

	port := ":" + os.Getenv("SERVE_PORT")
	log.Printf("Accounts service running on %s", port)
	log.Fatal(http.ListenAndServe(port, mux))
}

func getUserAccounts(username string) []*Account {
	// get accounts from the user from the lazy lazy slice
	userAccounts := []*Account{}
	for _, acc := range openAccounts {
		if acc.Username == username {
			userAccounts = append(userAccounts, acc)
		}
	}
	return userAccounts
}

func getAccountById(accountId int32, accounts []*Account) *Account {
	// get account matching id.
	// optionally pass a pre-filtered account list, or nil to search all.
	if accounts == nil {
		accounts = openAccounts
	}

	for _, acc := range accounts {
		if acc.AccountId == accountId {
			return acc
		}
	}

	return nil
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(getUserAccounts(user.Username))
}

type newAccountRequest struct {
	Name                 string `json:"name"`
	SourceFundsAccountId int32  `json:"sourceFundsAccountId"`
	InitialBalance       int64  `json:"initialBalance"`
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	user, err := cmn.GetUserFromClaims(r)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	var req *newAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(w, "Nope")
		return
	}

	var sourceAcc *Account
	userAccounts := getUserAccounts(user.Username)
	if len(userAccounts) > 0 {
		if req.InitialBalance <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"errorReason": "Must transfer with an initial balance"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		sourceAcc = getAccountById(req.SourceFundsAccountId, userAccounts)
		if sourceAcc == nil {
			log.Printf(
				"Account %d not found. Doesn't exist or not owned by user %s",
				req.SourceFundsAccountId,
				user.Username,
			)
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"errorReason": "Invalid source account"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		if sourceAcc.Balance < req.InitialBalance {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"errorReason": "Source account doesn't have enough funds"}
			json.NewEncoder(w).Encode(resp)
			return
		}
	} else {
		req.InitialBalance = rand.Int64N(10e5)
	}

	newAccount := &Account{
		AccountId: int32(len(openAccounts) + 1),
		Name:      req.Name,
		Username:  user.Username,
		Balance:   req.InitialBalance,
		BankId:    banks[0].Id,
		BankName:  banks[0].Name,
	}

	log.Println("Opened new account", newAccount)

	openAccounts = append(openAccounts, newAccount)
	respAccounts := []Account{*newAccount}

	// lazy. TODO: transaction
	if sourceAcc != nil {
		sourceAcc.Balance -= req.InitialBalance
		respAccounts = append(respAccounts, *sourceAcc)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respAccounts)
}
