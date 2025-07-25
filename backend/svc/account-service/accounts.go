package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

var (
	banks []*cmn.Bank = []*cmn.Bank{{Name: "Bonzo", ID: 1}} // tmp
)

func main() {
	kafkaBroker := cmn.KafkaBroker()

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	db, err := cmn.InitDB(cmn.DefaultConfig)
	if err != nil {
		log.Panicln("Failed to initialise postgres")
	}

	env := appEnv{
		BaseEnv: cmn.BaseEnv{}.WithCancelCtx(cancelCtx).WithDB(db),
		payReqReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{kafkaBroker},
			Topic:   cmn.Topics.PaymentRequested().S(),
			GroupID: "payment-validator",
		}),
		writer: &kafka.Writer{
			Addr:         kafka.TCP(kafkaBroker),
			RequiredAcks: 1,
			MaxAttempts:  5,
		},
	}

	defer env.payReqReader.Close()
	defer env.writer.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/banks", getAllBanksHandler)
	mux.HandleFunc("/myaccounts", getUserAccountsHandler)
	mux.HandleFunc("/new", createUserAccountHandler)

	go paymentValidator(env)

	port := ":" + os.Getenv("SERVE_PORT")
	env.Logger().Printf("Accounts service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetUserIDMiddlewareHandler(
			cmn.SetContextValuesMiddleware(
				map[cmn.ContextKey]any{cmn.AppKey: env})(mux))))
}

func getAllBanksHandler(w http.ResponseWriter, r *http.Request) {
	env, _ := r.Context().Value(cmn.AppKey).(appEnv)
	userID, ok := r.Context().Value(cmn.UserIDKey).(int32)

	if userID == 0 || !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	user, err := env.DB().LoadUserByID(userID)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
	}

	env.Logger().Printf("getallbanks user %s", user.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(banks)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	env, _ := r.Context().Value(cmn.AppKey).(appEnv)
	userID, _ := r.Context().Value(cmn.UserIDKey).(int32)

	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	accs, err := env.DB().GetUserAccounts(userID)

	if err == nil || err == sql.ErrNoRows {
		if len(accs) == 0 {
			accs = []cmn.Account{}
		}
		env.Logger().Printf("Accs: %+v", accs)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accs)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Println(w, "Nope")
}

type newAccountRequest struct {
	Name                 string `json:"name"`
	SourceFundsAccountID int32  `json:"sourceFundsAccountId"`
	InitialBalance       int64  `json:"initialBalance"`
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	env, _ := r.Context().Value(cmn.AppKey).(appEnv)
	userID, _ := r.Context().Value(cmn.UserIDKey).(int32)

	var req *newAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(w, "Nope")
		return
	}

	var sourceAcc *cmn.Account
	userAccounts, err := env.DB().GetUserAccounts(userID)
	if err != nil {
		// TODO: handle err
	}

	if len(userAccounts) > 0 {
		if req.InitialBalance <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			resp := map[string]string{"errorReason": "Must transfer with an initial balance"}
			json.NewEncoder(w).Encode(resp)
			return
		}

		sourceAcc, err = env.DB().GetAccountByID(req.SourceFundsAccountID)
		if sourceAcc == nil {
			env.Logger().Printf(
				"Account %d not found. Doesn't exist or not owned by user %d",
				req.SourceFundsAccountID,
				userID,
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

	newAccount := cmn.Account{
		Name:     req.Name,
		UserID:   userID,
		Balance:  req.InitialBalance,
		BankID:   banks[0].ID,
		BankName: banks[0].Name,
	}

	if accID, err := env.DB().CreateAccount(newAccount); err != nil || accID <= 0 {
		env.Logger().Println("ERROR: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(w, "Error creating account")
		return
	} else {
		newAccount.AccountID = accID
	}

	env.Logger().Println("Opened new account", newAccount)

	respAccounts := []cmn.Account{newAccount}

	txJson, _ := json.Marshal(cmn.Transaction{
		TxID:      uuid.NewString(),
		Amount:    newAccount.Balance,
		AccountID: newAccount.AccountID,
	})
	env.Writer().WriteMessages(env.CancelCtx(),
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested().S(),
			Value: txJson,
		})

	if sourceAcc != nil {
		txJson, _ := json.Marshal(cmn.Transaction{
			TxID:      uuid.NewString(),
			Amount:    -newAccount.Balance,
			AccountID: req.SourceFundsAccountID,
		})
		env.Writer().WriteMessages(env.CancelCtx(),
			kafka.Message{
				Topic: cmn.Topics.TransactionRequested().S(),
				Value: txJson,
			})
		sourceAcc.Balance -= req.InitialBalance
		respAccounts = append(respAccounts, *sourceAcc)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respAccounts)
}
