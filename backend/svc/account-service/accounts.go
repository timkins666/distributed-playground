package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/timkins666/distributed-playground/backend/pkg/appdb"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

var (
	banks []*cmn.Bank = []*cmn.Bank{{Name: "Bonzo", ID: 1}} // tmp
)

type app struct {
	cancelCtx    context.Context
	db           *appdb.DB
	payReqReader *kafka.Reader
	writer       *kafka.Writer
	log          *log.Logger
}

func main() {
	kafkaBroker := cmn.KafkaBroker()

	cancelCtx, stop := cmn.GetCancelContext()
	defer stop()

	db, err := appdb.InitDB(appdb.DefaultConfig)
	if err != nil {
		log.Panicln("Failed to initialise pstgres")
	}

	app := app{
		cancelCtx: cancelCtx,
		log:       cmn.AppLogger(),
		payReqReader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{kafkaBroker},
			Topic:   cmn.Topics.PaymentRequested(),
			GroupID: "payment-validator",
		}),
		writer: &kafka.Writer{
			Addr:         kafka.TCP(kafkaBroker),
			RequiredAcks: 1,
			MaxAttempts:  5,
		},
		db: db,
	}

	defer app.payReqReader.Close()
	defer app.writer.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/banks", getAllBanksHandler)
	mux.HandleFunc("/myaccounts", getUserAccountsHandler)
	mux.HandleFunc("/new", createUserAccountHandler)

	go paymentValidator(app)

	port := ":" + os.Getenv("SERVE_PORT")
	app.log.Printf("Accounts service running on %s", port)
	log.Fatal(http.ListenAndServe(port,
		cmn.SetUserIDMiddlewareHandler(
			cmn.SetContextValuesMiddleware(
				map[cmn.ContextKey]any{cmn.AppKey: app})(mux))))
}

func getAllBanksHandler(w http.ResponseWriter, r *http.Request) {
	app, _ := r.Context().Value(cmn.AppKey).(app)
	userID, ok := r.Context().Value(cmn.UserIDKey).(int)

	if userID == 0 || !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	user, err := app.db.LoadUserByID(userID)
	if err != nil || !user.Valid() {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
	}

	app.log.Printf("getallbanks user %s", user.Username)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(banks)
}

func getUserAccountsHandler(w http.ResponseWriter, r *http.Request) {
	app, _ := r.Context().Value(cmn.AppKey).(app)
	userID, _ := r.Context().Value(cmn.UserIDKey).(int)

	if userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w, "Nope")
		return
	}

	accs, err := app.db.GetUserAccounts(userID)

	if err == nil || err == sql.ErrNoRows {
		if len(accs) == 0 {
			accs = []cmn.Account{}
		}
		app.log.Printf("Accs: %+v", accs)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accs)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Println(w, "Nope")
}

type newAccountRequest struct {
	Name                 string `json:"name"`
	SourceFundsAccountID int    `json:"sourceFundsAccountId"`
	InitialBalance       int64  `json:"initialBalance"`
}

func createUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	app, _ := r.Context().Value(cmn.AppKey).(app)
	userID, _ := r.Context().Value(cmn.UserIDKey).(int)

	var req *newAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(w, "Nope")
		return
	}

	var sourceAcc *cmn.Account
	userAccounts, err := app.db.GetUserAccounts(userID)
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

		sourceAcc, err = app.db.GetAccountByID(req.SourceFundsAccountID)
		if sourceAcc == nil {
			app.log.Printf(
				"Account %d not found. Doesn't exist or not owned by user %s",
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

	if accID, err := app.db.CreateAccount(newAccount); err != nil || accID <= 0 {
		app.log.Println("ERROR: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(w, "Error creating account")
		return
	} else {
		newAccount.AccountID = accID
	}

	app.log.Println("Opened new account", newAccount)

	respAccounts := []cmn.Account{newAccount}

	txJson, _ := json.Marshal(cmn.Transaction{
		TxID:      uuid.NewString(),
		Amount:    newAccount.Balance,
		AccountID: newAccount.AccountID,
	})
	app.writer.WriteMessages(app.cancelCtx,
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested(),
			Value: txJson,
		})

	if sourceAcc != nil {
		txJson, _ := json.Marshal(cmn.Transaction{
			TxID:      uuid.NewString(),
			Amount:    -newAccount.Balance,
			AccountID: req.SourceFundsAccountID,
		})
		app.writer.WriteMessages(app.cancelCtx,
			kafka.Message{
				Topic: cmn.Topics.TransactionRequested(),
				Value: txJson,
			})
		sourceAcc.Balance -= req.InitialBalance
		respAccounts = append(respAccounts, *sourceAcc)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respAccounts)
}
