package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
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

func paymentValidator(cancelCtx context.Context) {
	// handles validation of requested payments

	// TODO: init kafkas in main
	broker := os.Getenv("KAFKA_BROKER")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   cmn.Topics.PaymentRequested(),
		GroupID: "payment-validator",
	})
	defer reader.Close()

	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		RequiredAcks: 1,
		MaxAttempts:  5,
	}

	max_errors := 10
	for {
		select {
		case <-cancelCtx.Done():
			log.Println("Context cancelled")
			return
		default:
			msg, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Println("READ MSG ERROR", err)
				max_errors--
				if max_errors == 0 {
					log.Println("Max errors reached, something seems wrong...")
					break
				}
				continue
			}
			go handlePaymentRequestedMessage(msg, writer, cancelCtx)
		}
	}
}

func handlePaymentRequestedMessage(message kafka.Message, writer *kafka.Writer, cancelCtx context.Context) {
	var req cmn.PaymentRequest
	err := json.Unmarshal(message.Value, &req)
	if err != nil {
		log.Println("ERROR failed to parse message:", err)
		return
	}
	log.Println("Read message: ", req)

	validateResult := &paymentValidationResult{
		paymentRequest: &req,
	}

	numChecks := 2 // being lazy
	results := make(chan checkResult, numChecks)
	go checkBalance(&req, results)
	go checkTargetAccount(&req, results)

	for {
		select {
		case res := <-results:
			// update central results here for single update point
			log.Println(res.checkName, ":", res.result)
			validateResult.results = append(validateResult.results, res)

			if len(validateResult.results) == numChecks {
				log.Println("all checks complete for request", req.SystemID)
				go handleResults(validateResult, writer, cancelCtx)
				return
			}
			log.Printf("waiting for %d remaining check for request %s", numChecks-len(validateResult.results), req.SystemID)
		case <-time.After(4500 * time.Millisecond): // 4.5s timeout to return so some will fail
			log.Println("Checks timed out for request", req.SystemID)
			validateResult.timedOut = true
			go handleResults(validateResult, writer, cancelCtx)
			return
		case <-cancelCtx.Done():
			return
		}
	}
}

func handleResults(result *paymentValidationResult, writer *kafka.Writer, cancelCtx context.Context) {
	if result.timedOut {
		sendPaymentFailed(result.paymentRequest, "timeout", writer, cancelCtx)
		return
	}

	errs := []string{}
	for _, check := range result.results {
		if !check.result {
			errs = append(errs, fmt.Sprintf("%s failed", check.checkName))
		}
	}

	if len(errs) > 0 {
		go sendPaymentFailed(result.paymentRequest, strings.Join(errs, ", "), writer, cancelCtx)
		return
	}

	// TODO: lock funds to prevent races before submitting transaction
	go initiateTransaction(result.paymentRequest, writer, cancelCtx)
}

func sendPaymentFailed(req *cmn.PaymentRequest, reason string, _ *kafka.Writer, _ context.Context) {
	// TODO: send payment failed message for gateway (or future notification service)
	log.Printf("Payment of £%d failed for account %d: %s", req.Amount, req.TargetAccountID, reason)
}

func initiateTransaction(req *cmn.PaymentRequest, writer *kafka.Writer, cancelCtx context.Context) {
	// send message(s) for transaction service

	// TODO: include both accounts for transfers within bank?

	log.Printf("Initiate transaction of £%d from account %d to account %d", req.Amount, req.SourceAccountID, req.TargetAccountID)

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
		sendPaymentFailed(req, "processing error", writer, cancelCtx)
	}

	err := writer.WriteMessages(cancelCtx,
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested(),
			Value: txOut,
		},
		kafka.Message{
			Topic: cmn.Topics.TransactionRequested(),
			Value: txIn,
		})

	if err != nil {
		sendPaymentFailed(req, "failed to initiate transaction", writer, cancelCtx)
	}

}

func checkBalance(req *cmn.PaymentRequest, chn chan<- checkResult) {
	// check source account has required funds
	checkName := "balance"

	// artificial delay
	sleep := rand.N(5000)
	log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	srcAcc := getAccountByID(req.SourceAccountID, nil)

	if srcAcc == nil {
		log.Printf("ERROR: Account id %d not found", req.SourceAccountID)
		result.result = false // TODO: unneccessary cos default but wait for tests
		chn <- result
		return
	}

	log.Printf("Account %d current balance £%d, requested payment of £%d", srcAcc.AccountID, srcAcc.Balance, req.Amount)
	result.result = srcAcc.Balance >= req.Amount
	chn <- result
}

func checkTargetAccount(req *cmn.PaymentRequest, chn chan<- checkResult) {
	// check target account is valid
	checkName := "targetAccount"

	// artificial delay
	sleep := rand.N(5000)
	log.Printf("%s sleeping for %d", checkName, sleep)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	result := checkResult{checkName: checkName}

	tgtAcc := getAccountByID(req.SourceAccountID, nil)

	if tgtAcc == nil {
		log.Printf("ERROR: Target account id %d not found", req.TargetAccountID)
		result.result = false
		chn <- result
		return
	}

	result.result = true
	chn <- result
}
