package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	cmn "github.com/timkins666/distributed-playground/backend/pkg/common"
)

type checkResult struct {
	checkName string
	result    bool
}

type paymentValidationResult struct {
	paymentRequest *cmn.PaymentRequest
	results        []checkResult
	timedOut       bool
}

func paymentValidator(app app) {
	// handles validation of requested payments

	max_errors := 10
	for {
		select {
		case <-app.cancelCtx.Done():
			app.log.Println("Context cancelled")
			return
		default:
			msg, err := app.payReqReader.ReadMessage(context.Background())
			if err != nil {
				app.log.Println("READ MSG ERROR", err)
				max_errors--
				if max_errors == 0 {
					app.log.Println("Max errors reached, something seems wrong...")
					break
				}
				continue
			}
			go handlePaymentRequestedMessage(msg, app)
		}
	}
}

func handlePaymentRequestedMessage(message kafka.Message, app app) {
	var req cmn.PaymentRequest
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		app.log.Println("ERROR failed to parse message:", err)
		return
	}
	app.log.Println("Read message: ", req)

	validateResult := &paymentValidationResult{
		paymentRequest: &req,
	}

	numChecks := 2 // being lazy
	results := make(chan checkResult, numChecks)
	go checkBalance(&req, results, app)
	go checkTargetAccount(&req, results, app)

	for {
		select {
		case res := <-results:
			// update central results here for single update point
			app.log.Println(res.checkName, ":", res.result)
			validateResult.results = append(validateResult.results, res)

			if len(validateResult.results) == numChecks {
				app.log.Println("all checks complete for request", req.SystemID)
				go handleResults(validateResult, app)
				return
			}
			app.log.Printf("waiting for %d remaining check for request %s", numChecks-len(validateResult.results), req.SystemID)
		case <-time.After(4500 * time.Millisecond): // 4.5s timeout to return so some will fail
			app.log.Println("Checks timed out for request", req.SystemID)
			validateResult.timedOut = true
			go handleResults(validateResult, app)
			return
		case <-app.cancelCtx.Done():
			return
		}
	}
}

func handleResults(result *paymentValidationResult, app app) {
	if result.timedOut {
		sendPaymentFailed(result.paymentRequest, "timeout", app)
		return
	}

	errs := []string{}
	for _, check := range result.results {
		if !check.result {
			errs = append(errs, fmt.Sprintf("%s failed", check.checkName))
		}
	}

	if len(errs) > 0 {
		go sendPaymentFailed(result.paymentRequest, strings.Join(errs, ", "), app)
		return
	}

	// TODO: lock funds to prevent races before submitting transaction
	go initiateTransaction(result.paymentRequest, app)
}

func sendPaymentFailed(req *cmn.PaymentRequest, reason string, app app) {
	// TODO: send payment failed message for gateway (or future notification service)
	app.log.Printf("Payment of £%d failed for account %d: %s", req.Amount, req.TargetAccountID, reason)
}

func initiateTransaction(req *cmn.PaymentRequest, app app) {
	// send message(s) for transaction service

	app.log.Printf("Initiate transaction of £%d from account %d to account %d", req.Amount, req.SourceAccountID, req.TargetAccountID)

	txOut, err1 := json.Marshal(cmn.Transaction{
		TxID:      req.SystemID,
		AccountID: req.SourceAccountID,
		Amount:    -req.Amount,
	})
	txIn, err2 := json.Marshal(cmn.Transaction{
		TxID:      req.SystemID,
		AccountID: req.TargetAccountID,
		Amount:    req.Amount,
	})

	if err1 != nil || err2 != nil {
		sendPaymentFailed(req, "processing error", app)
	}

	err := app.writer.WriteMessages(app.cancelCtx,
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested(),
			Value: txOut,
		},
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested(),
			Value: txIn,
		})

	if err != nil {
		sendPaymentFailed(req, "failed to initiate transaction", app)
	}

}

func checkBalance(req *cmn.PaymentRequest, chn chan<- checkResult, app app) {
	// check source account has required funds
	checkName := "balance"

	// artificial delay
	sleep := rand.N(5000)
	app.log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	srcAcc, err := app.db.GetAccountByID(req.SourceAccountID)

	if err != nil {
		app.log.Printf("ERROR: Account id %d not found. %s", req.SourceAccountID, err)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	app.log.Printf("Account %d current balance £%d, requested payment of £%d", srcAcc.AccountID, srcAcc.Balance, req.Amount)
	result.result = srcAcc.Balance >= req.Amount
	chn <- result
}

func checkTargetAccount(req *cmn.PaymentRequest, chn chan<- checkResult, app app) {
	// check target account exists

	checkName := "targetAccount"

	// artificial delay
	sleep := rand.N(5000)
	app.log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	_, err := app.db.GetAccountByID(req.SourceAccountID)

	if err != nil {
		app.log.Printf("ERROR: Target account id %d not found. %s", req.TargetAccountID, err)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	result.result = true
	chn <- result
}
